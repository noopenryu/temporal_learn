package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// OnboardingStatus represents the lifecycle state of a customer onboarding application
type OnboardingStatus string

const (
	StatusDraft                OnboardingStatus = "DRAFT"
	StatusSubmitted            OnboardingStatus = "SUBMITTED"
	StatusProcessing           OnboardingStatus = "PROCESSING"
	StatusRequiresManualReview OnboardingStatus = "REQUIRES_MANUAL_REVIEW"
	StatusWaitingUpdatedDocs   OnboardingStatus = "WAITING_UPDATED_DOCS"
	StatusApproved             OnboardingStatus = "APPROVED"
	StatusRejected             OnboardingStatus = "REJECTED"
	StatusExpired              OnboardingStatus = "EXPIRED"
	StatusCancelled            OnboardingStatus = "CANCELLED"
)

// Step identifiers
const (
	StepPersonalInfo = "personal_info"
	StepAddressInfo  = "address_info"
	StepIdentityDocs = "identity_documents"
	StepFinancial    = "financial_info"
	StepWalletTier   = "wallet_tier"
)

// AllSteps list in sequential order for onboarding wizard
var AllSteps = []string{
	StepPersonalInfo,
	StepAddressInfo,
	StepIdentityDocs,
	StepFinancial,
	StepWalletTier,
}

// PersonalInfo holds personal details of the customer
type PersonalInfo struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality"` // ISO 3166-1 alpha-2 or alpha-3 (e.g. "VN", "SG")
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}

func (p PersonalInfo) FullName() string {
	return strings.TrimSpace(p.FirstName + " " + p.LastName)
}

// AddressInfo holds residential and contact address details
type AddressInfo struct {
	Country           string `json:"country"` // ISO country code
	State             string `json:"state,omitempty"`
	City              string `json:"city"`
	StreetAddress     string `json:"street_address"`
	PostalCode        string `json:"postal_code"`
	ProofOfAddressURL string `json:"proof_of_address_url,omitempty"`
}

// IdentityDocuments holds verification documents uploaded by applicant
type IdentityDocuments struct {
	DocumentType     string `json:"document_type"` // e.g. "NATIONAL_ID", "PASSPORT", "DRIVING_LICENSE"
	DocumentNumber   string `json:"document_number"`
	IssuingCountry   string `json:"issuing_country"`
	FrontDocumentURL string `json:"front_document_url"`
	BackDocumentURL  string `json:"back_document_url,omitempty"`
	SelfieURL        string `json:"selfie_url"`
}

// FinancialInfo holds financial details and regulatory declarations
type FinancialInfo struct {
	Occupation    string  `json:"occupation"`
	EmployerName  string  `json:"employer_name,omitempty"`
	AnnualIncome  float64 `json:"annual_income"`
	SourceOfFunds string  `json:"source_of_funds"`
	DeclaredPEP   bool    `json:"declared_pep"`
	TaxID         string  `json:"tax_id,omitempty"`
}

// ConfirmedData represents staff-verified information recorded during manual review
type ConfirmedData struct {
	ConfirmedAt     time.Time       `json:"confirmed_at"`
	ReviewerID      string          `json:"reviewer_id"`
	FirstName       string          `json:"first_name"`
	LastName        string          `json:"last_name"`
	DateOfBirth     string          `json:"date_of_birth"`
	Nationality     string          `json:"nationality"`
	DocumentType    string          `json:"document_type"`
	DocumentNumber  string          `json:"document_number"`
	Country         string          `json:"country"`
	City            string          `json:"city"`
	StreetAddress   string          `json:"street_address"`
	PostalCode      string          `json:"postal_code"`
	AnnualIncome    float64         `json:"annual_income"`
	ApprovedTier    string          `json:"approved_tier"`
	ReviewChecklist map[string]bool `json:"review_checklist,omitempty"`
	Notes           string          `json:"notes,omitempty"`
}

// StepCheckResult contains status and any missing fields for an individual onboarding step
type StepCheckResult struct {
	StepName      string   `json:"step_name"`
	Completed     bool     `json:"completed"`
	MissingFields []string `json:"missing_fields,omitempty"`
}

// OnboardingApplication represents the central aggregate root for customer onboarding
type OnboardingApplication struct {
	ID                  uuid.UUID         `json:"id"`
	CustomerID          string            `json:"customer_id"`
	Status              OnboardingStatus  `json:"status"`
	CurrentStep         string            `json:"current_step"`
	CompletedSteps      []string          `json:"completed_steps"`
	PersonalInfo        PersonalInfo      `json:"personal_info"`
	AddressInfo         AddressInfo       `json:"address_info"`
	IdentityDocs        IdentityDocuments `json:"identity_docs"`
	FinancialInfo       FinancialInfo     `json:"financial_info"`
	ConfirmedData       *ConfirmedData    `json:"confirmed_data,omitempty"`
	WalletTierRequested string            `json:"wallet_tier_requested"` // "TIER_1_BASIC" or "TIER_2_PREMIUM"
	ApprovedTier        string            `json:"approved_tier,omitempty"`
	WalletAccountID     string            `json:"wallet_account_id,omitempty"`
	RejectionReasons    []string          `json:"rejection_reasons,omitempty"`
	WorkflowID          string            `json:"workflow_id,omitempty"`
	RunID               string            `json:"run_id,omitempty"`
	ExpiresAt           time.Time         `json:"expires_at"`
	SubmittedAt         *time.Time        `json:"submitted_at,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// IsExpired checks if the draft session has passed its expiration time
func (a *OnboardingApplication) IsExpired() bool {
	return a.Status == StatusDraft && time.Now().After(a.ExpiresAt)
}

// CanBeModified returns true if the customer is allowed to patch/update the application
func (a *OnboardingApplication) CanBeModified() bool {
	return a.Status == StatusDraft || a.Status == StatusWaitingUpdatedDocs
}

// CalculateStepResults checks each step and returns its completion and missing fields
func (a *OnboardingApplication) CalculateStepResults() []StepCheckResult {
	results := make([]StepCheckResult, 0, len(AllSteps))

	// 1. Personal Info
	var missingPersonal []string
	if strings.TrimSpace(a.PersonalInfo.FirstName) == "" {
		missingPersonal = append(missingPersonal, "first_name")
	}
	if strings.TrimSpace(a.PersonalInfo.LastName) == "" {
		missingPersonal = append(missingPersonal, "last_name")
	}
	if a.PersonalInfo.DateOfBirth.IsZero() {
		missingPersonal = append(missingPersonal, "date_of_birth")
	}
	if strings.TrimSpace(a.PersonalInfo.Nationality) == "" {
		missingPersonal = append(missingPersonal, "nationality")
	}
	if strings.TrimSpace(a.PersonalInfo.PhoneNumber) == "" {
		missingPersonal = append(missingPersonal, "phone_number")
	}
	if strings.TrimSpace(a.PersonalInfo.Email) == "" {
		missingPersonal = append(missingPersonal, "email")
	}
	results = append(results, StepCheckResult{
		StepName:      StepPersonalInfo,
		Completed:     len(missingPersonal) == 0,
		MissingFields: missingPersonal,
	})

	// 2. Address Info
	var missingAddress []string
	if strings.TrimSpace(a.AddressInfo.Country) == "" {
		missingAddress = append(missingAddress, "country")
	}
	if strings.TrimSpace(a.AddressInfo.City) == "" {
		missingAddress = append(missingAddress, "city")
	}
	if strings.TrimSpace(a.AddressInfo.StreetAddress) == "" {
		missingAddress = append(missingAddress, "street_address")
	}
	if strings.TrimSpace(a.AddressInfo.PostalCode) == "" {
		missingAddress = append(missingAddress, "postal_code")
	}
	results = append(results, StepCheckResult{
		StepName:      StepAddressInfo,
		Completed:     len(missingAddress) == 0,
		MissingFields: missingAddress,
	})

	// 3. Identity Documents
	var missingDocs []string
	if strings.TrimSpace(a.IdentityDocs.DocumentType) == "" {
		missingDocs = append(missingDocs, "document_type")
	}
	if strings.TrimSpace(a.IdentityDocs.DocumentNumber) == "" {
		missingDocs = append(missingDocs, "document_number")
	}
	if strings.TrimSpace(a.IdentityDocs.IssuingCountry) == "" {
		missingDocs = append(missingDocs, "issuing_country")
	}
	if strings.TrimSpace(a.IdentityDocs.FrontDocumentURL) == "" {
		missingDocs = append(missingDocs, "front_document_url")
	}
	if strings.TrimSpace(a.IdentityDocs.SelfieURL) == "" {
		missingDocs = append(missingDocs, "selfie_url")
	}
	results = append(results, StepCheckResult{
		StepName:      StepIdentityDocs,
		Completed:     len(missingDocs) == 0,
		MissingFields: missingDocs,
	})

	// 4. Financial Info
	var missingFinancial []string
	if strings.TrimSpace(a.FinancialInfo.Occupation) == "" {
		missingFinancial = append(missingFinancial, "occupation")
	}
	if strings.TrimSpace(a.FinancialInfo.SourceOfFunds) == "" {
		missingFinancial = append(missingFinancial, "source_of_funds")
	}
	// For Premium tier, income must be greater than 0
	if a.WalletTierRequested == "TIER_2_PREMIUM" && a.FinancialInfo.AnnualIncome <= 0 {
		missingFinancial = append(missingFinancial, "annual_income_greater_than_zero")
	}
	results = append(results, StepCheckResult{
		StepName:      StepFinancial,
		Completed:     len(missingFinancial) == 0,
		MissingFields: missingFinancial,
	})

	// 5. Wallet Tier
	var missingTier []string
	if strings.TrimSpace(a.WalletTierRequested) == "" {
		missingTier = append(missingTier, "wallet_tier_requested")
	}
	results = append(results, StepCheckResult{
		StepName:      StepWalletTier,
		Completed:     len(missingTier) == 0,
		MissingFields: missingTier,
	})

	return results
}

// ProgressPercentage calculates the 0-100 completion percentage
func (a *OnboardingApplication) ProgressPercentage() int {
	stepResults := a.CalculateStepResults()
	completedCount := 0
	for _, res := range stepResults {
		if res.Completed {
			completedCount++
		}
	}
	if len(stepResults) == 0 {
		return 0
	}
	return (completedCount * 100) / len(stepResults)
}

// UpdateStepProgress updates the CompletedSteps list and sets the Next/CurrentStep
func (a *OnboardingApplication) UpdateStepProgress() {
	stepResults := a.CalculateStepResults()
	completed := make([]string, 0)
	var nextStep string

	for _, res := range stepResults {
		if res.Completed {
			completed = append(completed, res.StepName)
		} else if nextStep == "" {
			nextStep = res.StepName
		}
	}

	a.CompletedSteps = completed
	if nextStep != "" {
		a.CurrentStep = nextStep
	} else {
		a.CurrentStep = "review_and_submit"
	}
}

// CanSubmit returns true if all mandatory steps are complete and the application is in draft
func (a *OnboardingApplication) CanSubmit() (bool, []string) {
	if a.Status != StatusDraft {
		return false, []string{"application is not in draft status"}
	}
	if a.IsExpired() {
		return false, []string{"application has expired"}
	}

	stepResults := a.CalculateStepResults()
	var uncompleted []string
	for _, res := range stepResults {
		if !res.Completed {
			uncompleted = append(uncompleted, res.StepName+": "+strings.Join(res.MissingFields, ", "))
		}
	}

	return len(uncompleted) == 0, uncompleted
}

// AuditLog records state transitions and significant patch events
type AuditLog struct {
	ID            int64          `json:"id"`
	ApplicationID uuid.UUID      `json:"application_id"`
	Action        string         `json:"action"`
	FromStatus    string         `json:"from_status,omitempty"`
	ToStatus      string         `json:"to_status,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}
