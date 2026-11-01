package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"ms-cust-onboard-client/internal/domain"
	"ms-cust-onboard-client/internal/downstream"
	"ms-cust-onboard-client/internal/repository"
)

type StartSessionInput struct {
	CustomerID  string `json:"customer_id,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type ResumeSessionInput struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type PatchPersonalInfoInput struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type ClientService interface {
	StartSession(ctx context.Context, input StartSessionInput) (*downstream.DetailResponse, error)
	ResumeSession(ctx context.Context, input ResumeSessionInput) (*downstream.DetailResponse, error)
	GetApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error)
	CancelSession(ctx context.Context, id uuid.UUID, reason string) error
	SubmitApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error)
	PatchPersonalInfo(ctx context.Context, id uuid.UUID, input PatchPersonalInfoInput) (*downstream.DetailResponse, error)
	PatchAddress(ctx context.Context, id uuid.UUID, req downstream.PatchAddressRequest) (*downstream.DetailResponse, error)
	PatchIdentityDocs(ctx context.Context, id uuid.UUID, req downstream.PatchIdentityDocsRequest) (*downstream.DetailResponse, error)
	PatchFinancialInfo(ctx context.Context, id uuid.UUID, req downstream.PatchFinancialRequest) (*downstream.DetailResponse, error)
	PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*downstream.DetailResponse, error)
	ResubmitDocuments(ctx context.Context, id uuid.UUID, req downstream.ResubmitDocsRequest) error
	GetAuditLogs(ctx context.Context, id uuid.UUID) ([]downstream.AuditLog, error)
	GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*downstream.WorkflowProgressResponse, error)
	PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*downstream.DetailResponse, error)
	HealthCheck(ctx context.Context) (bool, error)
}

type clientService struct {
	repo       repository.CredentialRepository
	downstream downstream.CustomerOnboardClient
}

func NewClientService(repo repository.CredentialRepository, downstreamClient downstream.CustomerOnboardClient) ClientService {
	return &clientService{
		repo:       repo,
		downstream: downstreamClient,
	}
}

// CleanZeroDate transforms 0001-01-01 zero-date into empty string so the UI date picker starts fresh
func cleanDetailResponse(resp *downstream.DetailResponse) *downstream.DetailResponse {
	if resp == nil {
		return nil
	}
	dob := resp.Application.PersonalInfo.DateOfBirth
	if dob == "" || strings.HasPrefix(dob, "0001") || strings.HasPrefix(dob, "0000") {
		resp.Application.PersonalInfo.DateOfBirth = ""
	}
	return resp
}

func (s *clientService) StartSession(ctx context.Context, input StartSessionInput) (*downstream.DetailResponse, error) {
	// 1. Validate basic fields
	if err := domain.ValidatePhoneNumber(input.PhoneNumber); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	cleanPhone := domain.CleanPhoneNumber(input.PhoneNumber)

	// 2. Check if credentials already exist for this phone number
	existingCred, err := s.repo.GetByPhoneNumber(ctx, cleanPhone)
	if err == nil && existingCred != nil {
		// Existing customer session: verify password
		if bcryptErr := bcrypt.CompareHashAndPassword([]byte(existingCred.PasswordHash), []byte(input.Password)); bcryptErr != nil {
			return nil, domain.ErrInvalidCredentials
		}
		// Valid password: resume active application
		detail, resumeErr := s.downstream.ResumeSession(ctx, existingCred.ApplicationID)
		if resumeErr != nil {
			// Fallback to get active by phone
			detail, resumeErr = s.downstream.GetActiveByPhone(ctx, cleanPhone)
			if resumeErr != nil {
				return nil, fmt.Errorf("failed to resume application: %w", resumeErr)
			}
		}
		detail.IsResumed = true
		return cleanDetailResponse(detail), nil
	}

	// 3. Forward new session to downstream ms-customer-onboard
	downstreamReq := downstream.StartSessionRequest{
		CustomerID:  input.CustomerID,
		Email:       strings.TrimSpace(input.Email),
		PhoneNumber: cleanPhone,
	}
	detail, err := s.downstream.StartSession(ctx, downstreamReq)
	if err != nil {
		return nil, err
	}

	// 4. Hash password with bcrypt and save credentials
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	cred := &domain.Credential{
		PhoneNumber:   cleanPhone,
		ApplicationID: detail.Application.ID,
		PasswordHash:  string(hashBytes),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if saveErr := s.repo.Save(ctx, cred); saveErr != nil {
		// Log error, continue (in testing or degraded db)
		fmt.Printf("Warning: failed to save credentials: %v\n", saveErr)
	}

	return cleanDetailResponse(detail), nil
}

func (s *clientService) ResumeSession(ctx context.Context, input ResumeSessionInput) (*downstream.DetailResponse, error) {
	// 1. Validate basic inputs
	if err := domain.ValidatePhoneNumber(input.PhoneNumber); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, errors.New("password is required to resume application")
	}

	cleanPhone := domain.CleanPhoneNumber(input.PhoneNumber)

	// 2. Fetch credential
	cred, err := s.repo.GetByPhoneNumber(ctx, cleanPhone)
	if err != nil {
		return nil, domain.ErrCredentialNotFound
	}

	// 3. Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// 4. Retrieve application and trigger resume
	detail, err := s.downstream.GetApplication(ctx, cred.ApplicationID)
	if err != nil {
		// Try get active by phone
		detail, err = s.downstream.GetActiveByPhone(ctx, cleanPhone)
		if err != nil {
			return nil, fmt.Errorf("no active application found to resume: %w", err)
		}
	}

	// If application is in draft, call resume
	if detail.Application.Status == "DRAFT" {
		resumedDetail, resumeErr := s.downstream.ResumeSession(ctx, detail.Application.ID)
		if resumeErr == nil && resumedDetail != nil {
			detail = resumedDetail
		}
	}

	detail.IsResumed = true
	return cleanDetailResponse(detail), nil
}

func (s *clientService) GetApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	detail, err := s.downstream.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) CancelSession(ctx context.Context, id uuid.UUID, reason string) error {
	return s.downstream.CancelSession(ctx, id, reason)
}

func (s *clientService) SubmitApplication(ctx context.Context, id uuid.UUID) (*downstream.DetailResponse, error) {
	detail, err := s.downstream.SubmitApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) PatchPersonalInfo(ctx context.Context, id uuid.UUID, input PatchPersonalInfoInput) (*downstream.DetailResponse, error) {
	// 1. Field validations
	if err := domain.ValidateNames(input.FirstName, input.LastName); err != nil {
		return nil, err
	}
	if err := domain.ValidateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePhoneNumber(input.PhoneNumber); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Nationality) == "" {
		return nil, errors.New("nationality is required")
	}

	// 2. Validate Date of Birth (must not be year 0001, must not be future, age >= 18)
	parsedDob, err := domain.ValidateDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}

	// 3. Build downstream request
	req := downstream.PatchPersonalInfoRequest{
		FirstName:   strings.TrimSpace(input.FirstName),
		LastName:    strings.TrimSpace(input.LastName),
		DateOfBirth: parsedDob,
		Gender:      input.Gender,
		Nationality: strings.ToUpper(strings.TrimSpace(input.Nationality)),
		PhoneNumber: domain.CleanPhoneNumber(input.PhoneNumber),
		Email:       strings.TrimSpace(input.Email),
	}

	detail, err := s.downstream.PatchPersonalInfo(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) PatchAddress(ctx context.Context, id uuid.UUID, req downstream.PatchAddressRequest) (*downstream.DetailResponse, error) {
	if err := domain.ValidateAddress(req.Country, req.City, req.StreetAddress, req.PostalCode); err != nil {
		return nil, err
	}
	detail, err := s.downstream.PatchAddress(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) PatchIdentityDocs(ctx context.Context, id uuid.UUID, req downstream.PatchIdentityDocsRequest) (*downstream.DetailResponse, error) {
	if err := domain.ValidateIdentityDocs(req.DocumentType, req.DocumentNumber, req.IssuingCountry, req.FrontDocumentURL, req.SelfieURL); err != nil {
		return nil, err
	}
	detail, err := s.downstream.PatchIdentityDocs(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) PatchFinancialInfo(ctx context.Context, id uuid.UUID, req downstream.PatchFinancialRequest) (*downstream.DetailResponse, error) {
	if err := domain.ValidateFinancial(req.Occupation, req.AnnualIncome, req.SourceOfFunds); err != nil {
		return nil, err
	}
	detail, err := s.downstream.PatchFinancialInfo(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*downstream.DetailResponse, error) {
	if err := domain.ValidateWalletTier(tier); err != nil {
		return nil, err
	}
	detail, err := s.downstream.PatchWalletTier(ctx, id, tier)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) ResubmitDocuments(ctx context.Context, id uuid.UUID, req downstream.ResubmitDocsRequest) error {
	return s.downstream.ResubmitDocuments(ctx, id, req)
}

func (s *clientService) GetAuditLogs(ctx context.Context, id uuid.UUID) ([]downstream.AuditLog, error) {
	return s.downstream.GetAuditLogs(ctx, id)
}

func (s *clientService) GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*downstream.WorkflowProgressResponse, error) {
	return s.downstream.GetWorkflowProgress(ctx, id)
}

func (s *clientService) PatchDraft(ctx context.Context, id uuid.UUID, body interface{}) (*downstream.DetailResponse, error) {
	detail, err := s.downstream.PatchDraft(ctx, id, body)
	if err != nil {
		return nil, err
	}
	return cleanDetailResponse(detail), nil
}

func (s *clientService) HealthCheck(ctx context.Context) (bool, error) {
	return s.downstream.HealthCheck(ctx)
}
