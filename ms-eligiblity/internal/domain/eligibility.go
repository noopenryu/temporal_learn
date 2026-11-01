package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EligibilityStatus represents the final status of an eligibility evaluation
type EligibilityStatus string

const (
	StatusApproved             EligibilityStatus = "APPROVED"
	StatusRejected             EligibilityStatus = "REJECTED"
	StatusRequiresManualReview EligibilityStatus = "REQUIRES_MANUAL_REVIEW"
	StatusPending              EligibilityStatus = "PENDING"
)

// RiskLevel classifies the calculated risk of an onboarding applicant
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// RuleSeverity indicates how a rule failure impacts the outcome
type RuleSeverity string

const (
	SeverityHardReject   RuleSeverity = "HARD_REJECT"
	SeverityManualReview RuleSeverity = "MANUAL_REVIEW"
	SeverityInfo         RuleSeverity = "INFO"
)

// EligibilityCheckRequest contains all customer data submitted for eligibility evaluation
type EligibilityCheckRequest struct {
	TraceID             string    `json:"trace_id,omitempty"`
	CustomerID          string    `json:"customer_id"`
	NationalID          string    `json:"national_id"`
	IDType              string    `json:"id_type"` // e.g. "NATIONAL_ID", "PASSPORT", "DRIVING_LICENSE"
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	DateOfBirth         time.Time `json:"date_of_birth"`
	Nationality         string    `json:"nationality"`         // ISO 3166-1 alpha-2 or alpha-3 (e.g. "VN", "SG", "US")
	ResidenceCountry    string    `json:"residence_country"`    // ISO 3166-1 alpha-2 or alpha-3
	Email               string    `json:"email"`
	PhoneNumber         string    `json:"phone_number"`
	DeclaredPEP         bool      `json:"declared_pep"`         // Politically Exposed Person declaration
	AnnualIncome        float64   `json:"annual_income,omitempty"`
	WalletTierRequested string    `json:"wallet_tier_requested"` // "TIER_1_BASIC" or "TIER_2_PREMIUM"
}

// FullName returns the combined first and last name
func (r *EligibilityCheckRequest) FullName() string {
	return strings.TrimSpace(r.FirstName + " " + r.LastName)
}

// Validate checks essential input fields
func (r *EligibilityCheckRequest) Validate() error {
	if strings.TrimSpace(r.CustomerID) == "" {
		return errors.New("customer_id is required")
	}
	if strings.TrimSpace(r.NationalID) == "" {
		return errors.New("national_id is required")
	}
	if strings.TrimSpace(r.FirstName) == "" || strings.TrimSpace(r.LastName) == "" {
		return errors.New("first_name and last_name are required")
	}
	if r.DateOfBirth.IsZero() {
		return errors.New("date_of_birth is required")
	}
	if strings.TrimSpace(r.Nationality) == "" {
		return errors.New("nationality is required")
	}
	if strings.TrimSpace(r.ResidenceCountry) == "" {
		return errors.New("residence_country is required")
	}
	if strings.TrimSpace(r.PhoneNumber) == "" {
		return errors.New("phone_number is required")
	}
	if strings.TrimSpace(r.WalletTierRequested) == "" {
		r.WalletTierRequested = "TIER_1_BASIC"
	}
	return nil
}

// RuleEvaluationResult represents the outcome of an individual rule check
type RuleEvaluationResult struct {
	RuleName string         `json:"rule_name"`
	Passed   bool           `json:"passed"`
	Severity RuleSeverity   `json:"severity"`
	Reason   string         `json:"reason,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

// EligibilityCheckResult contains the consolidated outcome and audit trail
type EligibilityCheckResult struct {
	ID               uuid.UUID              `json:"id"`
	CustomerID       string                 `json:"customer_id"`
	Status           EligibilityStatus      `json:"status"`
	RiskScore        int                    `json:"risk_score"`
	RiskLevel        RiskLevel              `json:"risk_level"`
	ApprovedTier     string                 `json:"approved_tier"`
	RejectionReasons []string               `json:"rejection_reasons"`
	RuleResults      []RuleEvaluationResult `json:"rule_results"`
	EvaluatedAt      time.Time              `json:"evaluated_at"`
}
