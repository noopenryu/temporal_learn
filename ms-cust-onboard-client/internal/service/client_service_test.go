package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"ms-cust-onboard-client/internal/domain"
	"ms-cust-onboard-client/internal/downstream"
	"ms-cust-onboard-client/internal/service"
)

type mockCredentialRepo struct {
	mock.Mock
}

func (m *mockCredentialRepo) Save(ctx context.Context, cred *domain.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *mockCredentialRepo) GetByPhoneNumber(ctx context.Context, phone string) (*domain.Credential, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Credential), args.Error(1)
}

func (m *mockCredentialRepo) GetByApplicationID(ctx context.Context, appID uuid.UUID) (*domain.Credential, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Credential), args.Error(1)
}

type mockDownstreamClient struct {
	mock.Mock
}

func (m *mockDownstreamClient) StartSession(ctx context.Context, req downstream.StartSessionRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) GetApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) GetActiveByPhone(ctx context.Context, phone string) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) ResumeSession(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) CancelSession(ctx context.Context, id uuid.UUID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockDownstreamClient) SubmitApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchPersonalInfo(ctx context.Context, id uuid.UUID, req downstream.PatchPersonalInfoRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchAddress(ctx context.Context, id uuid.UUID, req downstream.PatchAddressRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchIdentityDocs(ctx context.Context, id uuid.UUID, req downstream.PatchIdentityDocsRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchFinancialInfo(ctx context.Context, id uuid.UUID, req downstream.PatchFinancialRequest) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, tier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) ResubmitDocuments(ctx context.Context, id uuid.UUID, req downstream.ResubmitDocsRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *mockDownstreamClient) GetAuditLogs(ctx context.Context, id uuid.UUID) ([]downstream.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]downstream.AuditLog), args.Error(1)
}

func (m *mockDownstreamClient) GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*downstream.WorkflowProgressResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.WorkflowProgressResponse), args.Error(1)
}

func (m *mockDownstreamClient) PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*downstream.DetailResponse, error) {
	args := m.Called(ctx, id, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*downstream.DetailResponse), args.Error(1)
}

func (m *mockDownstreamClient) HealthCheck(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func TestStartSession_NewCustomer(t *testing.T) {
	repo := new(mockCredentialRepo)
	downstreamClient := new(mockDownstreamClient)
	svc := service.NewClientService(repo, downstreamClient)

	appID := uuid.New()
	expectedDetail := &downstream.DetailResponse{
		Application: downstream.Application{
			ID:          appID,
			CustomerID:  "cust_123",
			Status:      "DRAFT",
			CurrentStep: "personal_info",
			PersonalInfo: downstream.PersonalInfo{
				DateOfBirth: "0001-01-01T00:00:00Z", // downstream zero date
			},
		},
	}

	repo.On("GetByPhoneNumber", mock.Anything, "+15551234567").Return(nil, domain.ErrCredentialNotFound)
	downstreamClient.On("StartSession", mock.Anything, downstream.StartSessionRequest{
		Email:       "alice@example.com",
		PhoneNumber: "+15551234567",
	}).Return(expectedDetail, nil)
	repo.On("Save", mock.Anything, mock.MatchedBy(func(c *domain.Credential) bool {
		return c.PhoneNumber == "+15551234567" && c.ApplicationID == appID
	})).Return(nil)

	res, err := svc.StartSession(context.Background(), service.StartSessionInput{
		Email:       "alice@example.com",
		PhoneNumber: "+15551234567",
		Password:    "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	// Zero date must be cleaned to empty
	assert.Equal(t, "", res.Application.PersonalInfo.DateOfBirth)
}

func TestResumeSession_Success(t *testing.T) {
	repo := new(mockCredentialRepo)
	downstreamClient := new(mockDownstreamClient)
	svc := service.NewClientService(repo, downstreamClient)

	appID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	cred := &domain.Credential{
		PhoneNumber:   "+15551234567",
		ApplicationID: appID,
		PasswordHash:  string(hash),
	}

	repo.On("GetByPhoneNumber", mock.Anything, "+15551234567").Return(cred, nil)
	downstreamClient.On("GetApplication", mock.Anything, appID).Return(&downstream.DetailResponse{
		Application: downstream.Application{
			ID:     appID,
			Status: "DRAFT",
		},
	}, nil)
	downstreamClient.On("ResumeSession", mock.Anything, appID).Return(&downstream.DetailResponse{
		Application: downstream.Application{
			ID:     appID,
			Status: "DRAFT",
		},
	}, nil)

	res, err := svc.ResumeSession(context.Background(), service.ResumeSessionInput{
		PhoneNumber: "+15551234567",
		Password:    "secret123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsResumed)
}

func TestResumeSession_InvalidPassword(t *testing.T) {
	repo := new(mockCredentialRepo)
	downstreamClient := new(mockDownstreamClient)
	svc := service.NewClientService(repo, downstreamClient)

	appID := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	cred := &domain.Credential{
		PhoneNumber:   "+15551234567",
		ApplicationID: appID,
		PasswordHash:  string(hash),
	}

	repo.On("GetByPhoneNumber", mock.Anything, "+15551234567").Return(cred, nil)

	res, err := svc.ResumeSession(context.Background(), service.ResumeSessionInput{
		PhoneNumber: "+15551234567",
		Password:    "wrong-password",
	})

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
}

func TestPatchPersonalInfo_RejectsUnderageAndZeroDate(t *testing.T) {
	repo := new(mockCredentialRepo)
	downstreamClient := new(mockDownstreamClient)
	svc := service.NewClientService(repo, downstreamClient)

	appID := uuid.New()

	// 1. Year 0001 must be rejected
	_, err := svc.PatchPersonalInfo(context.Background(), appID, service.PatchPersonalInfoInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		DateOfBirth: "0001-01-01",
		Nationality: "US",
		PhoneNumber: "+15551234567",
		Email:       "alice@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be year 0001")

	// 2. Underage (<18) must be rejected
	underageDob := time.Now().AddDate(-17, 0, 0).Format("2006-01-02")
	_, err = svc.PatchPersonalInfo(context.Background(), appID, service.PatchPersonalInfoInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		DateOfBirth: underageDob,
		Nationality: "US",
		PhoneNumber: "+15551234567",
		Email:       "alice@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 18 years old")
}
