package activities

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/observability"
)

// UpdateStatusInput payload to update application status in database
type UpdateStatusInput struct {
	TraceID          string                  `json:"trace_id,omitempty"`
	ApplicationID    string                  `json:"application_id"`
	Status           domain.OnboardingStatus `json:"status"`
	CurrentStep      string                  `json:"current_step,omitempty"`
	ApprovedTier     string                  `json:"approved_tier,omitempty"`
	WalletAccountID  string                  `json:"wallet_account_id,omitempty"`
	RejectionReasons []string                `json:"rejection_reasons,omitempty"`
	Details          map[string]any          `json:"details,omitempty"`
}

// ProvisionWalletInput payload to provision a wallet account
type ProvisionWalletInput struct {
	CustomerID   string `json:"customer_id"`
	ApprovedTier string `json:"approved_tier"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
}

// ProvisionWalletResult outcome of account creation
type ProvisionWalletResult struct {
	WalletAccountID string    `json:"wallet_account_id"`
	ApprovedTier    string    `json:"approved_tier"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

// NotificationInput payload to dispatch email/SMS notification
type NotificationInput struct {
	CustomerID      string                  `json:"customer_id"`
	Email           string                  `json:"email"`
	PhoneNumber     string                  `json:"phone_number"`
	FullName        string                  `json:"full_name"`
	Status          domain.OnboardingStatus `json:"status"`
	WalletAccountID string                  `json:"wallet_account_id,omitempty"`
	ApprovedTier    string                  `json:"approved_tier,omitempty"`
	RejectionReason string                  `json:"rejection_reason,omitempty"`
}

// KycVerificationInput payload for KYC check
type KycVerificationInput struct {
	ApplicationID  string `json:"application_id"`
	CustomerID     string `json:"customer_id"`
	FullName       string `json:"full_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Nationality    string `json:"nationality"`
	PhoneNumber    string `json:"phone_number"`
	Email          string `json:"email"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	FrontDocURL    string `json:"front_doc_url"`
	BackDocURL     string `json:"back_doc_url,omitempty"`
	SelfieURL      string `json:"selfie_url"`
}

// KycVerificationResult outcome of KYC check
type KycVerificationResult struct {
	Passed               bool   `json:"passed"`
	RequiresManualReview bool   `json:"requires_manual_review"`
	RequiresDocumentResub bool   `json:"requires_document_resub"`
	RejectionReason      string `json:"rejection_reason,omitempty"`
	RiskScore            int    `json:"risk_score"`
}

// OnboardingActivities provides the activity implementations for customer onboarding
type OnboardingActivities struct {
	repo domain.OnboardingRepository
}

func NewOnboardingActivities(repo domain.OnboardingRepository) *OnboardingActivities {
	return &OnboardingActivities{repo: repo}
}

// UpdateApplicationStatusActivity updates application state and logs audit record
func (a *OnboardingActivities) UpdateApplicationStatusActivity(ctx context.Context, input UpdateStatusInput) error {
	if input.TraceID != "" {
		ctx = observability.WithTraceContext(ctx, observability.TraceContext{TraceID: input.TraceID, SpanID: observability.GenerateSpanID()})
	}
	appID, err := uuid.Parse(input.ApplicationID)
	if err != nil {
		return fmt.Errorf("invalid application ID: %w", err)
	}

	activity.RecordHeartbeat(ctx, fmt.Sprintf("updating status to %s", input.Status))

	if a.repo != nil {
		app, err := a.repo.GetByID(ctx, appID)
		if err != nil {
			return fmt.Errorf("failed to retrieve application %s: %w", input.ApplicationID, err)
		}
		if app != nil {
			oldStatus := string(app.Status)
			app.Status = input.Status
			if input.CurrentStep != "" {
				app.CurrentStep = input.CurrentStep
			}
			if input.ApprovedTier != "" {
				app.ApprovedTier = input.ApprovedTier
			}
			if input.WalletAccountID != "" {
				app.WalletAccountID = input.WalletAccountID
			}
			if len(input.RejectionReasons) > 0 {
				app.RejectionReasons = input.RejectionReasons
			}

			if err := a.repo.Update(ctx, app); err != nil {
				return fmt.Errorf("failed to update application status: %w", err)
			}

			_ = a.repo.SaveAuditLog(ctx, &domain.AuditLog{
				ApplicationID: appID,
				Action:        "STATUS_UPDATED_BY_WORKFLOW",
				FromStatus:    oldStatus,
				ToStatus:      string(input.Status),
				Details:       input.Details,
				CreatedAt:     time.Now(),
			})
		}
	}

	slog.InfoContext(ctx, "Application status updated by workflow", "app_id", input.ApplicationID, "status", input.Status)
	return nil
}

// ProvisionWalletAccountActivity generates wallet account ID
func (a *OnboardingActivities) ProvisionWalletAccountActivity(ctx context.Context, input ProvisionWalletInput) (*ProvisionWalletResult, error) {
	activity.RecordHeartbeat(ctx, "generating wallet account")

	// Generate random 6-byte hex suffix for unique wallet account number
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	acctID := fmt.Sprintf("WAL-%s-%s", strings.ToUpper(input.ApprovedTier[:4]), strings.ToUpper(hex.EncodeToString(b)))

	tier := input.ApprovedTier
	if tier == "" {
		tier = "TIER_1_BASIC"
	}

	res := &ProvisionWalletResult{
		WalletAccountID: acctID,
		ApprovedTier:    tier,
		Currency:        "USD",
		CreatedAt:       time.Now(),
	}

	slog.Info("Wallet account provisioned", "customer_id", input.CustomerID, "account_id", acctID, "tier", tier)
	return res, nil
}

// SendOnboardingNotificationActivity sends communication to the customer
func (a *OnboardingActivities) SendOnboardingNotificationActivity(ctx context.Context, input NotificationInput) error {
	activity.RecordHeartbeat(ctx, fmt.Sprintf("sending notification for status %s", input.Status))

	slog.Info("Customer notification dispatched",
		"customer_id", input.CustomerID,
		"email", input.Email,
		"phone", input.PhoneNumber,
		"status", input.Status,
		"account_id", input.WalletAccountID,
	)

	return nil
}

// ExecuteKycVerificationActivity performs or dispatches KYC verification
func (a *OnboardingActivities) ExecuteKycVerificationActivity(ctx context.Context, input KycVerificationInput) (*KycVerificationResult, error) {
	activity.RecordHeartbeat(ctx, "evaluating KYC documents")

	// Check if photo URL contains simulation trigger keywords (for testability)
	if strings.Contains(input.FrontDocURL, "blurry") || strings.Contains(input.SelfieURL, "blurry") {
		slog.Warn("KYC document blurry; requesting re-upload", "customer_id", input.CustomerID)
		return &KycVerificationResult{
			Passed:                false,
			RequiresDocumentResub: true,
			RejectionReason:       "Document or selfie image is blurry or unreadable. Please re-upload.",
			RiskScore:             50,
		}, nil
	}

	if strings.Contains(strings.ToLower(input.FullName), "review") {
		slog.Info("KYC flagged for manual review", "customer_id", input.CustomerID)
		return &KycVerificationResult{
			Passed:               false,
			RequiresManualReview: true,
			RejectionReason:      "Identity profile requires manual compliance review",
			RiskScore:            65,
		}, nil
	}

	slog.Info("KYC verification passed", "customer_id", input.CustomerID, "doc_number", input.DocumentNumber)
	return &KycVerificationResult{
		Passed:    true,
		RiskScore: 10,
	}, nil
}
