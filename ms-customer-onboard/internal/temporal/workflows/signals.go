package workflows

import (
	"time"

	"ms-customer-onboard/internal/domain"
)

// Signal and Query names
const (
	SignalManualReviewDecision = "manual-review-decision"
	SignalUpdateDocuments      = "update-documents"
	SignalCancelOnboarding     = "cancel-onboarding"
	QueryOnboardingProgress    = "get-onboarding-progress"
)

// ManualReviewDecision contains compliance officer's decision
type ManualReviewDecision struct {
	ReviewerID    string                `json:"reviewer_id"`
	Approved      bool                  `json:"approved"`
	Notes         string                `json:"notes"`
	DecisionTime  time.Time             `json:"decision_time"`
	ConfirmedTier string                `json:"confirmed_tier,omitempty"`
	ConfirmedData *domain.ConfirmedData `json:"confirmed_data,omitempty"`
}

// UpdateDocumentsSignal carries refreshed identity documents from customer
type UpdateDocumentsSignal struct {
	IdentityDocs domain.IdentityDocuments `json:"identity_docs"`
	SubmittedAt  time.Time                `json:"submitted_at"`
}

// CancelOnboardingSignal carries cancellation reason
type CancelOnboardingSignal struct {
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// OnboardingProgressResponse is returned by query
type OnboardingProgressResponse struct {
	Stage           string    `json:"stage"`
	PercentComplete int       `json:"percent_complete"`
	StatusMessage   string    `json:"status_message"`
	LastUpdated     time.Time `json:"last_updated"`
}

// OnboardingWorkflowInput is the payload to launch the onboarding workflow
type OnboardingWorkflowInput struct {
	TraceID              string                   `json:"trace_id,omitempty"`
	SpanID               string                   `json:"span_id,omitempty"`
	ApplicationID        string                   `json:"application_id"`
	CustomerID           string                   `json:"customer_id"`
	PersonalInfo         domain.PersonalInfo      `json:"personal_info"`
	AddressInfo          domain.AddressInfo       `json:"address_info"`
	IdentityDocs         domain.IdentityDocuments `json:"identity_docs"`
	FinancialInfo        domain.FinancialInfo     `json:"financial_info"`
	WalletTierRequested  string                   `json:"wallet_tier_requested"`
	EligibilityTaskQueue string                   `json:"eligibility_task_queue,omitempty"`
	KycTaskQueue         string                   `json:"kyc_task_queue,omitempty"`
}

// OnboardingWorkflowResult is returned when workflow completes
type OnboardingWorkflowResult struct {
	ApplicationID    string                  `json:"application_id"`
	CustomerID       string                  `json:"customer_id"`
	FinalStatus      domain.OnboardingStatus `json:"final_status"`
	ApprovedTier     string                  `json:"approved_tier,omitempty"`
	WalletAccountID  string                  `json:"wallet_account_id,omitempty"`
	RejectionReasons []string                `json:"rejection_reasons,omitempty"`
	CompletedAt      time.Time               `json:"completed_at"`
}
