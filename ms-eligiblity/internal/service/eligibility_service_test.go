package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"ms-eligiblity/internal/config"
	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/service"
)

type MockRepository struct {
	savedCheck *domain.EligibilityCheckResult
	registeredIDs map[string]bool
	blacklistedIDs map[string]string
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		registeredIDs:  make(map[string]bool),
		blacklistedIDs: make(map[string]string),
	}
}

func (m *MockRepository) SaveCheck(_ context.Context, check *domain.EligibilityCheckResult, _ *domain.EligibilityCheckRequest) error {
	m.savedCheck = check
	return nil
}

func (m *MockRepository) GetCheckByID(_ context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
	if m.savedCheck != nil && m.savedCheck.ID == id {
		return m.savedCheck, nil
	}
	return nil, nil
}

func (m *MockRepository) GetLatestCheckByCustomerID(_ context.Context, customerID string) (*domain.EligibilityCheckResult, error) {
	if m.savedCheck != nil && m.savedCheck.CustomerID == customerID {
		return m.savedCheck, nil
	}
	return nil, nil
}

func (m *MockRepository) IsNationalIDRegistered(_ context.Context, nationalID, _ string) (bool, error) {
	return m.registeredIDs[nationalID], nil
}

func (m *MockRepository) IsSanctionedOrBlacklisted(_ context.Context, nationalID, _, _ string) (bool, string, error) {
	reason, found := m.blacklistedIDs[nationalID]
	return found, reason, nil
}

func (m *MockRepository) RegisterExistingIdentity(_ context.Context, _, nationalID, _, _, _ string) error {
	m.registeredIDs[nationalID] = true
	return nil
}

func (m *MockRepository) Ping(_ context.Context) error {
	return nil
}

func TestEligibilityService_Evaluate(t *testing.T) {
	cfg := &config.Config{
		MinEligibilityAge: 18,
	}

	t.Run("Standard adult applicant gets approved", func(t *testing.T) {
		repo := NewMockRepository()
		svc := service.NewEligibilityService(cfg, repo)

		req := &domain.EligibilityCheckRequest{
			CustomerID:          "cust-001",
			NationalID:          "VN123456789",
			IDType:              "NATIONAL_ID",
			FirstName:           "Van A",
			LastName:            "Nguyen",
			DateOfBirth:         time.Now().AddDate(-25, 0, 0),
			Nationality:         "VN",
			ResidenceCountry:    "VN",
			Email:               "a@example.com",
			PhoneNumber:         "+84901234567",
			DeclaredPEP:         false,
			WalletTierRequested: "TIER_1_BASIC",
		}

		res, err := svc.Evaluate(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, domain.StatusApproved, res.Status)
		assert.Equal(t, "TIER_1_BASIC", res.ApprovedTier)
		assert.Equal(t, domain.RiskLevelLow, res.RiskLevel)
		assert.Empty(t, res.RejectionReasons)
		assert.Equal(t, res, repo.savedCheck)
	})

	t.Run("Underage applicant is rejected", func(t *testing.T) {
		repo := NewMockRepository()
		svc := service.NewEligibilityService(cfg, repo)

		req := &domain.EligibilityCheckRequest{
			CustomerID:          "cust-minor",
			NationalID:          "VN999000111",
			IDType:              "NATIONAL_ID",
			FirstName:           "Minor",
			LastName:            "Kid",
			DateOfBirth:         time.Now().AddDate(-15, 0, 0),
			Nationality:         "VN",
			ResidenceCountry:    "VN",
			Email:               "minor@example.com",
			PhoneNumber:         "+84901112222",
			DeclaredPEP:         false,
			WalletTierRequested: "TIER_1_BASIC",
		}

		res, err := svc.Evaluate(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, domain.StatusRejected, res.Status)
		assert.Equal(t, "NONE", res.ApprovedTier)
		assert.NotEmpty(t, res.RejectionReasons)
		assert.Equal(t, 100, res.RiskScore)
	})

	t.Run("PEP applicant requires manual review", func(t *testing.T) {
		repo := NewMockRepository()
		svc := service.NewEligibilityService(cfg, repo)

		req := &domain.EligibilityCheckRequest{
			CustomerID:          "cust-pep",
			NationalID:          "VN777888999",
			IDType:              "NATIONAL_ID",
			FirstName:           "Senator",
			LastName:            "Public",
			DateOfBirth:         time.Now().AddDate(-45, 0, 0),
			Nationality:         "VN",
			ResidenceCountry:    "VN",
			Email:               "senator@example.com",
			PhoneNumber:         "+84903334444",
			DeclaredPEP:         true,
			WalletTierRequested: "TIER_1_BASIC",
		}

		res, err := svc.Evaluate(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, domain.StatusRequiresManualReview, res.Status)
		assert.Equal(t, "PENDING_REVIEW", res.ApprovedTier)
		assert.GreaterOrEqual(t, res.RiskScore, 50)
	})

	t.Run("Duplicate national ID is rejected", func(t *testing.T) {
		repo := NewMockRepository()
		repo.registeredIDs["VN_EXISTING_123"] = true

		svc := service.NewEligibilityService(cfg, repo)

		req := &domain.EligibilityCheckRequest{
			CustomerID:          "cust-dup",
			NationalID:          "VN_EXISTING_123",
			IDType:              "NATIONAL_ID",
			FirstName:           "Another",
			LastName:            "Person",
			DateOfBirth:         time.Now().AddDate(-30, 0, 0),
			Nationality:         "VN",
			ResidenceCountry:    "VN",
			Email:               "another@example.com",
			PhoneNumber:         "+84905556666",
			DeclaredPEP:         false,
			WalletTierRequested: "TIER_1_BASIC",
		}

		res, err := svc.Evaluate(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, domain.StatusRejected, res.Status)
		assert.NotEmpty(t, res.RejectionReasons)
	})
}
