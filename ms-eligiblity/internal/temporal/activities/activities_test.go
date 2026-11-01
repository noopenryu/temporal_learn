package activities_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"

	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/temporal/activities"
)

type mockService struct {
	evalFn    func(ctx context.Context, req *domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error)
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error)
}

func (m *mockService) Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error) {
	if m.evalFn != nil {
		return m.evalFn(ctx, req)
	}
	return &domain.EligibilityCheckResult{
		ID:         uuid.New(),
		CustomerID: req.CustomerID,
		Status:     domain.StatusApproved,
	}, nil
}

func (m *mockService) GetCheckByID(ctx context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &domain.EligibilityCheckResult{
		ID:     id,
		Status: domain.StatusApproved,
	}, nil
}

func (m *mockService) GetLatestByCustomerID(_ context.Context, _ string) (*domain.EligibilityCheckResult, error) {
	return nil, nil
}

func (m *mockService) ListRules() []string {
	return []string{"AgeVerificationRule"}
}

func TestEvaluateEligibilityActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	mockSvc := &mockService{
		evalFn: func(_ context.Context, req *domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error) {
			return &domain.EligibilityCheckResult{
				ID:           uuid.New(),
				CustomerID:   req.CustomerID,
				Status:       domain.StatusApproved,
				ApprovedTier: "TIER_1_BASIC",
				RiskLevel:    domain.RiskLevelLow,
			}, nil
		},
	}

	act := activities.NewEligibilityActivities(mockSvc)
	env.RegisterActivity(act.EvaluateEligibilityActivity)

	req := domain.EligibilityCheckRequest{
		CustomerID:          "cust-test-1",
		NationalID:          "VN123456789",
		IDType:              "NATIONAL_ID",
		FirstName:           "John",
		LastName:            "Doe",
		DateOfBirth:         time.Now().AddDate(-25, 0, 0),
		Nationality:         "VN",
		ResidenceCountry:    "VN",
		PhoneNumber:         "+84901234567",
		WalletTierRequested: "TIER_1_BASIC",
	}

	val, err := env.ExecuteActivity(act.EvaluateEligibilityActivity, req)
	assert.NoError(t, err)

	var result domain.EligibilityCheckResult
	assert.NoError(t, val.Get(&result))
	assert.Equal(t, domain.StatusApproved, result.Status)
	assert.Equal(t, "cust-test-1", result.CustomerID)
}

func TestQueryEligibilityStatusActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	checkID := uuid.New()
	mockSvc := &mockService{
		getByIDFn: func(_ context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
			if id == checkID {
				return &domain.EligibilityCheckResult{
					ID:         checkID,
					CustomerID: "cust-test-1",
					Status:     domain.StatusApproved,
				}, nil
			}
			return nil, nil
		},
	}

	act := activities.NewEligibilityActivities(mockSvc)
	env.RegisterActivity(act.QueryEligibilityStatusActivity)

	t.Run("Existing check found", func(t *testing.T) {
		val, err := env.ExecuteActivity(act.QueryEligibilityStatusActivity, checkID.String())
		assert.NoError(t, err)

		var result domain.EligibilityCheckResult
		assert.NoError(t, val.Get(&result))
		assert.Equal(t, checkID, result.ID)
	})

	t.Run("Non-existent check returns error", func(t *testing.T) {
		unknownID := uuid.New()
		_, err := env.ExecuteActivity(act.QueryEligibilityStatusActivity, unknownID.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ERR_CHECK_NOT_FOUND")
	})
}
