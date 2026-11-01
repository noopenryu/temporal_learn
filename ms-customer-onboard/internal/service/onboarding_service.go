package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"ms-customer-onboard/internal/config"
	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/observability"
	"ms-customer-onboard/internal/temporal/workflows"
)

// PatchDraftRequest allows updating multiple sections in a single request
type PatchDraftRequest struct {
	PersonalInfo        *domain.PersonalInfo      `json:"personal_info,omitempty"`
	AddressInfo         *domain.AddressInfo       `json:"address_info,omitempty"`
	IdentityDocs        *domain.IdentityDocuments `json:"identity_docs,omitempty"`
	FinancialInfo       *domain.FinancialInfo     `json:"financial_info,omitempty"`
	WalletTierRequested *string                   `json:"wallet_tier_requested,omitempty"`
}

// OnboardingService defines all business operations for customer onboarding
type OnboardingService interface {
	StartSession(ctx context.Context, customerID, email, phone string) (*domain.OnboardingApplication, bool, error)
	GetApplication(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error)
	GetActiveByCustomerID(ctx context.Context, customerID string) (*domain.OnboardingApplication, error)
	GetActiveByPhoneNumber(ctx context.Context, phone string) (*domain.OnboardingApplication, error)
	ResumeSession(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error)
	PatchPersonalInfo(ctx context.Context, id uuid.UUID, info domain.PersonalInfo) (*domain.OnboardingApplication, error)
	PatchAddress(ctx context.Context, id uuid.UUID, addr domain.AddressInfo) (*domain.OnboardingApplication, error)
	PatchIdentityDocs(ctx context.Context, id uuid.UUID, docs domain.IdentityDocuments) (*domain.OnboardingApplication, error)
	PatchFinancialInfo(ctx context.Context, id uuid.UUID, fin domain.FinancialInfo) (*domain.OnboardingApplication, error)
	PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*domain.OnboardingApplication, error)
	PatchDraft(ctx context.Context, id uuid.UUID, req PatchDraftRequest) (*domain.OnboardingApplication, error)
	SubmitApplication(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error)
	CancelSession(ctx context.Context, id uuid.UUID, reason string) error
	ResubmitDocuments(ctx context.Context, id uuid.UUID, docs domain.IdentityDocuments) error
	SubmitManualReviewDecision(ctx context.Context, id uuid.UUID, decision workflows.ManualReviewDecision) error
	ListApplications(ctx context.Context, filter domain.ListFilter) ([]*domain.OnboardingApplication, int, error)
	PutConfirmedData(ctx context.Context, id uuid.UUID, confirmed domain.ConfirmedData) (*domain.OnboardingApplication, error)
	GetAuditLogs(ctx context.Context, id uuid.UUID) ([]domain.AuditLog, error)
	GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*workflows.OnboardingProgressResponse, error)
}

type onboardingServiceImpl struct {
	cfg            *config.Config
	repo           domain.OnboardingRepository
	temporalClient client.Client
}

func NewOnboardingService(
	cfg *config.Config,
	repo domain.OnboardingRepository,
	temporalClient client.Client,
) OnboardingService {
	return &onboardingServiceImpl{
		cfg:            cfg,
		repo:           repo,
		temporalClient: temporalClient,
	}
}

// StartSession creates a new draft onboarding session, or resumes an active session if one already exists
func (s *onboardingServiceImpl) StartSession(ctx context.Context, customerID, email, phone string) (*domain.OnboardingApplication, bool, error) {
	phone = domain.NormalizePhoneNumber(phone)
	if err := domain.ValidateStartRequest(customerID, email, phone); err != nil {
		return nil, false, fmt.Errorf("invalid start request: %w", err)
	}

	// 1. Check if an application already exists by phone number
	existingByPhone, err := s.repo.GetActiveByPhoneNumber(ctx, phone)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check existing application by phone: %w", err)
	}
	if existingByPhone != nil {
		// Existing fully completed customer
		if existingByPhone.Status == domain.StatusApproved {
			return nil, false, domain.ErrAlreadyOnboarded
		}

		// Check if active draft has expired
		if existingByPhone.IsExpired() {
			existingByPhone.Status = domain.StatusExpired
			_ = s.repo.Update(ctx, existingByPhone)
			_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
				ApplicationID: existingByPhone.ID,
				Action:        "SESSION_EXPIRED",
				FromStatus:    string(domain.StatusDraft),
				ToStatus:      string(domain.StatusExpired),
				CreatedAt:     time.Now(),
			})
			// Proceed below to create a fresh draft session
		} else {
			// Resume in-progress session
			slog.Info("Resuming existing active onboarding session by phone", "phone", phone, "app_id", existingByPhone.ID)
			_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
				ApplicationID: existingByPhone.ID,
				Action:        "SESSION_RESUMED_BY_PHONE",
				FromStatus:    string(existingByPhone.Status),
				ToStatus:      string(existingByPhone.Status),
				Details: map[string]any{
					"phone":        phone,
					"current_step": existingByPhone.CurrentStep,
					"progress":     existingByPhone.ProgressPercentage(),
				},
				CreatedAt: time.Now(),
			})
			return existingByPhone, true, nil
		}
	}

	// 2. If customerID was provided, check if active draft exists for this customer ID (backwards compatibility)
	if strings.TrimSpace(customerID) != "" {
		existing, err := s.repo.GetActiveByCustomerID(ctx, customerID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to check existing application: %w", err)
		}
		if existing != nil {
			if !existing.IsExpired() {
				slog.Info("Returning existing active onboarding session", "customer_id", customerID, "app_id", existing.ID)
				return existing, true, nil
			}
			existing.Status = domain.StatusExpired
			_ = s.repo.Update(ctx, existing)
		}
	} else {
		// Auto-generate customer ID if not provided by client
		customerID = domain.GenerateCustomerID()
	}

	now := time.Now()
	ttlDays := s.cfg.DraftTTLDays
	if ttlDays <= 0 {
		ttlDays = 14
	}

	app := &domain.OnboardingApplication{
		ID:                  uuid.New(),
		CustomerID:          customerID,
		Status:              domain.StatusDraft,
		CurrentStep:         domain.StepPersonalInfo,
		CompletedSteps:      []string{},
		PersonalInfo:        domain.PersonalInfo{Email: email, PhoneNumber: phone},
		WalletTierRequested: "TIER_1_BASIC",
		ExpiresAt:           now.AddDate(0, 0, ttlDays),
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	app.UpdateStepProgress()

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, false, fmt.Errorf("failed to persist new onboarding application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "SESSION_STARTED",
		ToStatus:      string(domain.StatusDraft),
		Details: map[string]any{
			"customer_id": customerID,
			"phone":       phone,
			"expires_at":  app.ExpiresAt,
		},
		CreatedAt: now,
	})

	slog.Info("Started new onboarding session", "app_id", app.ID, "customer_id", customerID, "phone", phone)
	return app, false, nil
}

// GetApplication retrieves an application by ID and checks for expiration
func (s *onboardingServiceImpl) GetApplication(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, nil
	}

	// Check if draft has expired
	if app.IsExpired() {
		app.Status = domain.StatusExpired
		_ = s.repo.Update(ctx, app)
		_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
			ApplicationID: app.ID,
			Action:        "SESSION_EXPIRED",
			FromStatus:    string(domain.StatusDraft),
			ToStatus:      string(domain.StatusExpired),
			CreatedAt:     time.Now(),
		})
	}

	return app, nil
}

// GetActiveByCustomerID returns the active application for a customer
func (s *onboardingServiceImpl) GetActiveByCustomerID(ctx context.Context, customerID string) (*domain.OnboardingApplication, error) {
	if strings.TrimSpace(customerID) == "" {
		return nil, errors.New("customer_id is required")
	}
	app, err := s.repo.GetActiveByCustomerID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active application: %w", err)
	}
	if app == nil {
		return nil, nil
	}
	if app.IsExpired() {
		app.Status = domain.StatusExpired
		_ = s.repo.Update(ctx, app)
		return nil, nil
	}
	return app, nil
}

// GetActiveByPhoneNumber returns the active application for a phone number
func (s *onboardingServiceImpl) GetActiveByPhoneNumber(ctx context.Context, phone string) (*domain.OnboardingApplication, error) {
	normalized := domain.NormalizePhoneNumber(phone)
	if normalized == "" {
		return nil, errors.New("phone_number is required")
	}
	app, err := s.repo.GetActiveByPhoneNumber(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("failed to query active application by phone: %w", err)
	}
	if app == nil {
		return nil, nil
	}
	if app.IsExpired() {
		app.Status = domain.StatusExpired
		_ = s.repo.Update(ctx, app)
		return nil, nil
	}
	return app, nil
}

// ResumeSession verifies that an application exists, has not expired, and is resume-ready
func (s *onboardingServiceImpl) ResumeSession(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, err := s.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("onboarding application not found")
	}
	if app.Status == domain.StatusExpired {
		return nil, errors.New("onboarding session has expired; please start a new session")
	}
	if app.Status == domain.StatusCancelled {
		return nil, errors.New("onboarding session has been cancelled")
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "SESSION_RESUMED",
		FromStatus:    string(app.Status),
		ToStatus:      string(app.Status),
		Details: map[string]any{
			"current_step": app.CurrentStep,
			"progress":     app.ProgressPercentage(),
		},
		CreatedAt: time.Now(),
	})

	return app, nil
}

// PatchPersonalInfo updates personal information in the draft
func (s *onboardingServiceImpl) PatchPersonalInfo(ctx context.Context, id uuid.UUID, info domain.PersonalInfo) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidatePersonalInfo(&info); err != nil {
		return nil, fmt.Errorf("invalid personal info: %w", err)
	}

	app.PersonalInfo = info
	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_PERSONAL_INFO",
		Details: map[string]any{
			"step":     domain.StepPersonalInfo,
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// PatchAddress updates address details in the draft
func (s *onboardingServiceImpl) PatchAddress(ctx context.Context, id uuid.UUID, addr domain.AddressInfo) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidateAddressInfo(&addr); err != nil {
		return nil, fmt.Errorf("invalid address info: %w", err)
	}

	app.AddressInfo = addr
	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_ADDRESS",
		Details: map[string]any{
			"step":     domain.StepAddressInfo,
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// PatchIdentityDocs updates identity documents in the draft
func (s *onboardingServiceImpl) PatchIdentityDocs(ctx context.Context, id uuid.UUID, docs domain.IdentityDocuments) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidateIdentityDocs(&docs); err != nil {
		return nil, fmt.Errorf("invalid identity documents: %w", err)
	}

	app.IdentityDocs = docs
	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_IDENTITY_DOCS",
		Details: map[string]any{
			"step":     domain.StepIdentityDocs,
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// PatchFinancialInfo updates economic details in the draft
func (s *onboardingServiceImpl) PatchFinancialInfo(ctx context.Context, id uuid.UUID, fin domain.FinancialInfo) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidateFinancialInfo(&fin, app.WalletTierRequested); err != nil {
		return nil, fmt.Errorf("invalid financial info: %w", err)
	}

	app.FinancialInfo = fin
	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_FINANCIAL_INFO",
		Details: map[string]any{
			"step":     domain.StepFinancial,
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// PatchWalletTier updates the requested wallet tier
func (s *onboardingServiceImpl) PatchWalletTier(ctx context.Context, id uuid.UUID, tier string) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := domain.ValidateWalletTier(tier); err != nil {
		return nil, fmt.Errorf("invalid wallet tier: %w", err)
	}

	app.WalletTierRequested = strings.ToUpper(strings.TrimSpace(tier))
	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_WALLET_TIER",
		Details: map[string]any{
			"tier":     app.WalletTierRequested,
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// PatchDraft allows patching any combination of sections in a single call
func (s *onboardingServiceImpl) PatchDraft(ctx context.Context, id uuid.UUID, req PatchDraftRequest) (*domain.OnboardingApplication, error) {
	app, err := s.getModifiableApplication(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.PersonalInfo != nil {
		if err := domain.ValidatePersonalInfo(req.PersonalInfo); err != nil {
			return nil, fmt.Errorf("invalid personal info: %w", err)
		}
		app.PersonalInfo = *req.PersonalInfo
	}

	if req.AddressInfo != nil {
		if err := domain.ValidateAddressInfo(req.AddressInfo); err != nil {
			return nil, fmt.Errorf("invalid address info: %w", err)
		}
		app.AddressInfo = *req.AddressInfo
	}

	if req.IdentityDocs != nil {
		if err := domain.ValidateIdentityDocs(req.IdentityDocs); err != nil {
			return nil, fmt.Errorf("invalid identity docs: %w", err)
		}
		app.IdentityDocs = *req.IdentityDocs
	}

	if req.WalletTierRequested != nil {
		if err := domain.ValidateWalletTier(*req.WalletTierRequested); err != nil {
			return nil, fmt.Errorf("invalid wallet tier: %w", err)
		}
		app.WalletTierRequested = strings.ToUpper(strings.TrimSpace(*req.WalletTierRequested))
	}

	if req.FinancialInfo != nil {
		if err := domain.ValidateFinancialInfo(req.FinancialInfo, app.WalletTierRequested); err != nil {
			return nil, fmt.Errorf("invalid financial info: %w", err)
		}
		app.FinancialInfo = *req.FinancialInfo
	}

	app.UpdateStepProgress()

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "PATCH_DRAFT_BATCH",
		Details: map[string]any{
			"progress": app.ProgressPercentage(),
		},
	})

	return app, nil
}

// SubmitApplication validates readiness, transitions status to SUBMITTED, and starts the Temporal workflow
func (s *onboardingServiceImpl) SubmitApplication(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve application: %w", err)
	}
	if app == nil {
		return nil, errors.New("onboarding application not found")
	}

	canSubmit, uncompleted := app.CanSubmit()
	if !canSubmit {
		return nil, fmt.Errorf("cannot submit incomplete application: %s", strings.Join(uncompleted, "; "))
	}

	now := time.Now()
	app.Status = domain.StatusSubmitted
	app.SubmittedAt = &now

	// Start Temporal Workflow
	workflowID := fmt.Sprintf("customer-onboard-%s", app.ID.String())
	app.WorkflowID = workflowID

	if s.temporalClient != nil {
		workflowOptions := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: s.cfg.TemporalTaskQueue,
		}

		tc, _ := observability.FromContext(ctx)
		workflowInput := workflows.OnboardingWorkflowInput{
			TraceID:              tc.TraceID,
			SpanID:               tc.SpanID,
			ApplicationID:        app.ID.String(),
			CustomerID:           app.CustomerID,
			PersonalInfo:         app.PersonalInfo,
			AddressInfo:          app.AddressInfo,
			IdentityDocs:         app.IdentityDocs,
			FinancialInfo:        app.FinancialInfo,
			WalletTierRequested:  app.WalletTierRequested,
			EligibilityTaskQueue: s.cfg.EligibilityTaskQueue,
			KycTaskQueue:         s.cfg.KycTaskQueue,
		}

		run, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflows.CustomerOnboardingWorkflow, workflowInput)
		if err != nil {
			slog.Error("Failed to trigger Temporal onboarding workflow", "error", err, "workflow_id", workflowID)
			return nil, fmt.Errorf("failed to initiate onboarding workflow: %w", err)
		}

		app.RunID = run.GetRunID()
		app.Status = domain.StatusProcessing
		slog.Info("Started Temporal onboarding workflow", "workflow_id", workflowID, "run_id", app.RunID)
	}

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application after submission: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "APPLICATION_SUBMITTED",
		FromStatus:    string(domain.StatusDraft),
		ToStatus:      string(app.Status),
		Details: map[string]any{
			"workflow_id": workflowID,
			"run_id":      app.RunID,
		},
		CreatedAt: now,
	})

	return app, nil
}

// CancelSession cancels an active onboarding application
func (s *onboardingServiceImpl) CancelSession(ctx context.Context, id uuid.UUID, reason string) error {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return errors.New("onboarding application not found")
	}

	if app.Status == domain.StatusApproved || app.Status == domain.StatusRejected {
		return fmt.Errorf("cannot cancel application with terminal status %s", app.Status)
	}

	oldStatus := string(app.Status)
	app.Status = domain.StatusCancelled
	app.RejectionReasons = append(app.RejectionReasons, "Cancelled by user: "+reason)

	if err := s.repo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to cancel application: %w", err)
	}

	// If workflow was running, send cancel signal
	if s.temporalClient != nil && app.WorkflowID != "" {
		_ = s.temporalClient.SignalWorkflow(ctx, app.WorkflowID, app.RunID, workflows.SignalCancelOnboarding, workflows.CancelOnboardingSignal{
			Reason:      reason,
			CancelledAt: time.Now(),
		})
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "APPLICATION_CANCELLED",
		FromStatus:    oldStatus,
		ToStatus:      string(domain.StatusCancelled),
		Details: map[string]any{
			"reason": reason,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

// ResubmitDocuments signals the running workflow with refreshed documents
func (s *onboardingServiceImpl) ResubmitDocuments(ctx context.Context, id uuid.UUID, docs domain.IdentityDocuments) error {
	if err := domain.ValidateIdentityDocs(&docs); err != nil {
		return fmt.Errorf("invalid identity documents: %w", err)
	}

	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return errors.New("onboarding application not found")
	}

	if app.Status != domain.StatusWaitingUpdatedDocs && app.Status != domain.StatusDraft {
		return fmt.Errorf("cannot resubmit documents in status %s", app.Status)
	}

	app.IdentityDocs = docs
	if err := s.repo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update application documents: %w", err)
	}

	if s.temporalClient != nil && app.WorkflowID != "" {
		err := s.temporalClient.SignalWorkflow(ctx, app.WorkflowID, app.RunID, workflows.SignalUpdateDocuments, workflows.UpdateDocumentsSignal{
			IdentityDocs: docs,
			SubmittedAt:  time.Now(),
		})
		if err != nil {
			return fmt.Errorf("failed to signal workflow with updated documents: %w", err)
		}
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "DOCUMENTS_RESUBMITTED",
		Details: map[string]any{
			"doc_type":   docs.DocumentType,
			"doc_number": docs.DocumentNumber,
		},
	})

	return nil
}

// SubmitManualReviewDecision signals the running workflow with compliance review decision
func (s *onboardingServiceImpl) SubmitManualReviewDecision(ctx context.Context, id uuid.UUID, decision workflows.ManualReviewDecision) error {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return errors.New("onboarding application not found")
	}

	if app.Status != domain.StatusRequiresManualReview {
		return fmt.Errorf("application is not pending manual review (current status: %s)", app.Status)
	}

	decision.DecisionTime = time.Now()

	// If confirmed data or confirmed tier provided, attach and save
	if decision.ConfirmedData != nil {
		if decision.ConfirmedData.ConfirmedAt.IsZero() {
			decision.ConfirmedData.ConfirmedAt = decision.DecisionTime
		}
		if decision.ConfirmedData.ReviewerID == "" {
			decision.ConfirmedData.ReviewerID = decision.ReviewerID
		}
		app.ConfirmedData = decision.ConfirmedData
	}
	if decision.ConfirmedTier != "" {
		app.ApprovedTier = decision.ConfirmedTier
	} else if decision.ConfirmedData != nil && decision.ConfirmedData.ApprovedTier != "" {
		decision.ConfirmedTier = decision.ConfirmedData.ApprovedTier
		app.ApprovedTier = decision.ConfirmedData.ApprovedTier
	}

	if err := s.repo.Update(ctx, app); err != nil {
		return fmt.Errorf("failed to update application with confirmed review data: %w", err)
	}

	if s.temporalClient != nil && app.WorkflowID != "" {
		err := s.temporalClient.SignalWorkflow(ctx, app.WorkflowID, app.RunID, workflows.SignalManualReviewDecision, decision)
		if err != nil {
			return fmt.Errorf("failed to signal workflow with manual review decision: %w", err)
		}
	}

	auditDetails := map[string]any{
		"reviewer_id": decision.ReviewerID,
		"approved":    decision.Approved,
		"notes":       decision.Notes,
	}
	if decision.ConfirmedTier != "" {
		auditDetails["confirmed_tier"] = decision.ConfirmedTier
	}
	if decision.ConfirmedData != nil {
		auditDetails["has_confirmed_data"] = true
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "MANUAL_REVIEW_DECISION_RECORDED",
		Details:       auditDetails,
		CreatedAt:     time.Now(),
	})

	return nil
}

// ListApplications queries applications using repository with optional filters
func (s *onboardingServiceImpl) ListApplications(ctx context.Context, filter domain.ListFilter) ([]*domain.OnboardingApplication, int, error) {
	if s.repo == nil {
		return nil, 0, errors.New("repository not configured")
	}
	return s.repo.List(ctx, filter)
}

// PutConfirmedData records staff confirmed verification data on the application
func (s *onboardingServiceImpl) PutConfirmedData(ctx context.Context, id uuid.UUID, confirmed domain.ConfirmedData) (*domain.OnboardingApplication, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, errors.New("onboarding application not found")
	}

	if confirmed.ConfirmedAt.IsZero() {
		confirmed.ConfirmedAt = time.Now()
	}

	app.ConfirmedData = &confirmed
	if confirmed.ApprovedTier != "" {
		app.ApprovedTier = confirmed.ApprovedTier
	}

	if err := s.repo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to save confirmed data: %w", err)
	}

	_ = s.repo.SaveAuditLog(ctx, &domain.AuditLog{
		ApplicationID: app.ID,
		Action:        "STAFF_DATA_CONFIRMED",
		FromStatus:    string(app.Status),
		ToStatus:      string(app.Status),
		Details: map[string]any{
			"reviewer_id":   confirmed.ReviewerID,
			"approved_tier": confirmed.ApprovedTier,
			"notes":         confirmed.Notes,
		},
		CreatedAt: time.Now(),
	})

	return app, nil
}

// GetAuditLogs retrieves history of transitions for an application
func (s *onboardingServiceImpl) GetAuditLogs(ctx context.Context, id uuid.UUID) ([]domain.AuditLog, error) {
	return s.repo.GetAuditLogs(ctx, id)
}

// GetWorkflowProgress queries the live Temporal workflow progress if active, or infers it from persisted state
func (s *onboardingServiceImpl) GetWorkflowProgress(ctx context.Context, id uuid.UUID) (*workflows.OnboardingProgressResponse, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, errors.New("onboarding application not found")
	}

	// 1. If workflow is running and temporalClient is available, query Temporal directly
	if s.temporalClient != nil && app.WorkflowID != "" {
		queryVal, err := s.temporalClient.QueryWorkflow(ctx, app.WorkflowID, app.RunID, workflows.QueryOnboardingProgress)
		if err == nil {
			var resp workflows.OnboardingProgressResponse
			if err := queryVal.Get(&resp); err == nil {
				return &resp, nil
			}
		}
		// If Query failed because workflow has closed or completed, fall through to status-based synthesis
	}

	// 2. Synthesize progress from DB status
	resp := &workflows.OnboardingProgressResponse{
		LastUpdated: app.UpdatedAt,
	}

	switch app.Status {
	case domain.StatusDraft:
		resp.Stage = "DRAFT"
		resp.PercentComplete = 10
		resp.StatusMessage = "Application draft in progress"
	case domain.StatusSubmitted:
		resp.Stage = "SUBMITTED"
		resp.PercentComplete = 20
		resp.StatusMessage = "Application submitted, initiating automated verification saga"
	case domain.StatusProcessing:
		resp.Stage = "PROCESSING"
		resp.PercentComplete = 60
		resp.StatusMessage = "Automated KYC and eligibility criteria evaluation in progress"
	case domain.StatusRequiresManualReview:
		resp.Stage = "REQUIRES_MANUAL_REVIEW"
		resp.PercentComplete = 75
		resp.StatusMessage = "Application queued for compliance officer manual inspection"
	case domain.StatusWaitingUpdatedDocs:
		resp.Stage = "WAITING_UPDATED_DOCS"
		resp.PercentComplete = 50
		resp.StatusMessage = "Awaiting updated identity documentation from applicant"
	case domain.StatusApproved:
		resp.Stage = "APPROVED"
		resp.PercentComplete = 100
		resp.StatusMessage = "Application approved and internet wallet ledger created"
	case domain.StatusRejected:
		resp.Stage = "REJECTED"
		resp.PercentComplete = 100
		resp.StatusMessage = "Application did not meet compliance criteria"
	case domain.StatusExpired:
		resp.Stage = "EXPIRED"
		resp.PercentComplete = 100
		resp.StatusMessage = "Application session expired"
	case domain.StatusCancelled:
		resp.Stage = "CANCELLED"
		resp.PercentComplete = 100
		resp.StatusMessage = "Application was cancelled"
	default:
		resp.Stage = string(app.Status)
		resp.PercentComplete = 0
		resp.StatusMessage = "Application in unknown state"
	}

	return resp, nil
}

func (s *onboardingServiceImpl) getModifiableApplication(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	app, err := s.GetApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("onboarding application not found")
	}
	if !app.CanBeModified() {
		return nil, fmt.Errorf("application in status %s cannot be modified", app.Status)
	}
	return app, nil
}
