package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ms-cust-onboard-client/internal/domain"
	"ms-cust-onboard-client/internal/downstream"
	"ms-cust-onboard-client/internal/service"
)

type Handler struct {
	svc service.ClientService
}

func NewHandler(svc service.ClientService) *Handler {
	return &Handler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func writeFieldErrors(w http.ResponseWriter, status int, message string, fieldErrors map[string]string) {
	writeJSON(w, status, ErrorResponse{Error: message, FieldErrors: fieldErrors})
}

func parseUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	return uuid.Parse(strings.TrimSpace(param))
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ok, _ := h.svc.HealthCheck(r.Context())
	status := http.StatusOK
	if !ok {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]interface{}{
		"status":   "ok",
		"service":  "ms-cust-onboard-client",
		"downstream": ok,
	})
}

// StartSessionHandler validates input, initializes onboarding session and stores password hash
func (h *Handler) StartSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.StartSession(r.Context(), service.StartSessionInput{
		CustomerID:  req.CustomerID,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "An application already exists for this phone number. Please enter your application password to resume.")
			return
		}
		if strings.Contains(err.Error(), "already registered") || strings.Contains(err.Error(), "already onboarded") {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := http.StatusCreated
	if detail.IsResumed {
		status = http.StatusOK
	}
	writeJSON(w, status, detail)
}

// ResumeSessionHandler verifies phone + password credentials and restores active session
func (h *Handler) ResumeSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req ResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.ResumeSession(r.Context(), service.ResumeSessionInput{
		PhoneNumber: req.PhoneNumber,
		Password:    req.Password,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "Invalid phone number or password. Please verify and try again.")
			return
		}
		if errors.Is(err, domain.ErrCredentialNotFound) {
			writeError(w, http.StatusNotFound, "No active application found for this phone number, or invalid credentials.")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

// GetApplicationHandler fetches application details
func (h *Handler) GetApplicationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	detail, err := h.svc.GetApplication(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Application not found: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// CancelSessionHandler cancels an in-progress draft
func (h *Handler) CancelSessionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req CancelRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	reason := req.Reason
	if reason == "" {
		reason = "Customer cancelled onboarding"
	}

	if err := h.svc.CancelSession(r.Context(), id, reason); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to cancel session: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Application cancelled successfully"})
}

// SubmitApplicationHandler submits application to initiate Temporal onboarding workflow
func (h *Handler) SubmitApplicationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	detail, err := h.svc.SubmitApplication(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// PatchPersonalInfoHandler validates, enriches and forwards personal details
func (h *Handler) PatchPersonalInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var input service.PatchPersonalInfoInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchPersonalInfo(r.Context(), id, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// PatchAddressHandler validates and forwards residential address
func (h *Handler) PatchAddressHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req downstream.PatchAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchAddress(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// PatchIdentityDocsHandler validates and forwards identity documents
func (h *Handler) PatchIdentityDocsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req downstream.PatchIdentityDocsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchIdentityDocs(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// PatchFinancialInfoHandler validates and forwards financial details
func (h *Handler) PatchFinancialInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req downstream.PatchFinancialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchFinancialInfo(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// PatchWalletTierHandler validates and forwards wallet tier choice
func (h *Handler) PatchWalletTierHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req downstream.PatchWalletTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchWalletTier(r.Context(), id, req.WalletTier)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// ResubmitDocumentsHandler forwards resubmitted documents
func (h *Handler) ResubmitDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req downstream.ResubmitDocsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if err := h.svc.ResubmitDocuments(r.Context(), id, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Documents resubmitted successfully"})
}

// GetAuditLogsHandler fetches audit trail
func (h *Handler) GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	logs, err := h.svc.GetAuditLogs(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve audit logs: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// GetWorkflowProgressHandler retrieves live Temporal progress from downstream
func (h *Handler) GetWorkflowProgressHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	progress, err := h.svc.GetWorkflowProgress(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve workflow progress: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

// PatchDraftHandler updates draft fields in a single call
func (h *Handler) PatchDraftHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	detail, err := h.svc.PatchDraft(r.Context(), id, raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// UploadDocumentHandler simulates secure document upload for e-KYC scanner
func (h *Handler) UploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DocType  string `json:"doc_type"`
		FileName string `json:"file_name"`
		DataURL  string `json:"data_url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	targetURL := req.DataURL
	if targetURL == "" {
		targetURL = fmt.Sprintf("https://storage.payflow.internal/kyc/%s/%d.jpg", req.DocType, time.Now().UnixNano())
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"url":           targetURL,
		"quality_score": 98,
		"status":        "VERIFIED_LEGIBLE",
		"created_at":    time.Now(),
	})
}
