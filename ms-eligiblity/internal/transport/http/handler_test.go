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

	"ms-eligiblity/internal/config"
	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/service"
	transporthttp "ms-eligiblity/internal/transport/http"
)

type mockRepo struct{}

func (m *mockRepo) SaveCheck(_ context.Context, _ *domain.EligibilityCheckResult, _ *domain.EligibilityCheckRequest) error {
	return nil
}
func (m *mockRepo) GetCheckByID(_ context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
	return &domain.EligibilityCheckResult{
		ID:         id,
		CustomerID: "cust-1",
		Status:     domain.StatusApproved,
	}, nil
}
func (m *mockRepo) GetLatestCheckByCustomerID(_ context.Context, customerID string) (*domain.EligibilityCheckResult, error) {
	return &domain.EligibilityCheckResult{
		ID:         uuid.New(),
		CustomerID: customerID,
		Status:     domain.StatusApproved,
	}, nil
}
func (m *mockRepo) IsNationalIDRegistered(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}
func (m *mockRepo) IsSanctionedOrBlacklisted(_ context.Context, _, _, _ string) (bool, string, error) {
	return false, "", nil
}
func (m *mockRepo) RegisterExistingIdentity(_ context.Context, _, _, _, _, _ string) error {
	return nil
}
func (m *mockRepo) Ping(_ context.Context) error {
	return nil
}

func setupTestServer() http.Handler {
	cfg := &config.Config{MinEligibilityAge: 18}
	repo := &mockRepo{}
	svc := service.NewEligibilityService(cfg, repo)
	handler := transporthttp.NewHandler(svc, repo)
	return transporthttp.SetupRouter(handler)
}

func TestCheckEligibilityHandler(t *testing.T) {
	router := setupTestServer()

	t.Run("Valid request returns 200 and result", func(t *testing.T) {
		reqBody := domain.EligibilityCheckRequest{
			CustomerID:          "cust-http-1",
			NationalID:          "VN123456789",
			IDType:              "NATIONAL_ID",
			FirstName:           "Van",
			LastName:            "A",
			DateOfBirth:         time.Now().AddDate(-25, 0, 0),
			Nationality:         "VN",
			ResidenceCountry:    "VN",
			Email:               "van@example.com",
			PhoneNumber:         "+84901234567",
			WalletTierRequested: "TIER_1_BASIC",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/eligibility/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp domain.EligibilityCheckResult
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusApproved, resp.Status)
		assert.Equal(t, "cust-http-1", resp.CustomerID)
	})

	t.Run("Invalid request returns 400 Bad Request", func(t *testing.T) {
		// Missing required customer_id and national_id
		reqBody := domain.EligibilityCheckRequest{
			FirstName: "Incomplete",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/eligibility/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "customer_id is required")
	})
}

func TestListRulesHandler(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/eligibility/rules", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "AgeVerificationRule")
}

func TestHealthCheckHandler(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "healthy")
}
