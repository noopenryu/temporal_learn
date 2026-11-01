package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/service"
	"ms-customer-onboard/internal/temporal/workflows"
)

type Handler struct {
	svc  service.OnboardingService
	repo domain.OnboardingRepository
}

func NewHandler(svc service.OnboardingService, repo domain.OnboardingRepository) *Handler {
	return &Handler{
		svc:  svc,
		repo: repo,
	}
}

// StartSessionHandler initiates a new onboarding session or resumes an active one
func (h *Handler) StartSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req StartSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	app, isResumed, err := h.svc.StartSession(r.Context(), req.CustomerID, req.Email, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyOnboarded) || strings.Contains(err.Error(), "already onboarded") {
			writeError(w, http.StatusConflict, "A wallet account is already registered for this phone number. Please log in to your wallet.")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := http.StatusCreated
	if isResumed {
		status = http.StatusOK
	}

	writeJSON(w, status, BuildDetailResponseWithResume(app, isResumed))
}

// GetApplicationHandler retrieves the current application state, step checklist, and progress
func (h *Handler) GetApplicationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	app, err := h.svc.GetApplication(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "Onboarding application not found")
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// GetActiveByCustomerIDHandler finds active application by customer identifier
func (h *Handler) GetActiveByCustomerIDHandler(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if strings.TrimSpace(customerID) == "" {
		writeError(w, http.StatusBadRequest, "customer_id is required")
		return
	}

	app, err := h.svc.GetActiveByCustomerID(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "No active onboarding application found for customer")
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}

// GetActiveByPhoneNumberHandler finds active application by phone number
func (h *Handler) GetActiveByPhoneNumberHandler(w http.ResponseWriter, r *http.Request) {
	rawPhone := chi.URLParam(r, "phoneNumber")
	var phone string
	if strings.TrimSpace(rawPhone) != "" {
		unescaped, err := url.PathUnescape(rawPhone)
		if err == nil && unescaped != "" {
			phone = unescaped
		} else {
			phone = rawPhone
		}
	} else {
		phone = r.URL.Query().Get("phone")
	}

	phone = strings.TrimSpace(phone)
	if phone == "" {
		writeError(w, http.StatusBadRequest, "phone_number is required")
		return
	}

	app, err := h.svc.GetActiveByPhoneNumber(r.Context(), phone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "No active onboarding application found for this phone number")
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponseWithResume(app, true))
}

// ResumeSessionHandler verifies session viability and logs resume activity
func (h *Handler) ResumeSessionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	app, err := h.svc.ResumeSession(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "cancelled") {
			writeError(w, http.StatusGone, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchPersonalInfoHandler updates personal information
func (h *Handler) PatchPersonalInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchPersonalInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchPersonalInfo(r.Context(), id, req.ToDomain())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchAddressHandler updates address information
func (h *Handler) PatchAddressHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchAddress(r.Context(), id, req.ToDomain())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchIdentityDocsHandler updates identity documents
func (h *Handler) PatchIdentityDocsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchIdentityDocsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchIdentityDocs(r.Context(), id, req.ToDomain())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchFinancialInfoHandler updates financial details
func (h *Handler) PatchFinancialInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchFinancialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchFinancialInfo(r.Context(), id, req.ToDomain())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchWalletTierHandler updates requested wallet tier
func (h *Handler) PatchWalletTierHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchWalletTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchWalletTier(r.Context(), id, req.WalletTier)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// PatchDraftHandler updates multiple sections simultaneously
func (h *Handler) PatchDraftHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req service.PatchDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PatchDraft(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}


// SubmitApplicationHandler validates completeness and triggers Temporal workflow
func (h *Handler) SubmitApplicationHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	app, err := h.svc.SubmitApplication(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "cannot submit") {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, BuildDetailResponse(app))
}

// CancelSessionHandler cancels an in-progress draft
func (h *Handler) CancelSessionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req CancelSessionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Reason == "" {
		req.Reason = "User requested cancellation"
	}

	if err := h.svc.CancelSession(r.Context(), id, req.Reason); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Onboarding session successfully cancelled",
	})
}

// ResubmitDocumentsHandler signals workflow with updated document files
func (h *Handler) ResubmitDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PatchIdentityDocsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	if err := h.svc.ResubmitDocuments(r.Context(), id, req.ToDomain()); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Documents resubmitted successfully to ongoing review workflow",
	})
}

// ManualReviewDecisionHandler allows compliance officers to submit review outcome
func (h *Handler) ManualReviewDecisionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req ManualReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	decision := workflows.ManualReviewDecision{
		ReviewerID:    req.ReviewerID,
		Approved:      req.Approved,
		Notes:         req.Notes,
		ConfirmedTier: req.ConfirmedTier,
		ConfirmedData: req.ConfirmedData,
	}

	if err := h.svc.SubmitManualReviewDecision(r.Context(), id, decision); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Manual review decision submitted successfully to workflow",
	})
}

// PutConfirmedDataHandler updates confirmed verification data from staff review
func (h *Handler) PutConfirmedDataHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	var req PutConfirmedDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload: "+err.Error())
		return
	}

	app, err := h.svc.PutConfirmedData(r.Context(), id, req.ToDomain())
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BuildDetailResponse(app))
}

// ListApplicationsHandler returns a paginated list of onboarding applications for staff
func (h *Handler) ListApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	apps, total, err := h.svc.ListApplications(r.Context(), domain.ListFilter{
		Status: status,
		Search: search,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list applications: "+err.Error())
		return
	}

	items := make([]ApplicationDetailResponse, len(apps))
	for i, app := range apps {
		items[i] = BuildDetailResponse(app)
	}

	writeJSON(w, http.StatusOK, ListApplicationsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// GetAuditLogsHandler returns audit events for an application
func (h *Handler) GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application ID format")
		return
	}

	logs, err := h.svc.GetAuditLogs(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to retrieve audit logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"application_id": id,
		"logs":           logs,
		"count":          len(logs),
	})
}

// HealthCheckHandler verifies server health and database status
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"service": "ms-customer-onboard",
		"status":  "healthy",
	}

	if h.repo != nil {
		if err := h.repo.Ping(r.Context()); err != nil {
			status["status"] = "degraded"
			status["database"] = "unhealthy: " + err.Error()
			writeJSON(w, http.StatusServiceUnavailable, status)
			return
		}
		status["database"] = "connected"
	}

	writeJSON(w, http.StatusOK, status)
}

// GetWorkflowProgressHandler returns real-time Temporal workflow progress
func (h *Handler) GetWorkflowProgressHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid application UUID: "+err.Error())
		return
	}

	progress, err := h.svc.GetWorkflowProgress(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, progress)
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	return uuid.Parse(param)
}

func handleServiceError(w http.ResponseWriter, err error) {
	msg := err.Error()
	if strings.Contains(msg, "not found") {
		writeError(w, http.StatusNotFound, msg)
		return
	}
	if strings.Contains(msg, "cannot be modified") || strings.Contains(msg, "invalid") {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	writeError(w, http.StatusInternalServerError, msg)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
