package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ms-eligiblity/internal/config"
	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/service/rules"
)

// EligibilityService manages the evaluation and retrieval of customer eligibility
type EligibilityService interface {
	Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error)
	GetCheckByID(ctx context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error)
	GetLatestByCustomerID(ctx context.Context, customerID string) (*domain.EligibilityCheckResult, error)
	ListRules() []string
}

type eligibilityService struct {
	repo  domain.EligibilityRepository
	rules []rules.Rule
}

// NewEligibilityService instantiates the service with registered rules
func NewEligibilityService(cfg *config.Config, repo domain.EligibilityRepository) EligibilityService {
	ruleList := []rules.Rule{
		rules.NewAgeRule(cfg.MinEligibilityAge),
		rules.NewJurisdictionRule(),
		rules.NewBlacklistRule(repo),
		rules.NewDuplicateIdentityRule(repo),
		rules.NewPEPRule(),
		rules.NewWalletTierRule(),
	}

	return &eligibilityService{
		repo:  repo,
		rules: ruleList,
	}
}

// Evaluate runs all business rules, computes risk score, determines status, and persists the result
func (s *eligibilityService) Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) (*domain.EligibilityCheckResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid eligibility request: %w", err)
	}

	checkID := uuid.New()
	ruleResults := make([]domain.RuleEvaluationResult, 0, len(s.rules))
	rejectionReasons := make([]string, 0)
	hasHardReject := false
	hasManualReview := false

	// Execute all registered rules
	for _, r := range s.rules {
		res := r.Evaluate(ctx, req)
		ruleResults = append(ruleResults, res)

		if !res.Passed {
			switch res.Severity {
			case domain.SeverityHardReject:
				hasHardReject = true
				rejectionReasons = append(rejectionReasons, fmt.Sprintf("[%s] %s", res.RuleName, res.Reason))
			case domain.SeverityManualReview:
				hasManualReview = true
			}
		}
	}

	// Compute Risk Score (0-100)
	riskScore := s.calculateRiskScore(req, hasHardReject, hasManualReview)
	riskLevel := s.mapRiskLevel(riskScore)

	// Determine final eligibility status
	var finalStatus domain.EligibilityStatus
	var approvedTier string

	if hasHardReject {
		finalStatus = domain.StatusRejected
		approvedTier = "NONE"
	} else if hasManualReview || riskScore >= 50 {
		finalStatus = domain.StatusRequiresManualReview
		approvedTier = "PENDING_REVIEW"
	} else {
		finalStatus = domain.StatusApproved
		approvedTier = req.WalletTierRequested
		if approvedTier == "" {
			approvedTier = "TIER_1_BASIC"
		}
	}

	result := &domain.EligibilityCheckResult{
		ID:               checkID,
		CustomerID:       req.CustomerID,
		Status:           finalStatus,
		RiskScore:        riskScore,
		RiskLevel:        riskLevel,
		ApprovedTier:     approvedTier,
		RejectionReasons: rejectionReasons,
		RuleResults:      ruleResults,
		EvaluatedAt:      time.Now().UTC(),
	}

	// Persist the audit log if repository is provided
	if s.repo != nil {
		if err := s.repo.SaveCheck(ctx, result, req); err != nil {
			// Log error but still return result with evaluation details
			return result, fmt.Errorf("evaluation completed but failed to persist record: %w", err)
		}
	}

	return result, nil
}

// GetCheckByID retrieves evaluation by ID
func (s *eligibilityService) GetCheckByID(ctx context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured")
	}
	return s.repo.GetCheckByID(ctx, id)
}

// GetLatestByCustomerID retrieves the most recent evaluation for a customer
func (s *eligibilityService) GetLatestByCustomerID(ctx context.Context, customerID string) (*domain.EligibilityCheckResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured")
	}
	return s.repo.GetLatestCheckByCustomerID(ctx, customerID)
}

// ListRules returns names of all active evaluation rules
func (s *eligibilityService) ListRules() []string {
	names := make([]string, 0, len(s.rules))
	for _, r := range s.rules {
		names = append(names, r.Name())
	}
	return names
}

// calculateRiskScore estimates risk score (0-100)
func (s *eligibilityService) calculateRiskScore(req *domain.EligibilityCheckRequest, hasHardReject, hasManualReview bool) int {
	if hasHardReject {
		return 100
	}

	score := 10 // Baseline safe score

	// PEP adds substantial compliance risk
	if req.DeclaredPEP {
		score += 45
	}

	// Non-resident customer risk factor
	if !strings.EqualFold(req.Nationality, req.ResidenceCountry) {
		score += 15
	}

	// Young adult (18-21) risk factor
	age := time.Now().Year() - req.DateOfBirth.Year()
	if age >= 18 && age <= 21 {
		score += 10
	}

	// High tier request without high income
	if strings.EqualFold(req.WalletTierRequested, "TIER_2_PREMIUM") && req.AnnualIncome < 30000 {
		score += 15
	}

	if hasManualReview && score < 50 {
		score = 50
	}

	if score > 100 {
		score = 100
	}

	return score
}

// mapRiskLevel maps integer risk score to categorical level
func (s *eligibilityService) mapRiskLevel(score int) domain.RiskLevel {
	switch {
	case score < 25:
		return domain.RiskLevelLow
	case score < 50:
		return domain.RiskLevelMedium
	case score < 75:
		return domain.RiskLevelHigh
	default:
		return domain.RiskLevelCritical
	}
}
