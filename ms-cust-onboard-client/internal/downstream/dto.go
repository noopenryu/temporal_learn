package downstream

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type StartSessionRequest struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type PatchPersonalInfoRequest struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}

type PatchAddressRequest struct {
	Country           string `json:"country"`
	State             string `json:"state,omitempty"`
	City              string `json:"city"`
	StreetAddress     string `json:"street_address"`
	PostalCode        string `json:"postal_code"`
	ProofOfAddressURL string `json:"proof_of_address_url,omitempty"`
}

type PatchIdentityDocsRequest struct {
	DocumentType     string `json:"document_type"`
	DocumentNumber   string `json:"document_number"`
	IssuingCountry   string `json:"issuing_country"`
	FrontDocumentURL string `json:"front_document_url"`
	BackDocumentURL  string `json:"back_document_url,omitempty"`
	SelfieURL        string `json:"selfie_url"`
}

type PatchFinancialRequest struct {
	Occupation    string  `json:"occupation"`
	EmployerName  string  `json:"employer_name,omitempty"`
	AnnualIncome  float64 `json:"annual_income"`
	SourceOfFunds string  `json:"source_of_funds"`
	DeclaredPEP   bool    `json:"declared_pep"`
	TaxID         string  `json:"tax_id,omitempty"`
}

type PatchWalletTierRequest struct {
	WalletTier string `json:"wallet_tier"`
}

type CancelSessionRequest struct {
	Reason string `json:"reason"`
}

type WorkflowProgressResponse struct {
	Stage           string    `json:"stage"`
	PercentComplete int       `json:"percent_complete"`
	StatusMessage   string    `json:"status_message"`
	LastUpdated     time.Time `json:"last_updated"`
}

type ResubmitDocsRequest struct {
	DocumentType     string `json:"document_type,omitempty"`
	DocumentNumber   string `json:"document_number,omitempty"`
	IssuingCountry   string `json:"issuing_country,omitempty"`
	FrontDocumentURL string `json:"front_document_url,omitempty"`
	BackDocumentURL  string `json:"back_document_url,omitempty"`
	SelfieURL        string `json:"selfie_url,omitempty"`
}

type StepCheckResult struct {
	StepName      string   `json:"step_name"`
	Completed     bool     `json:"completed"`
	MissingFields []string `json:"missing_fields,omitempty"`
}

type PersonalInfo struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type AddressInfo struct {
	Country           string `json:"country"`
	State             string `json:"state,omitempty"`
	City              string `json:"city"`
	StreetAddress     string `json:"street_address"`
	PostalCode        string `json:"postal_code"`
	ProofOfAddressURL string `json:"proof_of_address_url,omitempty"`
}

type IdentityDocuments struct {
	DocumentType     string `json:"document_type"`
	DocumentNumber   string `json:"document_number"`
	IssuingCountry   string `json:"issuing_country"`
	FrontDocumentURL string `json:"front_document_url"`
	BackDocumentURL  string `json:"back_document_url,omitempty"`
	SelfieURL        string `json:"selfie_url"`
}

type FinancialInfo struct {
	Occupation    string  `json:"occupation"`
	EmployerName  string  `json:"employer_name,omitempty"`
	AnnualIncome  float64 `json:"annual_income"`
	SourceOfFunds string  `json:"source_of_funds"`
	DeclaredPEP   bool    `json:"declared_pep"`
	TaxID         string  `json:"tax_id,omitempty"`
}

type Application struct {
	ID                  uuid.UUID         `json:"id"`
	CustomerID          string            `json:"customer_id"`
	Status              string            `json:"status"`
	CurrentStep         string            `json:"current_step"`
	CompletedSteps      []string          `json:"completed_steps"`
	PersonalInfo        PersonalInfo      `json:"personal_info"`
	AddressInfo         AddressInfo       `json:"address_info"`
	IdentityDocs        IdentityDocuments `json:"identity_docs"`
	FinancialInfo       FinancialInfo     `json:"financial_info"`
	WalletTierRequested string            `json:"wallet_tier_requested"`
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

type DetailResponse struct {
	Application     Application       `json:"application"`
	StepResults     []StepCheckResult `json:"step_results"`
	ProgressPercent int               `json:"progress_percent"`
	CanSubmit       bool              `json:"can_submit"`
	Uncompleted     []string          `json:"uncompleted_steps,omitempty"`
	IsResumed       bool              `json:"is_resumed,omitempty"`
}

type AuditLog struct {
	ID            int                    `json:"id"`
	ApplicationID uuid.UUID              `json:"application_id"`
	Action        string                 `json:"action"`
	FromStatus    string                 `json:"from_status,omitempty"`
	ToStatus      string                 `json:"to_status,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// UnmarshalJSON custom logic for downstream Application to clean zero dates
func (a *Application) UnmarshalJSON(data []byte) error {
	type Alias Application
	aux := &struct {
		*Alias
		PersonalInfoRaw json.RawMessage `json:"personal_info"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.PersonalInfoRaw) > 0 {
		var pi struct {
			FirstName   string `json:"first_name"`
			LastName    string `json:"last_name"`
			DateOfBirth string `json:"date_of_birth"`
			Gender      string `json:"gender,omitempty"`
			Nationality string `json:"nationality"`
			PhoneNumber string `json:"phone_number"`
			Email       string `json:"email"`
		}
		if err := json.Unmarshal(aux.PersonalInfoRaw, &pi); err == nil {
			a.PersonalInfo.FirstName = pi.FirstName
			a.PersonalInfo.LastName = pi.LastName
			a.PersonalInfo.Gender = pi.Gender
			a.PersonalInfo.Nationality = pi.Nationality
			a.PersonalInfo.PhoneNumber = pi.PhoneNumber
			a.PersonalInfo.Email = pi.Email
			// Transform year 0001 to empty
			if pi.DateOfBirth == "" || len(pi.DateOfBirth) >= 4 && pi.DateOfBirth[:4] == "0001" {
				a.PersonalInfo.DateOfBirth = ""
			} else {
				a.PersonalInfo.DateOfBirth = pi.DateOfBirth
			}
		}
	}
	return nil
}
