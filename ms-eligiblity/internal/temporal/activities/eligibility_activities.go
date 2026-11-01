package activities

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"ms-eligiblity/internal/domain"
	"ms-eligiblity/internal/observability"
	"ms-eligiblity/internal/service"
)

// EligibilityActivities holds the Temporal activity handlers for customer eligibility
type EligibilityActivities struct {
	svc service.EligibilityService
}

func NewEligibilityActivities(svc service.EligibilityService) *EligibilityActivities {
	return &EligibilityActivities{svc: svc}
}

// EvaluateEligibilityActivity evaluates customer onboarding eligibility
func (a *EligibilityActivities) EvaluateEligibilityActivity(
	ctx context.Context,
	req domain.EligibilityCheckRequest,
) (*domain.EligibilityCheckResult, error) {
	if req.TraceID != "" {
		ctx = observability.WithTraceContext(ctx, observability.TraceContext{TraceID: req.TraceID, SpanID: observability.GenerateSpanID()})
	}

	slog.InfoContext(ctx, "Evaluating customer eligibility rules", "customer_id", req.CustomerID, "tier", req.WalletTierRequested)
	activity.RecordHeartbeat(ctx, "evaluating customer eligibility rules")

	result, err := a.svc.Evaluate(ctx, &req)
	if err != nil {
		slog.ErrorContext(ctx, "Eligibility evaluation failed", "error", err, "customer_id", req.CustomerID)
		// Non-retryable error if validation failed
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("eligibility evaluation error: %v", err),
			"ERR_ELIGIBILITY_EVALUATION",
			err,
		)
	}

	slog.InfoContext(ctx, "Completed customer eligibility evaluation", "customer_id", req.CustomerID, "status", result.Status)
	activity.RecordHeartbeat(ctx, fmt.Sprintf("completed evaluation: status=%s", result.Status))
	return result, nil
}

// QueryEligibilityStatusActivity retrieves an existing check result by check ID
func (a *EligibilityActivities) QueryEligibilityStatusActivity(
	ctx context.Context,
	checkID string,
) (*domain.EligibilityCheckResult, error) {
	id, err := uuid.Parse(checkID)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("invalid check ID format: %v", err),
			"ERR_INVALID_CHECK_ID",
			err,
		)
	}

	result, err := a.svc.GetCheckByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get check by ID: %w", err)
	}
	if result == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("eligibility check %s not found", checkID),
			"ERR_CHECK_NOT_FOUND",
			nil,
		)
	}

	return result, nil
}
