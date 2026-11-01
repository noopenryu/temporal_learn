package http

import (
	"time"

	"ms-customer-onboard/internal/domain"
)

// StartSessionRequest payload to initiate onboarding
type StartSessionRequest struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

// PatchPersonalInfoRequest payload for personal info step
type PatchPersonalInfoRequest struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}

func (r PatchPersonalInfoRequest) ToDomain() domain.PersonalInfo {
	return domain.PersonalInfo{
		FirstName:   r.FirstName,
		LastName:    r.LastName,
		DateOfBirth: r.DateOfBirth,
		Gender:      r.Gender,
		Nationality: r.Nationality,
		PhoneNumber: r.PhoneNumber,
		Email:       r.Email,
	}
}

// PatchAddressRequest payload for address info step
type PatchAddressRequest struct {
	Country           string `json:"country"`
	State             string `json:"state,omitempty"`
	City              string `json:"city"`
	StreetAddress     string `json:"street_address"`
	PostalCode        string `json:"postal_code"`
	ProofOfAddressURL string `json:"proof_of_address_url,omitempty"`
}

func (r PatchAddressRequest) ToDomain() domain.AddressInfo {
	return domain.AddressInfo{
		Country:           r.Country,
		State:             r.State,
		City:              r.City,
		StreetAddress:     r.StreetAddress,
		PostalCode:        r.PostalCode,
		ProofOfAddressURL: r.ProofOfAddressURL,
	}
}

// PatchIdentityDocsRequest payload for document upload step
type PatchIdentityDocsRequest struct {
	DocumentType     string `json:"document_type"`
	DocumentNumber   string `json:"document_number"`
	IssuingCountry   string `json:"issuing_country"`
	FrontDocumentURL string `json:"front_document_url"`
	BackDocumentURL  string `json:"back_document_url,omitempty"`
	SelfieURL        string `json:"selfie_url"`
}

func (r PatchIdentityDocsRequest) ToDomain() domain.IdentityDocuments {
	return domain.IdentityDocuments{
		DocumentType:     r.DocumentType,
		DocumentNumber:   r.DocumentNumber,
		IssuingCountry:   r.IssuingCountry,
		FrontDocumentURL: r.FrontDocumentURL,
		BackDocumentURL:  r.BackDocumentURL,
		SelfieURL:        r.SelfieURL,
	}
}

// PatchFinancialRequest payload for financial step
type PatchFinancialRequest struct {
	Occupation    string  `json:"occupation"`
	EmployerName  string  `json:"employer_name,omitempty"`
	AnnualIncome  float64 `json:"annual_income"`
	SourceOfFunds string  `json:"source_of_funds"`
	DeclaredPEP   bool    `json:"declared_pep"`
	TaxID         string  `json:"tax_id,omitempty"`
}

func (r PatchFinancialRequest) ToDomain() domain.FinancialInfo {
	return domain.FinancialInfo{
		Occupation:    r.Occupation,
		EmployerName:  r.EmployerName,
		AnnualIncome:  r.AnnualIncome,
		SourceOfFunds: r.SourceOfFunds,
		DeclaredPEP:   r.DeclaredPEP,
		TaxID:         r.TaxID,
	}
}

// PatchWalletTierRequest payload for tier selection
type PatchWalletTierRequest struct {
	WalletTier string `json:"wallet_tier"`
}

// CancelSessionRequest payload to abandon onboarding
type CancelSessionRequest struct {
	Reason string `json:"reason"`
}

// ManualReviewRequest payload for compliance officer decision
type ManualReviewRequest struct {
	ReviewerID    string                `json:"reviewer_id"`
	Approved      bool                  `json:"approved"`
	Notes         string                `json:"notes"`
	ConfirmedTier string                `json:"confirmed_tier,omitempty"`
	ConfirmedData *domain.ConfirmedData `json:"confirmed_data,omitempty"`
}

// PutConfirmedDataRequest payload for staff confirming customer details
type PutConfirmedDataRequest struct {
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

func (r PutConfirmedDataRequest) ToDomain() domain.ConfirmedData {
	return domain.ConfirmedData{
		ConfirmedAt:     time.Now(),
		ReviewerID:      r.ReviewerID,
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		DateOfBirth:     r.DateOfBirth,
		Nationality:     r.Nationality,
		DocumentType:    r.DocumentType,
		DocumentNumber:  r.DocumentNumber,
		Country:         r.Country,
		City:            r.City,
		StreetAddress:   r.StreetAddress,
		PostalCode:      r.PostalCode,
		AnnualIncome:    r.AnnualIncome,
		ApprovedTier:    r.ApprovedTier,
		ReviewChecklist: r.ReviewChecklist,
		Notes:           r.Notes,
	}
}

// ListApplicationsResponse payload for queue view
type ListApplicationsResponse struct {
	Items  []ApplicationDetailResponse `json:"items"`
	Total  int                         `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
}

// ApplicationDetailResponse is returned for comprehensive status display
type ApplicationDetailResponse struct {
	Application     *domain.OnboardingApplication `json:"application"`
	StepResults     []domain.StepCheckResult      `json:"step_results"`
	ProgressPercent int                           `json:"progress_percent"`
	CanSubmit       bool                          `json:"can_submit"`
	Uncompleted     []string                      `json:"uncompleted_steps,omitempty"`
	IsResumed       bool                          `json:"is_resumed,omitempty"`
}

func BuildDetailResponse(app *domain.OnboardingApplication) ApplicationDetailResponse {
	return BuildDetailResponseWithResume(app, false)
}

func BuildDetailResponseWithResume(app *domain.OnboardingApplication, isResumed bool) ApplicationDetailResponse {
	canSubmit, uncompleted := app.CanSubmit()
	return ApplicationDetailResponse{
		Application:     app,
		StepResults:     app.CalculateStepResults(),
		ProgressPercent: app.ProgressPercentage(),
		CanSubmit:       canSubmit,
		Uncompleted:     uncompleted,
		IsResumed:       isResumed,
	}
}
