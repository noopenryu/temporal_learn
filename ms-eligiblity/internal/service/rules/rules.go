package rules

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ms-eligiblity/internal/domain"
)

// Rule represents an individual eligibility check rule
type Rule interface {
	Name() string
	Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult
}

// ---------------- Age Verification Rule ----------------

// AgeRule validates customer age against minimum threshold
type AgeRule struct {
	MinAge int
	NowFn  func() time.Time
}

func NewAgeRule(minAge int) *AgeRule {
	return &AgeRule{
		MinAge: minAge,
		NowFn:  time.Now,
	}
}

func (r *AgeRule) Name() string {
	return "AgeVerificationRule"
}

func (r *AgeRule) Evaluate(_ context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	now := r.NowFn()
	age := calculateAge(req.DateOfBirth, now)

	if age < r.MinAge {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityHardReject,
			Reason:   fmt.Sprintf("Customer age is %d, which is below the minimum required age of %d", age, r.MinAge),
			Details: map[string]any{
				"calculated_age": age,
				"required_age":   r.MinAge,
			},
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   fmt.Sprintf("Customer meets minimum age requirement (%d >= %d)", age, r.MinAge),
		Details: map[string]any{
			"calculated_age": age,
		},
	}
}

func calculateAge(birthDate, now time.Time) int {
	years := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		years--
	}
	return years
}

// ---------------- Sanctioned Jurisdiction Rule ----------------

// JurisdictionRule checks for FATF blacklisted / sanctioned countries
type JurisdictionRule struct {
	SanctionedCountries map[string]string // ISO alpha-2 or alpha-3 -> Country name
}

func NewJurisdictionRule() *JurisdictionRule {
	return &JurisdictionRule{
		SanctionedCountries: map[string]string{
			"KP":  "Democratic People's Republic of Korea",
			"PRK": "Democratic People's Republic of Korea",
			"IR":  "Islamic Republic of Iran",
			"IRN": "Islamic Republic of Iran",
			"SY":  "Syrian Arab Republic",
			"SYR": "Syrian Arab Republic",
			"CU":  "Republic of Cuba",
			"CUB": "Republic of Cuba",
			"MM":  "Myanmar",
			"MMR": "Myanmar",
		},
	}
}

func (r *JurisdictionRule) Name() string {
	return "JurisdictionSanctionsRule"
}

func (r *JurisdictionRule) Evaluate(_ context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	nat := strings.ToUpper(strings.TrimSpace(req.Nationality))
	res := strings.ToUpper(strings.TrimSpace(req.ResidenceCountry))

	if countryName, exists := r.SanctionedCountries[nat]; exists {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityHardReject,
			Reason:   fmt.Sprintf("Applicant nationality (%s - %s) is on the restricted sanctions list", nat, countryName),
			Details: map[string]any{
				"restricted_field": "nationality",
				"country_code":     nat,
				"country_name":     countryName,
			},
		}
	}

	if countryName, exists := r.SanctionedCountries[res]; exists {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityHardReject,
			Reason:   fmt.Sprintf("Applicant residence country (%s - %s) is on the restricted sanctions list", res, countryName),
			Details: map[string]any{
				"restricted_field": "residence_country",
				"country_code":     res,
				"country_name":     countryName,
			},
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   "Applicant jurisdiction is permitted and compliant with international sanctions standards",
	}
}

// ---------------- Sanctions & AML Blacklist Screening Rule ----------------

// BlacklistRule checks against sanctions and internal fraud lists via repository
type BlacklistRule struct {
	Repo domain.EligibilityRepository
}

func NewBlacklistRule(repo domain.EligibilityRepository) *BlacklistRule {
	return &BlacklistRule{Repo: repo}
}

func (r *BlacklistRule) Name() string {
	return "SanctionsAndBlacklistRule"
}

func (r *BlacklistRule) Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	if r.Repo == nil {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   true,
			Severity: domain.SeverityInfo,
			Reason:   "Blacklist screening repository not configured (skipped)",
		}
	}

	hit, reason, err := r.Repo.IsSanctionedOrBlacklisted(ctx, req.NationalID, req.FullName(), req.Nationality)
	if err != nil {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityManualReview,
			Reason:   fmt.Sprintf("Blacklist screening lookup error: %v", err),
		}
	}

	if hit {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityHardReject,
			Reason:   fmt.Sprintf("Sanctions/Blacklist hit: %s", reason),
			Details: map[string]any{
				"blacklist_reason": reason,
			},
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   "Sanctions and AML blacklist screening clear",
	}
}

// ---------------- Duplicate Identity Rule ----------------

// DuplicateIdentityRule ensures customer's national ID is not already bound to an active wallet
type DuplicateIdentityRule struct {
	Repo domain.EligibilityRepository
}

func NewDuplicateIdentityRule(repo domain.EligibilityRepository) *DuplicateIdentityRule {
	return &DuplicateIdentityRule{Repo: repo}
}

func (r *DuplicateIdentityRule) Name() string {
	return "DuplicateIdentityRule"
}

func (r *DuplicateIdentityRule) Evaluate(ctx context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	if r.Repo == nil {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   true,
			Severity: domain.SeverityInfo,
			Reason:   "Duplicate check repository not configured (skipped)",
		}
	}

	isRegistered, err := r.Repo.IsNationalIDRegistered(ctx, req.NationalID, req.IDType)
	if err != nil {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityManualReview,
			Reason:   fmt.Sprintf("Duplicate identity verification failed due to system lookup error: %v", err),
		}
	}

	if isRegistered {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityHardReject,
			Reason:   fmt.Sprintf("National ID %s is already associated with an existing active customer account", req.NationalID),
			Details: map[string]any{
				"national_id": req.NationalID,
				"id_type":     req.IDType,
			},
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   "No duplicate identity detected; customer identity is unique",
	}
}

// ---------------- Politically Exposed Person (PEP) Rule ----------------

// PEPRule triggers manual review if the customer is declared as a Politically Exposed Person
type PEPRule struct{}

func NewPEPRule() *PEPRule {
	return &PEPRule{}
}

func (r *PEPRule) Name() string {
	return "PoliticallyExposedPersonRule"
}

func (r *PEPRule) Evaluate(_ context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	if req.DeclaredPEP {
		return domain.RuleEvaluationResult{
			RuleName: r.Name(),
			Passed:   false,
			Severity: domain.SeverityManualReview,
			Reason:   "Applicant declared as Politically Exposed Person (PEP); Enhanced Due Diligence (EDD) required",
			Details: map[string]any{
				"declared_pep": true,
			},
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   "Non-PEP applicant; standard due diligence applies",
	}
}

// ---------------- Wallet Tier Rule ----------------

// WalletTierRule validates applicant qualification for higher tiers (e.g., Premium wallet tier)
type WalletTierRule struct{}

func NewWalletTierRule() *WalletTierRule {
	return &WalletTierRule{}
}

func (r *WalletTierRule) Name() string {
	return "WalletTierRule"
}

func (r *WalletTierRule) Evaluate(_ context.Context, req *domain.EligibilityCheckRequest) domain.RuleEvaluationResult {
	if strings.EqualFold(req.WalletTierRequested, "TIER_2_PREMIUM") {
		// Premium tier requires declared annual income > 0 or manual verification
		if req.AnnualIncome <= 0 {
			return domain.RuleEvaluationResult{
				RuleName: r.Name(),
				Passed:   false,
				Severity: domain.SeverityManualReview,
				Reason:   "Tier 2 Premium wallet requested without declared income; requires compliance approval",
				Details: map[string]any{
					"tier_requested": req.WalletTierRequested,
					"annual_income":  req.AnnualIncome,
				},
			}
		}
	}

	return domain.RuleEvaluationResult{
		RuleName: r.Name(),
		Passed:   true,
		Severity: domain.SeverityInfo,
		Reason:   fmt.Sprintf("Applicant qualified for %s tier", req.WalletTierRequested),
	}
}
