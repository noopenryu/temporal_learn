package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ms-cust-onboard-client/internal/domain"
	"ms-cust-onboard-client/internal/downstream"
	"ms-cust-onboard-client/internal/service"
	transporthttp "ms-cust-onboard-client/internal/transport/http"
)

type mockClientService struct {
	mock.Mock
}

func (m *mockClientService) StartSession(ctx context.Context, input service.StartSessionInput) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) ResumeSession(ctx context.Context, input service.ResumeSessionInput) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) GetApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) CancelSession(ctx context.Context, id uuid.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockClientService) SubmitApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) PatchPersonalInfo(ctx context.Context, id uuid.UUID, input service.PatchPersonalInfoInput) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) PatchAddress(ctx context.Context, id uuid.UUID, req downstream.PatchAddressRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) PatchIdentityDocs(ctx context.Context, id uuid.UUID, req downstream.PatchIdentityDocsRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) PatchFinancialInfo(ctx context.Context, id uuid.UUID, req downstream.PatchFinancialRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, tier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) ResubmitDocuments(ctx context.Context, id uuid.UUID, req downstream.ResubmitDocsRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *mockClientService) GetAuditLogs(ctx context.Context, id uuid.UUID) ([]downstream.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]downstream.AuditLog), args.Error(1)
}

func (m *mockClientService) GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*downstream.WorkflowProgressResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.WorkflowProgressResponse), args.Error(1)
}

func (m *mockClientService) PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockClientService) HealthCheck(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func TestStartSessionHandler_Success(t *testing.T) {
	mockSvc := new(mockClientService)
	handler := transporthttp.NewHandler(mockSvc)
	router := transporthttp.SetupRouter(handler)

	appID := uuid.New()
	detail := &downstream.DetailResponse{
		Application: downstream.Application{
			ID:          appID,
			Status:      "DRAFT",
			CurrentStep: "personal_info",
		},
	}

	mockSvc.On("StartSession", mock.Anything, service.StartSessionInput{
		Email:       "alice@example.com",
		PhoneNumber: "+15551234567",
		Password:    "secret123",
	}).Return(detail, nil)

	payload := map[string]string{
		"email":        "alice@example.com",
		"phone_number": "+15551234567",
		"password":     "secret123",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/client/onboarding/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestResumeSessionHandler_InvalidCredentials(t *testing.T) {
	mockSvc := new(mockClientService)
	handler := transporthttp.NewHandler(mockSvc)
	router := transporthttp.SetupRouter(handler)

	mockSvc.On("ResumeSession", mock.Anything, service.ResumeSessionInput{
		PhoneNumber: "+15551234567",
		Password:    "wrongpass",
	}).Return(nil, domain.ErrInvalidCredentials)

	payload := map[string]string{
		"phone_number": "+15551234567",
		"password":     "wrongpass",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/client/onboarding/resume", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHealthCheck(t *testing.T) {
	mockSvc := new(mockClientService)
	handler := transporthttp.NewHandler(mockSvc)
	router := transporthttp.SetupRouter(handler)

	mockSvc.On("HealthCheck", mock.Anything).Return(true, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
