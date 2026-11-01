package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ms-customer-onboard/internal/config"
	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/service"
	transporthttp "ms-customer-onboard/internal/transport/http"
)

type mockRepository struct {
	apps      map[uuid.UUID]*domain.OnboardingApplication
	auditLogs []domain.AuditLog
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		apps:      make(map[uuid.UUID]*domain.OnboardingApplication),
		auditLogs: make([]domain.AuditLog, 0),
	}
}

func (m *mockRepository) Create(_ context.Context, app *domain.OnboardingApplication) error {
	m.apps[app.ID] = app
	return nil
}

func (m *mockRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, ok := m.apps[id]
	if !ok {
		return nil, nil
	}
	clone := *app
	return &clone, nil
}

func (m *mockRepository) GetActiveByCustomerID(_ context.Context, customerID string) (*domain.OnboardingApplication, error) {
	for _, app := range m.apps {
		if app.CustomerID == customerID && app.Status != domain.StatusCancelled && app.Status != domain.StatusExpired {
			clone := *app
			return &clone, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetActiveByPhoneNumber(_ context.Context, phone string) (*domain.OnboardingApplication, error) {
	norm := domain.NormalizePhoneNumber(phone)
	for _, app := range m.apps {
		if domain.NormalizePhoneNumber(app.PersonalInfo.PhoneNumber) == norm && app.Status != domain.StatusCancelled && app.Status != domain.StatusExpired {
			clone := *app
			return &clone, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) List(_ context.Context, filter domain.ListFilter) ([]*domain.OnboardingApplication, int, error) {
	var list []*domain.OnboardingApplication
	for _, app := range m.apps {
		if filter.Status != "" && filter.Status != "ALL" && string(app.Status) != filter.Status {
			continue
		}
		clone := *app
		list = append(list, &clone)
	}
	return list, len(list), nil
}

func (m *mockRepository) Update(_ context.Context, app *domain.OnboardingApplication) error {
	m.apps[app.ID] = app
	return nil
}

func (m *mockRepository) SaveAuditLog(_ context.Context, log *domain.AuditLog) error {
	m.auditLogs = append(m.auditLogs, *log)
	return nil
}

func (m *mockRepository) GetAuditLogs(_ context.Context, _ uuid.UUID) ([]domain.AuditLog, error) {
	return m.auditLogs, nil
}

func (m *mockRepository) Ping(_ context.Context) error {
	return nil
}

func setupTestServer() (http.Handler, *mockRepository) {
	repo := newMockRepository()
	cfg := &config.Config{DraftTTLDays: 14}
	svc := service.NewOnboardingService(cfg, repo, nil)
	h := transporthttp.NewHandler(svc, repo)
	router := transporthttp.SetupRouter(h)
	return router, repo
}

func TestHTTPHandlers(t *testing.T) {
	router, _ := setupTestServer()

	var appID string

	t.Run("Start onboarding session", func(t *testing.T) {
		body := map[string]string{
			"customer_id":  "cust-http-1",
			"email":        "http-user@example.com",
			"phone_number": "+84912345678",
		}
		jsonBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/start", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp transporthttp.ApplicationDetailResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Application.ID)
		assert.Equal(t, domain.StatusDraft, resp.Application.Status)
		appID = resp.Application.ID.String()
	})

	t.Run("Start session without customer_id auto-generates identifier", func(t *testing.T) {
		body := map[string]string{
			"email":        "phoneonly@example.com",
			"phone_number": "+15551112222",
		}
		jsonBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/start", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp transporthttp.ApplicationDetailResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Application.CustomerID)
		assert.Contains(t, resp.Application.CustomerID, "CUST-")
	})

	t.Run("Get active by phone number returns active application", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/phone/%2B15551112222/active", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp transporthttp.ApplicationDetailResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "+15551112222", resp.Application.PersonalInfo.PhoneNumber)
		assert.True(t, resp.IsResumed)
	})

	t.Run("Get active by query param ?phone=", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/active?phone=%2B15551112222", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Get active by unknown phone returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/phone/%2B19999999999/active", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("Start session for already approved customer returns 409 Conflict", func(t *testing.T) {
		r2, repo2 := setupTestServer()
		approvedApp := &domain.OnboardingApplication{
			ID:           uuid.New(),
			CustomerID:   "CUST-APPROVED",
			Status:       domain.StatusApproved,
			PersonalInfo: domain.PersonalInfo{PhoneNumber: "+15558889999", Email: "approved@example.com"},
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		_ = repo2.Create(context.Background(), approvedApp)

		body := map[string]string{
			"email":        "approved@example.com",
			"phone_number": "+15558889999",
		}
		jsonBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/start", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		r2.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("Get application by ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/"+appID, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp transporthttp.ApplicationDetailResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, appID, resp.Application.ID.String())
	})

	t.Run("Patch personal info", func(t *testing.T) {
		payload := transporthttp.PatchPersonalInfoRequest{
			FirstName:   "Linh",
			LastName:    "Tran",
			DateOfBirth: time.Now().AddDate(-26, 0, 0),
			Nationality: "VN",
			PhoneNumber: "+84912345678",
			Email:       "linh@example.com",
		}
		jsonBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/onboarding/"+appID+"/personal-info", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp transporthttp.ApplicationDetailResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, "Linh", resp.Application.PersonalInfo.FirstName)
	})

	t.Run("Patch address info", func(t *testing.T) {
		payload := transporthttp.PatchAddressRequest{
			Country:       "VN",
			City:          "Hanoi",
			StreetAddress: "12 Trang Tien",
			PostalCode:    "100000",
		}
		jsonBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/onboarding/"+appID+"/address", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Submit incomplete application returns 422 Unprocessable Entity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/onboarding/"+appID+"/submit", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("Health check returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("List applications endpoint for staff", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/applications?status=DRAFT", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var listResp transporthttp.ListApplicationsResponse
		err := json.NewDecoder(rec.Body).Decode(&listResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, listResp.Items)
	})

	t.Run("Put confirmed data endpoint", func(t *testing.T) {
		confirmPayload := transporthttp.PutConfirmedDataRequest{
			ReviewerID:     "STAFF-REV-99",
			FirstName:      "Linh",
			LastName:       "Tran",
			DocumentNumber: "VN-DOC-12345",
			ApprovedTier:   "TIER_2_PREMIUM",
			Notes:          "Verified in staff portal",
		}
		jsonBytes, _ := json.Marshal(confirmPayload)
		req := httptest.NewRequest(http.MethodPut, "/api/v1/onboarding/"+appID+"/confirm-data", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var detailResp transporthttp.ApplicationDetailResponse
		_ = json.NewDecoder(rec.Body).Decode(&detailResp)
		assert.NotNil(t, detailResp.Application.ConfirmedData)
		assert.Equal(t, "VN-DOC-12345", detailResp.Application.ConfirmedData.DocumentNumber)
	})
}
