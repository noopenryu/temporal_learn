package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/temporal/activities"
)

// EligibilityCheckRequest matches the payload expected by ms-eligiblity's EvaluateEligibilityActivity
type EligibilityCheckRequest struct {
	TraceID             string    `json:"trace_id,omitempty"`
	CustomerID          string    `json:"customer_id"`
	NationalID          string    `json:"national_id"`
	IDType              string    `json:"id_type"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	DateOfBirth         time.Time `json:"date_of_birth"`
	Nationality         string    `json:"nationality"`
	ResidenceCountry    string    `json:"residence_country"`
	Email               string    `json:"email"`
	PhoneNumber         string    `json:"phone_number"`
	DeclaredPEP         bool      `json:"declared_pep"`
	AnnualIncome        float64   `json:"annual_income,omitempty"`
	WalletTierRequested string    `json:"wallet_tier_requested"`
}

// EligibilityCheckResult matches the response from ms-eligiblity
type EligibilityCheckResult struct {
	ID               string    `json:"id"`
	CustomerID       string    `json:"customer_id"`
	Status           string    `json:"status"` // APPROVED, REJECTED, REQUIRES_MANUAL_REVIEW
	RiskScore        int       `json:"risk_score"`
	RiskLevel        string    `json:"risk_level"`
	ApprovedTier     string    `json:"approved_tier"`
	RejectionReasons []string  `json:"rejection_reasons"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
}

// CustomerOnboardingWorkflow coordinates the end-to-end customer onboarding lifecycle
func CustomerOnboardingWorkflow(ctx workflow.Context, input OnboardingWorkflowInput) (*OnboardingWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CustomerOnboardingWorkflow started", "customer_id", input.CustomerID, "app_id", input.ApplicationID)

	// Progress state tracking for queries
	progress := OnboardingProgressResponse{
		Stage:           "INITIALIZED",
		PercentComplete: 10,
		StatusMessage:   "Onboarding workflow initialized",
		LastUpdated:     workflow.Now(ctx),
	}

	// Register Query handler so Frontend can query workflow progress in real-time
	err := workflow.SetQueryHandler(ctx, QueryOnboardingProgress, func() (OnboardingProgressResponse, error) {
		return progress, nil
	})
	if err != nil {
		logger.Error("Failed to register query handler", "error", err)
	}

	// Activity options for internal activities (status sync, wallet provisioning, notifications)
	localActivityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
		},
	}
	localCtx := workflow.WithActivityOptions(ctx, localActivityOpts)

	// Activity options for eligibility evaluation (dispatched to ms-eligiblity task queue)
	eligQueue := input.EligibilityTaskQueue
	if eligQueue == "" {
		eligQueue = "CUSTOMER_ONBOARDING_TASK_QUEUE"
	}
	eligibilityOpts := workflow.ActivityOptions{
		TaskQueue:           eligQueue,
		StartToCloseTimeout: 45 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"ERR_ELIGIBILITY_EVALUATION", "ERR_INVALID_CHECK_ID"},
		},
	}
	eligibilityCtx := workflow.WithActivityOptions(ctx, eligibilityOpts)

	// Setup Signal channels
	manualReviewChan := workflow.GetSignalChannel(ctx, SignalManualReviewDecision)
	updateDocsChan := workflow.GetSignalChannel(ctx, SignalUpdateDocuments)
	cancelChan := workflow.GetSignalChannel(ctx, SignalCancelOnboarding)

	var lastReviewDecision *ManualReviewDecision
	var acts *activities.OnboardingActivities

	// Milestone 1: Sync status to PROCESSING
	progress.Stage = "PROCESSING"
	progress.PercentComplete = 20
	progress.StatusMessage = "Evaluating eligibility and risk criteria"
	progress.LastUpdated = workflow.Now(ctx)

	err = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
		ApplicationID: input.ApplicationID,
		Status:        domain.StatusProcessing,
		CurrentStep:   "eligibility_check",
	}).Get(localCtx, nil)
	if err != nil {
		logger.Error("Failed to update status to PROCESSING", "error", err)
	}

	// Milestone 2: Eligibility Check via ms-eligiblity
	eligibilityReq := EligibilityCheckRequest{
		TraceID:             input.TraceID,
		CustomerID:          input.CustomerID,
		NationalID:          input.IdentityDocs.DocumentNumber,
		IDType:              input.IdentityDocs.DocumentType,
		FirstName:           input.PersonalInfo.FirstName,
		LastName:            input.PersonalInfo.LastName,
		DateOfBirth:         input.PersonalInfo.DateOfBirth,
		Nationality:         input.PersonalInfo.Nationality,
		ResidenceCountry:    input.AddressInfo.Country,
		Email:               input.PersonalInfo.Email,
		PhoneNumber:         input.PersonalInfo.PhoneNumber,
		DeclaredPEP:         input.FinancialInfo.DeclaredPEP,
		AnnualIncome:        input.FinancialInfo.AnnualIncome,
		WalletTierRequested: input.WalletTierRequested,
	}

	var eligResult EligibilityCheckResult
	err = workflow.ExecuteActivity(eligibilityCtx, "EvaluateEligibilityActivity", eligibilityReq).Get(eligibilityCtx, &eligResult)
	if err != nil {
		logger.Error("EvaluateEligibilityActivity failed", "error", err)
		return nil, fmt.Errorf("eligibility activity failed: %w", err)
	}

	logger.Info("Eligibility evaluation result", "status", eligResult.Status, "risk_score", eligResult.RiskScore)

	// If Eligibility rejected: terminate with REJECTED
	if eligResult.Status == "REJECTED" {
		progress.Stage = "REJECTED"
		progress.PercentComplete = 100
		progress.StatusMessage = "Application rejected during eligibility evaluation"
		progress.LastUpdated = workflow.Now(ctx)

		_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
			ApplicationID:    input.ApplicationID,
			Status:           domain.StatusRejected,
			RejectionReasons: eligResult.RejectionReasons,
		}).Get(localCtx, nil)

		_ = workflow.ExecuteActivity(localCtx, acts.SendOnboardingNotificationActivity, activities.NotificationInput{
			CustomerID:      input.CustomerID,
			Email:           input.PersonalInfo.Email,
			PhoneNumber:     input.PersonalInfo.PhoneNumber,
			FullName:        input.PersonalInfo.FullName(),
			Status:          domain.StatusRejected,
			RejectionReason: fmt.Sprintf("%v", eligResult.RejectionReasons),
		}).Get(localCtx, nil)

		return &OnboardingWorkflowResult{
			ApplicationID:    input.ApplicationID,
			CustomerID:       input.CustomerID,
			FinalStatus:      domain.StatusRejected,
			RejectionReasons: eligResult.RejectionReasons,
			CompletedAt:      workflow.Now(ctx),
		}, nil
	}

	// If Eligibility requires manual review: wait for manual-review-decision signal or timeout (7 days)
	if eligResult.Status == "REQUIRES_MANUAL_REVIEW" {
		progress.Stage = "REQUIRES_MANUAL_REVIEW"
		progress.PercentComplete = 35
		progress.StatusMessage = "Pending compliance manual review"
		progress.LastUpdated = workflow.Now(ctx)

		_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
			ApplicationID: input.ApplicationID,
			Status:        domain.StatusRequiresManualReview,
			CurrentStep:   "compliance_manual_review",
		}).Get(localCtx, nil)

		// Wait for compliance review signal or cancellation or SLA timeout (7 days)
		reviewSelector := workflow.NewSelector(ctx)
		var decision ManualReviewDecision
		var cancelled CancelOnboardingSignal
		var timeoutFired bool

		reviewSelector.AddReceive(manualReviewChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &decision)
		})
		reviewSelector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &cancelled)
		})
		reviewSelector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), func(f workflow.Future) {
			timeoutFired = true
		})

		reviewSelector.Select(ctx)

		if cancelled.Reason != "" {
			return handleCancellation(ctx, localCtx, acts, input, cancelled.Reason)
		}
		if timeoutFired || !decision.Approved {
			reasons := []string{"Compliance review timed out or rejected"}
			if decision.Notes != "" {
				reasons = append(reasons, decision.Notes)
			}
			return handleRejection(ctx, localCtx, acts, input, reasons)
		}

		decCopy := decision
		lastReviewDecision = &decCopy
		details := map[string]any{
			"reviewer_id": decision.ReviewerID,
			"notes":       decision.Notes,
		}
		if decision.ConfirmedTier != "" {
			details["confirmed_tier"] = decision.ConfirmedTier
		}
		_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
			ApplicationID: input.ApplicationID,
			Status:        domain.StatusProcessing,
			CurrentStep:   "manual_review_approved",
			ApprovedTier:  decision.ConfirmedTier,
			Details:       details,
		}).Get(localCtx, nil)
	}

	// Milestone 3: KYC Document & Identity Verification Loop
	progress.Stage = "KYC_VERIFICATION"
	progress.PercentComplete = 50
	progress.StatusMessage = "Verifying identity documents and biometrics"
	progress.LastUpdated = workflow.Now(ctx)

	kycApproved := false
	currentDocs := input.IdentityDocs

	for !kycApproved {
		kycInput := activities.KycVerificationInput{
			ApplicationID:  input.ApplicationID,
			CustomerID:     input.CustomerID,
			FullName:       input.PersonalInfo.FullName(),
			DateOfBirth:    input.PersonalInfo.DateOfBirth.Format("2006-01-02"),
			Nationality:    input.PersonalInfo.Nationality,
			PhoneNumber:    input.PersonalInfo.PhoneNumber,
			Email:          input.PersonalInfo.Email,
			DocumentType:   currentDocs.DocumentType,
			DocumentNumber: currentDocs.DocumentNumber,
			FrontDocURL:    currentDocs.FrontDocumentURL,
			BackDocURL:     currentDocs.BackDocumentURL,
			SelfieURL:      currentDocs.SelfieURL,
		}

		var kycRes activities.KycVerificationResult
		err = workflow.ExecuteActivity(localCtx, acts.ExecuteKycVerificationActivity, kycInput).Get(localCtx, &kycRes)
		if err != nil {
			logger.Error("ExecuteKycVerificationActivity failed", "error", err)
			return nil, fmt.Errorf("kyc verification failed: %w", err)
		}

		if kycRes.Passed {
			kycApproved = true
			break
		}

		// If document was blurry/unreadable: pause and wait for update-documents signal from customer!
		if kycRes.RequiresDocumentResub {
			progress.Stage = "WAITING_UPDATED_DOCS"
			progress.PercentComplete = 55
			progress.StatusMessage = kycRes.RejectionReason
			progress.LastUpdated = workflow.Now(ctx)

			_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
				ApplicationID: input.ApplicationID,
				Status:        domain.StatusWaitingUpdatedDocs,
				CurrentStep:   "resubmit_identity_documents",
			}).Get(localCtx, nil)

			docSelector := workflow.NewSelector(ctx)
			var updatedDocs UpdateDocumentsSignal
			var cancelled CancelOnboardingSignal
			var expired bool

			docSelector.AddReceive(updateDocsChan, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &updatedDocs)
			})
			docSelector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &cancelled)
			})
			docSelector.AddFuture(workflow.NewTimer(ctx, 14*24*time.Hour), func(f workflow.Future) {
				expired = true
			})

			docSelector.Select(ctx)

			if cancelled.Reason != "" {
				return handleCancellation(ctx, localCtx, acts, input, cancelled.Reason)
			}
			if expired {
				return handleRejection(ctx, localCtx, acts, input, []string{"Document resubmission period expired"})
			}

			// Refreshed docs received! Re-run KYC verification loop
			currentDocs = updatedDocs.IdentityDocs
			continue
		}

		// If KYC requires manual compliance review
		if kycRes.RequiresManualReview {
			progress.Stage = "REQUIRES_MANUAL_REVIEW"
			progress.PercentComplete = 60
			progress.StatusMessage = "Pending compliance manual review for identity documents"
			progress.LastUpdated = workflow.Now(ctx)

			_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
				ApplicationID: input.ApplicationID,
				Status:        domain.StatusRequiresManualReview,
				CurrentStep:   "kyc_manual_review",
			}).Get(localCtx, nil)

			reviewSelector := workflow.NewSelector(ctx)
			var decision ManualReviewDecision
			var cancelled CancelOnboardingSignal
			var timeoutFired bool

			reviewSelector.AddReceive(manualReviewChan, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &decision)
			})
			reviewSelector.AddReceive(cancelChan, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &cancelled)
			})
			reviewSelector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), func(f workflow.Future) {
				timeoutFired = true
			})

			reviewSelector.Select(ctx)

			if cancelled.Reason != "" {
				return handleCancellation(ctx, localCtx, acts, input, cancelled.Reason)
			}
			if timeoutFired || !decision.Approved {
				reasons := []string{"Compliance manual review rejected for KYC"}
				if decision.Notes != "" {
					reasons = append(reasons, decision.Notes)
				}
				return handleRejection(ctx, localCtx, acts, input, reasons)
			}

			decCopy := decision
			lastReviewDecision = &decCopy
			kycApprovedDetails := map[string]any{
				"reviewer_id": decision.ReviewerID,
				"notes":       decision.Notes,
			}
			if decision.ConfirmedTier != "" {
				kycApprovedDetails["confirmed_tier"] = decision.ConfirmedTier
			}
			_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
				ApplicationID: input.ApplicationID,
				Status:        domain.StatusProcessing,
				CurrentStep:   "kyc_manual_review_approved",
				ApprovedTier:  decision.ConfirmedTier,
				Details:       kycApprovedDetails,
			}).Get(localCtx, nil)

			kycApproved = true
			break
		}

		// Hard KYC rejection
		return handleRejection(ctx, localCtx, acts, input, []string{kycRes.RejectionReason})
	}

	// Milestone 4: Provision Customer Wallet Account
	progress.Stage = "PROVISIONING_ACCOUNT"
	progress.PercentComplete = 85
	progress.StatusMessage = "Provisioning customer digital wallet account"
	progress.LastUpdated = workflow.Now(ctx)

	targetTier := eligResult.ApprovedTier
	if lastReviewDecision != nil {
		if lastReviewDecision.ConfirmedTier != "" {
			targetTier = lastReviewDecision.ConfirmedTier
		} else if lastReviewDecision.ConfirmedData != nil && lastReviewDecision.ConfirmedData.ApprovedTier != "" {
			targetTier = lastReviewDecision.ConfirmedData.ApprovedTier
		}
	}
	if targetTier == "" {
		targetTier = input.WalletTierRequested
	}
	if targetTier == "" {
		targetTier = "TIER_1_BASIC"
	}

	var provResult activities.ProvisionWalletResult
	err = workflow.ExecuteActivity(localCtx, acts.ProvisionWalletAccountActivity, activities.ProvisionWalletInput{
		CustomerID:   input.CustomerID,
		ApprovedTier: targetTier,
		FullName:     input.PersonalInfo.FullName(),
		Email:        input.PersonalInfo.Email,
	}).Get(localCtx, &provResult)
	if err != nil {
		logger.Error("ProvisionWalletAccountActivity failed", "error", err)
		return nil, fmt.Errorf("wallet provisioning failed: %w", err)
	}

	// Milestone 5: Finalization & Notification
	progress.Stage = "COMPLETED"
	progress.PercentComplete = 100
	progress.StatusMessage = "Onboarding completed successfully"
	progress.LastUpdated = workflow.Now(ctx)

	_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
		ApplicationID:   input.ApplicationID,
		Status:          domain.StatusApproved,
		CurrentStep:     "completed",
		ApprovedTier:    provResult.ApprovedTier,
		WalletAccountID: provResult.WalletAccountID,
	}).Get(localCtx, nil)

	_ = workflow.ExecuteActivity(localCtx, acts.SendOnboardingNotificationActivity, activities.NotificationInput{
		CustomerID:      input.CustomerID,
		Email:           input.PersonalInfo.Email,
		PhoneNumber:     input.PersonalInfo.PhoneNumber,
		FullName:        input.PersonalInfo.FullName(),
		Status:          domain.StatusApproved,
		WalletAccountID: provResult.WalletAccountID,
		ApprovedTier:    provResult.ApprovedTier,
	}).Get(localCtx, nil)

	logger.Info("Customer onboarding workflow successfully completed",
		"customer_id", input.CustomerID,
		"wallet_account_id", provResult.WalletAccountID,
		"tier", provResult.ApprovedTier,
	)

	return &OnboardingWorkflowResult{
		ApplicationID:   input.ApplicationID,
		CustomerID:      input.CustomerID,
		FinalStatus:     domain.StatusApproved,
		ApprovedTier:    provResult.ApprovedTier,
		WalletAccountID: provResult.WalletAccountID,
		CompletedAt:     workflow.Now(ctx),
	}, nil
}

func handleRejection(
	ctx workflow.Context,
	localCtx workflow.Context,
	acts *activities.OnboardingActivities,
	input OnboardingWorkflowInput,
	reasons []string,
) (*OnboardingWorkflowResult, error) {
	_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
		ApplicationID:    input.ApplicationID,
		Status:           domain.StatusRejected,
		RejectionReasons: reasons,
	}).Get(localCtx, nil)

	_ = workflow.ExecuteActivity(localCtx, acts.SendOnboardingNotificationActivity, activities.NotificationInput{
		CustomerID:      input.CustomerID,
		Email:           input.PersonalInfo.Email,
		PhoneNumber:     input.PersonalInfo.PhoneNumber,
		FullName:        input.PersonalInfo.FullName(),
		Status:          domain.StatusRejected,
		RejectionReason: fmt.Sprintf("%v", reasons),
	}).Get(localCtx, nil)

	return &OnboardingWorkflowResult{
		ApplicationID:    input.ApplicationID,
		CustomerID:       input.CustomerID,
		FinalStatus:      domain.StatusRejected,
		RejectionReasons: reasons,
		CompletedAt:      workflow.Now(ctx),
	}, nil
}

func handleCancellation(
	ctx workflow.Context,
	localCtx workflow.Context,
	acts *activities.OnboardingActivities,
	input OnboardingWorkflowInput,
	reason string,
) (*OnboardingWorkflowResult, error) {
	_ = workflow.ExecuteActivity(localCtx, acts.UpdateApplicationStatusActivity, activities.UpdateStatusInput{
		ApplicationID:    input.ApplicationID,
		Status:           domain.StatusCancelled,
		RejectionReasons: []string{reason},
	}).Get(localCtx, nil)

	return &OnboardingWorkflowResult{
		ApplicationID:    input.ApplicationID,
		CustomerID:       input.CustomerID,
		FinalStatus:      domain.StatusCancelled,
		RejectionReasons: []string{reason},
		CompletedAt:      workflow.Now(ctx),
	}, nil
}
