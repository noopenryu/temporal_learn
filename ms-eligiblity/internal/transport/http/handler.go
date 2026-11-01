package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/service"
)

type Handler struct {
	svc  service.EligibilityService
	repo domain.EligibilityRepository
}

func NewHandler(svc service.EligibilityService, repo domain.EligibilityRepository) *Handler {
	return &Handler{
		svc:  svc,
		repo: repo,
	}
}

// CheckEligibilityHandler handles synchronous eligibility evaluation
func (h *Handler) CheckEligibilityHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.EligibilityCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	result, err := h.svc.Evaluate(r.Context(), &req)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid eligibility request") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		// If evaluation succeeded but persistence failed, result is still returned
		if result != nil {
			writeJSON(w, http.StatusOK, result)
			return
		}
		writeError(w, http.StatusInternalServerError, "Eligibility evaluation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetCheckByIDHandler retrieves evaluation details by check ID
func (h *Handler) GetCheckByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	checkID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid check ID format")
		return
	}

	result, err := h.svc.GetCheckByID(r.Context(), checkID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "Eligibility check not found")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetLatestByCustomerIDHandler retrieves latest evaluation for a customer
func (h *Handler) GetLatestByCustomerIDHandler(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerId")
	if strings.TrimSpace(customerID) == "" {
		writeError(w, http.StatusBadRequest, "Customer ID is required")
		return
	}

	result, err := h.svc.GetLatestByCustomerID(r.Context(), customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "No eligibility evaluation found for this customer")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ListRulesHandler returns all active rules
func (h *Handler) ListRulesHandler(w http.ResponseWriter, _ *http.Request) {
	rules := h.svc.ListRules()
	writeJSON(w, http.StatusOK, map[string]any{
		"rules": rules,
		"count": len(rules),
	})
}

// HealthCheckHandler verifies server health and database connection
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"service": "ms-eligiblity",
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
