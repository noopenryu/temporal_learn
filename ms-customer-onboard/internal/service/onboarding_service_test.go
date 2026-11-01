package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ms-customer-onboard/internal/config"
	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/service"
	"ms-customer-onboard/internal/temporal/workflows"
)

type mockRepository struct {
	applications map[uuid.UUID]*domain.OnboardingApplication
	auditLogs    []domain.AuditLog
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		applications: make(map[uuid.UUID]*domain.OnboardingApplication),
		auditLogs:    make([]domain.AuditLog, 0),
	}
}

func (m *mockRepository) Create(_ context.Context, app *domain.OnboardingApplication) error {
	m.applications[app.ID] = app
	return nil
}

func (m *mockRepository) GetByID(_ context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, ok := m.applications[id]
	if !ok {
		return nil, nil
	}
	clone := *app
	return &clone, nil
}

func (m *mockRepository) GetActiveByCustomerID(_ context.Context, customerID string) (*domain.OnboardingApplication, error) {
	for _, app := range m.applications {
		if app.CustomerID == customerID && app.Status != domain.StatusCancelled && app.Status != domain.StatusExpired && app.Status != domain.StatusRejected {
			clone := *app
			return &clone, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetActiveByPhoneNumber(_ context.Context, phone string) (*domain.OnboardingApplication, error) {
	norm := domain.NormalizePhoneNumber(phone)
	for _, app := range m.applications {
		if domain.NormalizePhoneNumber(app.PersonalInfo.PhoneNumber) == norm && app.Status != domain.StatusCancelled && app.Status != domain.StatusExpired {
			clone := *app
			return &clone, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) List(_ context.Context, filter domain.ListFilter) ([]*domain.OnboardingApplication, int, error) {
	var list []*domain.OnboardingApplication
	for _, app := range m.applications {
		if filter.Status != "" && filter.Status != "ALL" && string(app.Status) != filter.Status {
			continue
		}
		clone := *app
		list = append(list, &clone)
	}
	return list, len(list), nil
}

func (m *mockRepository) Update(_ context.Context, app *domain.OnboardingApplication) error {
	m.applications[app.ID] = app
	return nil
}

func (m *mockRepository) SaveAuditLog(_ context.Context, log *domain.AuditLog) error {
	m.auditLogs = append(m.auditLogs, *log)
	return nil
}

func (m *mockRepository) GetAuditLogs(_ context.Context, appID uuid.UUID) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	for _, l := range m.auditLogs {
		if l.ApplicationID == appID {
			logs = append(logs, l)
		}
	}
	return logs, nil
}

func (m *mockRepository) Ping(_ context.Context) error {
	return nil
}

func setupTestService() (service.OnboardingService, *mockRepository) {
	cfg := &config.Config{
		DraftTTLDays:      14,
		TemporalTaskQueue: "CUSTOMER_ONBOARDING_TASK_QUEUE",
	}
	repo := newMockRepository()
	svc := service.NewOnboardingService(cfg, repo, nil)
	return svc, repo
}

func TestOnboardingService_StartSession(t *testing.T) {
	svc, _ := setupTestService()
	ctx := context.Background()

	t.Run("Successfully start session", func(t *testing.T) {
		app, isResumed, err := svc.StartSession(ctx, "cust-1", "alice@example.com", "+8491234567")
		assert.NoError(t, err)
		assert.False(t, isResumed)
		assert.NotNil(t, app)
		assert.Equal(t, domain.StatusDraft, app.Status)
		assert.Equal(t, "cust-1", app.CustomerID)
		assert.Equal(t, "alice@example.com", app.PersonalInfo.Email)
	})

	t.Run("Starting session when active draft exists returns existing application", func(t *testing.T) {
		app1, isResumed1, err := svc.StartSession(ctx, "cust-2", "bob@example.com", "+8498765432")
		assert.NoError(t, err)
		assert.False(t, isResumed1)

		app2, isResumed2, err := svc.StartSession(ctx, "cust-2", "bob@example.com", "+8498765432")
		assert.NoError(t, err)
		assert.True(t, isResumed2)
		assert.Equal(t, app1.ID, app2.ID)
	})

	t.Run("Start session without customerID auto-generates customer_id", func(t *testing.T) {
		app, isResumed, err := svc.StartSession(ctx, "", "charlie@example.com", "+15551239999")
		assert.NoError(t, err)
		assert.False(t, isResumed)
		assert.NotNil(t, app)
		assert.NotEmpty(t, app.CustomerID)
		assert.Contains(t, app.CustomerID, "CUST-")
	})

	t.Run("Starting session with same phone number resumes in-progress session", func(t *testing.T) {
		app1, _, err := svc.StartSession(ctx, "", "dan@example.com", "+15557778888")
		assert.NoError(t, err)

		// Resume using slightly different formatting
		app2, isResumed2, err := svc.StartSession(ctx, "", "dan@example.com", "+1 (555) 777-8888")
		assert.NoError(t, err)
		assert.True(t, isResumed2)
		assert.Equal(t, app1.ID, app2.ID)
	})

	t.Run("Starting session when customer is already approved returns ErrAlreadyOnboarded", func(t *testing.T) {
		app, _, err := svc.StartSession(ctx, "", "eve@example.com", "+15559990000")
		assert.NoError(t, err)
		app.Status = domain.StatusApproved
		// update mock
		_, repo := setupTestService()
		_ = repo.Create(ctx, app)
		svcWithApproved := service.NewOnboardingService(&config.Config{DraftTTLDays: 14}, repo, nil)

		_, _, err = svcWithApproved.StartSession(ctx, "", "eve@example.com", "+15559990000")
		assert.ErrorIs(t, err, domain.ErrAlreadyOnboarded)
	})

	t.Run("GetActiveByPhoneNumber retrieves active application", func(t *testing.T) {
		app, _, err := svc.StartSession(ctx, "", "frank@example.com", "+15554443333")
		assert.NoError(t, err)

		found, err := svc.GetActiveByPhoneNumber(ctx, "+1 (555) 444-3333")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, app.ID, found.ID)
	})

	t.Run("Invalid start request fails validation", func(t *testing.T) {
		_, _, err := svc.StartSession(ctx, "", "bad-email", "")
		assert.Error(t, err)
	})
}

func TestOnboardingService_PatchSteps(t *testing.T) {
	svc, _ := setupTestService()
	ctx := context.Background()

	app, _, err := svc.StartSession(ctx, "cust-patch", "patch@example.com", "+123456789")
	assert.NoError(t, err)

	t.Run("Patch personal info", func(t *testing.T) {
		personal := domain.PersonalInfo{
			FirstName:   "Charlie",
			LastName:    "Brown",
			DateOfBirth: time.Now().AddDate(-30, 0, 0),
			Nationality: "US",
			PhoneNumber: "+123456789",
			Email:       "patch@example.com",
		}
		updated, err := svc.PatchPersonalInfo(ctx, app.ID, personal)
		assert.NoError(t, err)
		assert.Equal(t, "Charlie", updated.PersonalInfo.FirstName)
		assert.Contains(t, updated.CompletedSteps, domain.StepPersonalInfo)
	})

	t.Run("Patch address info", func(t *testing.T) {
		addr := domain.AddressInfo{
			Country:       "US",
			City:          "San Francisco",
			StreetAddress: "500 Market St",
			PostalCode:    "94105",
		}
		updated, err := svc.PatchAddress(ctx, app.ID, addr)
		assert.NoError(t, err)
		assert.Equal(t, "San Francisco", updated.AddressInfo.City)
		assert.Contains(t, updated.CompletedSteps, domain.StepAddressInfo)
	})

	t.Run("Patch identity docs", func(t *testing.T) {
		docs := domain.IdentityDocuments{
			DocumentType:     "PASSPORT",
			DocumentNumber:   "P99887766",
			IssuingCountry:   "US",
			FrontDocumentURL: "https://bucket/pass_front.jpg",
			SelfieURL:        "https://bucket/selfie.jpg",
		}
		updated, err := svc.PatchIdentityDocs(ctx, app.ID, docs)
		assert.NoError(t, err)
		assert.Equal(t, "P99887766", updated.IdentityDocs.DocumentNumber)
		assert.Contains(t, updated.CompletedSteps, domain.StepIdentityDocs)
	})

	t.Run("Patch financial info", func(t *testing.T) {
		fin := domain.FinancialInfo{
			Occupation:    "Product Manager",
			AnnualIncome:  85000,
			SourceOfFunds: "SALARY",
			DeclaredPEP:   false,
		}
		updated, err := svc.PatchFinancialInfo(ctx, app.ID, fin)
		assert.NoError(t, err)
		assert.Equal(t, "Product Manager", updated.FinancialInfo.Occupation)
		assert.Contains(t, updated.CompletedSteps, domain.StepFinancial)
	})

	t.Run("Patch wallet tier", func(t *testing.T) {
		updated, err := svc.PatchWalletTier(ctx, app.ID, "TIER_2_PREMIUM")
		assert.NoError(t, err)
		assert.Equal(t, "TIER_2_PREMIUM", updated.WalletTierRequested)
	})

	t.Run("Submit application when all steps are completed", func(t *testing.T) {
		submitted, err := svc.SubmitApplication(ctx, app.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusSubmitted, submitted.Status)
		assert.NotNil(t, submitted.SubmittedAt)
	})

	t.Run("Modifying an already submitted application returns error", func(t *testing.T) {
		_, err := svc.PatchPersonalInfo(ctx, app.ID, domain.PersonalInfo{
			FirstName: "ShouldFail",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be modified")
	})
}

func TestOnboardingService_ResubmitDocsAndManualReview(t *testing.T) {
	svc, repo := setupTestService()
	ctx := context.Background()

	appID := uuid.New()
	app := &domain.OnboardingApplication{
		ID:         appID,
		CustomerID: "cust-review",
		Status:     domain.StatusRequiresManualReview,
	}
	_ = repo.Create(ctx, app)

	t.Run("Submit manual review decision", func(t *testing.T) {
		decision := workflows.ManualReviewDecision{
			ReviewerID: "rev-01",
			Approved:   true,
			Notes:      "Identity confirmed",
		}
		err := svc.SubmitManualReviewDecision(ctx, appID, decision)
		assert.NoError(t, err)

		logs, err := svc.GetAuditLogs(ctx, appID)
		assert.NoError(t, err)
		assert.NotEmpty(t, logs)
		assert.Equal(t, "MANUAL_REVIEW_DECISION_RECORDED", logs[len(logs)-1].Action)
	})

	t.Run("Resubmit documents for application waiting on updated docs", func(t *testing.T) {
		app.Status = domain.StatusWaitingUpdatedDocs
		_ = repo.Update(ctx, app)

		docs := domain.IdentityDocuments{
			DocumentType:     "NATIONAL_ID",
			DocumentNumber:   "ID-RENEWED",
			IssuingCountry:   "VN",
			FrontDocumentURL: "https://bucket/clear_front.jpg",
			SelfieURL:        "https://bucket/clear_selfie.jpg",
		}
		err := svc.ResubmitDocuments(ctx, appID, docs)
		assert.NoError(t, err)

		reloaded, err := svc.GetApplication(ctx, appID)
		assert.NoError(t, err)
		assert.Equal(t, "ID-RENEWED", reloaded.IdentityDocs.DocumentNumber)
	})
}

func TestOnboardingService_CancelAndResume(t *testing.T) {
	svc, _ := setupTestService()
	ctx := context.Background()

	app, _, err := svc.StartSession(ctx, "cust-cancel", "cancel@example.com", "+11223344")
	assert.NoError(t, err)

	t.Run("Resume active session", func(t *testing.T) {
		resumed, err := svc.ResumeSession(ctx, app.ID)
		assert.NoError(t, err)
		assert.Equal(t, app.ID, resumed.ID)
	})

	t.Run("Cancel active session", func(t *testing.T) {
		err := svc.CancelSession(ctx, app.ID, "User changed their mind")
		assert.NoError(t, err)

		cancelled, err := svc.GetApplication(ctx, app.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusCancelled, cancelled.Status)

		// Resuming a cancelled session fails
		_, err = svc.ResumeSession(ctx, app.ID)
		assert.Error(t, err)
	})
}

func TestOnboardingService_AuditLogs(t *testing.T) {
	svc, _ := setupTestService()
	ctx := context.Background()

	app, _, err := svc.StartSession(ctx, "cust-audit", "audit@example.com", "+99887766")
	assert.NoError(t, err)

	logs, err := svc.GetAuditLogs(ctx, app.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Equal(t, "SESSION_STARTED", logs[0].Action)
}

func TestOnboardingService_StaffManualReviewAndConfirmedData(t *testing.T) {
	svc, repo := setupTestService()
	ctx := context.Background()

	app, _, err := svc.StartSession(ctx, "cust-staff-rev", "staffrev@example.com", "+15550001")
	assert.NoError(t, err)

	// Simulate app reaching manual review state
	app.Status = domain.StatusRequiresManualReview
	_ = repo.Update(ctx, app)

	t.Run("ListApplications with filter", func(t *testing.T) {
		apps, total, err := svc.ListApplications(ctx, domain.ListFilter{
			Status: "REQUIRES_MANUAL_REVIEW",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, app.ID, apps[0].ID)
	})

	t.Run("PutConfirmedData", func(t *testing.T) {
		confirmed := domain.ConfirmedData{
			ReviewerID:     "OFFICER-007",
			FirstName:      "John",
			LastName:       "Confirmed",
			DateOfBirth:    "1990-05-15",
			Nationality:    "VN",
			DocumentType:   "NATIONAL_ID",
			DocumentNumber: "VN-999888777",
			Country:        "VN",
			City:           "Hanoi",
			StreetAddress:  "123 Hoang Hoa Tham",
			PostalCode:     "100000",
			ApprovedTier:   "TIER_2_PREMIUM",
			Notes:          "Physical ID verified against identity DB",
		}

		updatedApp, err := svc.PutConfirmedData(ctx, app.ID, confirmed)
		assert.NoError(t, err)
		assert.NotNil(t, updatedApp.ConfirmedData)
		assert.Equal(t, "VN-999888777", updatedApp.ConfirmedData.DocumentNumber)
		assert.Equal(t, "TIER_2_PREMIUM", updatedApp.ApprovedTier)
	})

	t.Run("SubmitManualReviewDecision with confirmed data", func(t *testing.T) {
		decision := workflows.ManualReviewDecision{
			ReviewerID:    "OFFICER-007",
			Approved:      true,
			Notes:         "Manual approval granted with confirmed tier",
			ConfirmedTier: "TIER_2_PREMIUM",
		}

		err := svc.SubmitManualReviewDecision(ctx, app.ID, decision)
		assert.NoError(t, err)

		reloaded, err := svc.GetApplication(ctx, app.ID)
		assert.NoError(t, err)
		assert.Equal(t, "TIER_2_PREMIUM", reloaded.ApprovedTier)

		logs, err := svc.GetAuditLogs(ctx, app.ID)
		assert.NoError(t, err)
		foundReviewLog := false
		for _, l := range logs {
			if l.Action == "MANUAL_REVIEW_DECISION_RECORDED" {
				foundReviewLog = true
				break
			}
		}
		assert.True(t, foundReviewLog)
	})
}
