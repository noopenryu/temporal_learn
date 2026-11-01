package activities_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"

	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/temporal/activities"
)

type mockRepo struct {
	apps map[uuid.UUID]*domain.OnboardingApplication
}

func (m *mockRepo) Create(_ context.Context, app *domain.OnboardingApplication) error {
	m.apps[app.ID] = app
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, ok := m.apps[id]
	if !ok {
		return nil, nil
	}
	return app, nil
}

func (m *mockRepo) GetActiveByCustomerID(_ context.Context, _ string) (*domain.OnboardingApplication, error) {
	return nil, nil
}

func (m *mockRepo) GetActiveByPhoneNumber(_ context.Context, _ string) (*domain.OnboardingApplication, error) {
	return nil, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.OnboardingApplication, int, error) {
	var list []*domain.OnboardingApplication
	for _, app := range m.apps {
		list = append(list, app)
	}
	return list, len(list), nil
}

func (m *mockRepo) Update(_ context.Context, app *domain.OnboardingApplication) error {
	m.apps[app.ID] = app
	return nil
}

func (m *mockRepo) SaveAuditLog(_ context.Context, _ *domain.AuditLog) error {
	return nil
}

func (m *mockRepo) GetAuditLogs(_ context.Context, _ uuid.UUID) ([]domain.AuditLog, error) {
	return nil, nil
}

func (m *mockRepo) Ping(_ context.Context) error {
	return nil
}

func TestOnboardingActivities(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	repo := &mockRepo{apps: make(map[uuid.UUID]*domain.OnboardingApplication)}
	appID := uuid.New()
	repo.apps[appID] = &domain.OnboardingApplication{
		ID:         appID,
		CustomerID: "cust-act-1",
		Status:     domain.StatusProcessing,
	}

	acts := activities.NewOnboardingActivities(repo)
	env.RegisterActivity(acts.UpdateApplicationStatusActivity)
	env.RegisterActivity(acts.ProvisionWalletAccountActivity)
	env.RegisterActivity(acts.SendOnboardingNotificationActivity)
	env.RegisterActivity(acts.ExecuteKycVerificationActivity)

	t.Run("UpdateApplicationStatusActivity", func(t *testing.T) {
		_, err := env.ExecuteActivity(acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
			ApplicationID:   appID.String(),
			Status:          domain.StatusApproved,
			ApprovedTier:    "TIER_1_BASIC",
			WalletAccountID: "WAL-TIER-123456",
		})
		assert.NoError(t, err)
		assert.Equal(t, domain.StatusApproved, repo.apps[appID].Status)
	})

	t.Run("ProvisionWalletAccountActivity", func(t *testing.T) {
		val, err := env.ExecuteActivity(acts.ProvisionWalletAccountActivity, activities.ProvisionWalletInput{
			CustomerID:   "cust-act-1",
			ApprovedTier: "TIER_2_PREMIUM",
			FullName:     "Dave Miller",
			Email:        "dave@example.com",
		})
		assert.NoError(t, err)

		var res activities.ProvisionWalletResult
		assert.NoError(t, val.Get(&res))
		assert.NotEmpty(t, res.WalletAccountID)
		assert.Equal(t, "TIER_2_PREMIUM", res.ApprovedTier)
	})

	t.Run("ExecuteKycVerificationActivity - Success", func(t *testing.T) {
		val, err := env.ExecuteActivity(acts.ExecuteKycVerificationActivity, activities.KycVerificationInput{
			ApplicationID:  appID.String(),
			CustomerID:     "cust-act-1",
			FullName:       "Dave Miller",
			DocumentType:   "NATIONAL_ID",
			DocumentNumber: "VN123456",
			FrontDocURL:    "https://storage/clear_front.jpg",
			SelfieURL:      "https://storage/clear_selfie.jpg",
		})
		assert.NoError(t, err)

		var res activities.KycVerificationResult
		assert.NoError(t, val.Get(&res))
		assert.True(t, res.Passed)
	})

	t.Run("ExecuteKycVerificationActivity - Blurry Document triggers re-upload", func(t *testing.T) {
		val, err := env.ExecuteActivity(acts.ExecuteKycVerificationActivity, activities.KycVerificationInput{
			ApplicationID:  appID.String(),
			CustomerID:     "cust-act-1",
			FullName:       "Dave Miller",
			DocumentType:   "NATIONAL_ID",
			DocumentNumber: "VN123456",
			FrontDocURL:    "https://storage/blurry_front.jpg",
			SelfieURL:      "https://storage/clear_selfie.jpg",
		})
		assert.NoError(t, err)

		var res activities.KycVerificationResult
		assert.NoError(t, val.Get(&res))
		assert.False(t, res.Passed)
		assert.True(t, res.RequiresDocumentResub)
	})
}
