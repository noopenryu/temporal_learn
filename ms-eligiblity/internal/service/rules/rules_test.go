package rules_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/service/rules"
)

// MockRepo implements domain.EligibilityRepository for testing
type MockRepo struct {
	IsRegisteredFn func(nationalID, idType string) (bool, error)
	IsBlacklistedFn func(nationalID, fullName, country string) (bool, string, error)
}

func (m *MockRepo) SaveCheck(_ context.Context, _ *domain.EligibilityCheckResult, _ *domain.EligibilityCheckRequest) error {
	return nil
}
func (m *MockRepo) GetCheckByID(_ context.Context, _ uuid.UUID) (*domain.EligibilityCheckResult, error) {
	return nil, nil
}
func (m *MockRepo) GetLatestCheckByCustomerID(_ context.Context, _ string) (*domain.EligibilityCheckResult, error) {
	return nil, nil
}
func (m *MockRepo) IsNationalIDRegistered(_ context.Context, nationalID, idType string) (bool, error) {
	if m.IsRegisteredFn != nil {
		return m.IsRegisteredFn(nationalID, idType)
	}
	return false, nil
}
func (m *MockRepo) IsSanctionedOrBlacklisted(_ context.Context, nationalID, fullName, country string) (bool, string, error) {
	if m.IsBlacklistedFn != nil {
		return m.IsBlacklistedFn(nationalID, fullName, country)
	}
	return false, "", nil
}
func (m *MockRepo) RegisterExistingIdentity(_ context.Context, _, _, _, _, _ string) error {
	return nil
}
func (m *MockRepo) Ping(_ context.Context) error {
	return nil
}

func TestAgeRule(t *testing.T) {
	fixedNow := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ageRule := rules.NewAgeRule(18)
	ageRule.NowFn = func() time.Time { return fixedNow }

	t.Run("Underage applicant rejected", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			DateOfBirth: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC), // 16 years old
		}
		res := ageRule.Evaluate(context.Background(), req)
		assert.False(t, res.Passed)
		assert.Equal(t, domain.SeverityHardReject, res.Severity)
		assert.Contains(t, res.Reason, "below the minimum required age of 18")
	})

	t.Run("Adult applicant approved", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			DateOfBirth: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), // 26 years old
		}
		res := ageRule.Evaluate(context.Background(), req)
		assert.True(t, res.Passed)
		assert.Equal(t, domain.SeverityInfo, res.Severity)
	})
}

func TestJurisdictionRule(t *testing.T) {
	rule := rules.NewJurisdictionRule()

	t.Run("Sanctioned nationality rejected", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			Nationality:      "KP",
			ResidenceCountry: "VN",
		}
		res := rule.Evaluate(context.Background(), req)
		assert.False(t, res.Passed)
		assert.Equal(t, domain.SeverityHardReject, res.Severity)
		assert.Contains(t, res.Reason, "restricted sanctions list")
	})

	t.Run("Sanctioned residence country rejected", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			Nationality:      "VN",
			ResidenceCountry: "IR",
		}
		res := rule.Evaluate(context.Background(), req)
		assert.False(t, res.Passed)
		assert.Equal(t, domain.SeverityHardReject, res.Severity)
		assert.Contains(t, res.Reason, "restricted sanctions list")
	})

	t.Run("Allowed country approved", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			Nationality:      "VN",
			ResidenceCountry: "VN",
		}
		res := rule.Evaluate(context.Background(), req)
		assert.True(t, res.Passed)
	})
}

func TestDuplicateIdentityRule(t *testing.T) {
	t.Run("Already registered ID rejected", func(t *testing.T) {
		mock := &MockRepo{
			IsRegisteredFn: func(nationalID, idType string) (bool, error) {
				return true, nil
			},
		}
		rule := rules.NewDuplicateIdentityRule(mock)
		req := &domain.EligibilityCheckRequest{
			NationalID: "VN123456789",
			IDType:     "NATIONAL_ID",
		}

		res := rule.Evaluate(context.Background(), req)
		assert.False(t, res.Passed)
		assert.Equal(t, domain.SeverityHardReject, res.Severity)
		assert.Contains(t, res.Reason, "already associated with an existing active customer account")
	})

	t.Run("New unique ID approved", func(t *testing.T) {
		mock := &MockRepo{
			IsRegisteredFn: func(nationalID, idType string) (bool, error) {
				return false, nil
			},
		}
		rule := rules.NewDuplicateIdentityRule(mock)
		req := &domain.EligibilityCheckRequest{
			NationalID: "VN999999999",
			IDType:     "NATIONAL_ID",
		}

		res := rule.Evaluate(context.Background(), req)
		assert.True(t, res.Passed)
	})
}

func TestPEPRule(t *testing.T) {
	rule := rules.NewPEPRule()

	t.Run("Declared PEP triggers manual review", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			DeclaredPEP: true,
		}
		res := rule.Evaluate(context.Background(), req)
		assert.False(t, res.Passed)
		assert.Equal(t, domain.SeverityManualReview, res.Severity)
		assert.Contains(t, res.Reason, "Enhanced Due Diligence (EDD) required")
	})

	t.Run("Non PEP passes", func(t *testing.T) {
		req := &domain.EligibilityCheckRequest{
			DeclaredPEP: false,
		}
		res := rule.Evaluate(context.Background(), req)
		assert.True(t, res.Passed)
	})
}
