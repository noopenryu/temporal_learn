package workflows_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	"ms-customer-onboard/internal/domain"
	"ms-customer-onboard/internal/temporal/activities"
	"ms-customer-onboard/internal/temporal/workflows"
)

func createTestWorkflowInput() workflows.OnboardingWorkflowInput {
	return workflows.OnboardingWorkflowInput{
		ApplicationID: "app-test-uuid-1",
		CustomerID:    "cust-wf-1",
		PersonalInfo: domain.PersonalInfo{
			FirstName:   "Elena",
			LastName:    "Rostova",
			DateOfBirth: time.Now().AddDate(-28, 0, 0),
			Nationality: "SG",
			PhoneNumber: "+6581234567",
			Email:       "elena@example.com",
		},
		AddressInfo: domain.AddressInfo{
			Country:       "SG",
			City:          "Singapore",
			StreetAddress: "10 Marina Blvd",
			PostalCode:    "018983",
		},
		IdentityDocs: domain.IdentityDocuments{
			DocumentType:     "PASSPORT",
			DocumentNumber:   "S1234567A",
			IssuingCountry:   "SG",
			FrontDocumentURL: "https://storage/pass.jpg",
			SelfieURL:        "https://storage/selfie.jpg",
		},
		FinancialInfo: domain.FinancialInfo{
			Occupation:    "Consultant",
			SourceOfFunds: "SALARY",
			AnnualIncome:  90000,
			DeclaredPEP:   false,
		},
		WalletTierRequested: "TIER_1_BASIC",
	}
}

func evaluateEligibilityStub(_ context.Context, _ workflows.EligibilityCheckRequest) (*workflows.EligibilityCheckResult, error) {
	return nil, nil
}

func setupTestEnv(testSuite *testsuite.WorkflowTestSuite) (*testsuite.TestWorkflowEnvironment, *activities.OnboardingActivities) {
	env := testSuite.NewTestWorkflowEnvironment()

	// Register cross-service activity definition
	env.RegisterActivityWithOptions(evaluateEligibilityStub, activity.RegisterOptions{
		Name: "EvaluateEligibilityActivity",
	})

	var acts *activities.OnboardingActivities
	env.RegisterActivity(acts.UpdateApplicationStatusActivity)
	env.RegisterActivity(acts.ExecuteKycVerificationActivity)
	env.RegisterActivity(acts.ProvisionWalletAccountActivity)
	env.RegisterActivity(acts.SendOnboardingNotificationActivity)

	return env, acts
}

func TestCustomerOnboardingWorkflow_HappyPath(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env, acts := setupTestEnv(testSuite)

	// Mock activities
	env.OnActivity(acts.UpdateApplicationStatusActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("EvaluateEligibilityActivity", mock.Anything, mock.Anything).Return(&workflows.EligibilityCheckResult{
		Status:       "APPROVED",
		RiskScore:    15,
		ApprovedTier: "TIER_1_BASIC",
	}, nil)
	env.OnActivity(acts.ExecuteKycVerificationActivity, mock.Anything, mock.Anything).Return(&activities.KycVerificationResult{
		Passed:    true,
		RiskScore: 10,
	}, nil)
	env.OnActivity(acts.ProvisionWalletAccountActivity, mock.Anything, mock.Anything).Return(&activities.ProvisionWalletResult{
		WalletAccountID: "WAL-TIER-100100",
		ApprovedTier:    "TIER_1_BASIC",
		Currency:        "USD",
	}, nil)
	env.OnActivity(acts.SendOnboardingNotificationActivity, mock.Anything, mock.Anything).Return(nil)

	input := createTestWorkflowInput()
	env.ExecuteWorkflow(workflows.CustomerOnboardingWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result workflows.OnboardingWorkflowResult
	assert.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, domain.StatusApproved, result.FinalStatus)
	assert.Equal(t, "WAL-TIER-100100", result.WalletAccountID)
	assert.Equal(t, "TIER_1_BASIC", result.ApprovedTier)

	// Verify query handler
	queryVal, err := env.QueryWorkflow(workflows.QueryOnboardingProgress)
	assert.NoError(t, err)
	var progress workflows.OnboardingProgressResponse
	assert.NoError(t, queryVal.Get(&progress))
	assert.Equal(t, 100, progress.PercentComplete)
	assert.Equal(t, "COMPLETED", progress.Stage)
}

func TestCustomerOnboardingWorkflow_EligibilityRejection(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env, acts := setupTestEnv(testSuite)

	env.OnActivity(acts.UpdateApplicationStatusActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("EvaluateEligibilityActivity", mock.Anything, mock.Anything).Return(&workflows.EligibilityCheckResult{
		Status:           "REJECTED",
		RiskScore:        90,
		RejectionReasons: []string{"Applicant is under 18 years old"},
	}, nil)
	env.OnActivity(acts.SendOnboardingNotificationActivity, mock.Anything, mock.Anything).Return(nil)

	input := createTestWorkflowInput()
	env.ExecuteWorkflow(workflows.CustomerOnboardingWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result workflows.OnboardingWorkflowResult
	assert.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, domain.StatusRejected, result.FinalStatus)
	assert.Contains(t, result.RejectionReasons[0], "under 18")
}

func TestCustomerOnboardingWorkflow_KYCBlurryDocSignalResume(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env, acts := setupTestEnv(testSuite)

	env.OnActivity(acts.UpdateApplicationStatusActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity("EvaluateEligibilityActivity", mock.Anything, mock.Anything).Return(&workflows.EligibilityCheckResult{
		Status:       "APPROVED",
		ApprovedTier: "TIER_1_BASIC",
	}, nil)

	// First KYC call returns blurry image request; second call returns pass
	firstCall := true
	env.OnActivity(acts.ExecuteKycVerificationActivity, mock.Anything, mock.Anything).Return(
		func(_ context.Context, _ activities.KycVerificationInput) (*activities.KycVerificationResult, error) {
			if firstCall {
				firstCall = false
				return &activities.KycVerificationResult{
					Passed:                false,
					RequiresDocumentResub: true,
					RejectionReason:       "Document blurry; please resubmit",
				}, nil
			}
			return &activities.KycVerificationResult{
				Passed: true,
			}, nil
		},
	)

	env.OnActivity(acts.ProvisionWalletAccountActivity, mock.Anything, mock.Anything).Return(&activities.ProvisionWalletResult{
		WalletAccountID: "WAL-TIER-RESUBMITTED",
		ApprovedTier:    "TIER_1_BASIC",
	}, nil)
	env.OnActivity(acts.SendOnboardingNotificationActivity, mock.Anything, mock.Anything).Return(nil)

	// Register delayed signal simulating customer re-uploading documents
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(workflows.SignalUpdateDocuments, workflows.UpdateDocumentsSignal{
			IdentityDocs: domain.IdentityDocuments{
				DocumentType:     "PASSPORT",
				DocumentNumber:   "S1234567A",
				IssuingCountry:   "SG",
				FrontDocumentURL: "https://storage/clear_pass.jpg",
				SelfieURL:        "https://storage/clear_selfie.jpg",
			},
			SubmittedAt: time.Now(),
		})
	}, 100*time.Millisecond)

	input := createTestWorkflowInput()
	env.ExecuteWorkflow(workflows.CustomerOnboardingWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result workflows.OnboardingWorkflowResult
	assert.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, domain.StatusApproved, result.FinalStatus)
	assert.Equal(t, "WAL-TIER-RESUBMITTED", result.WalletAccountID)
}

func TestCustomerOnboardingWorkflow_ManualReviewSignalApproved(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env, acts := setupTestEnv(testSuite)

	env.OnActivity(acts.UpdateApplicationStatusActivity, mock.Anything, mock.Anything).Return(nil)
	// Eligibility triggers manual review
	env.OnActivity("EvaluateEligibilityActivity", mock.Anything, mock.Anything).Return(&workflows.EligibilityCheckResult{
		Status:       "REQUIRES_MANUAL_REVIEW",
		ApprovedTier: "TIER_1_BASIC",
	}, nil)
	env.OnActivity(acts.ExecuteKycVerificationActivity, mock.Anything, mock.Anything).Return(&activities.KycVerificationResult{
		Passed: true,
	}, nil)
	env.OnActivity(acts.ProvisionWalletAccountActivity, mock.Anything, mock.Anything).Return(&activities.ProvisionWalletResult{
		WalletAccountID: "WAL-MANUAL-APPROVED",
		ApprovedTier:    "TIER_1_BASIC",
	}, nil)
	env.OnActivity(acts.SendOnboardingNotificationActivity, mock.Anything, mock.Anything).Return(nil)

	// Register compliance officer approval signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(workflows.SignalManualReviewDecision, workflows.ManualReviewDecision{
			ReviewerID: "compliance-officer-01",
			Approved:   true,
			Notes:      "Cleared after address and background verification",
		})
	}, 50*time.Millisecond)

	input := createTestWorkflowInput()
	env.ExecuteWorkflow(workflows.CustomerOnboardingWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())

	var result workflows.OnboardingWorkflowResult
	assert.NoError(t, env.GetWorkflowResult(&result))
	assert.Equal(t, domain.StatusApproved, result.FinalStatus)
	assert.Equal(t, "WAL-MANUAL-APPROVED", result.WalletAccountID)
}
