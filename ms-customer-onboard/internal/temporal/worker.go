package temporal

import (
	"fmt"
	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ms-customer-onboard/internal/temporal/activities"
	"ms-customer-onboard/internal/temporal/workflows"
)

// WorkerWrapper encapsulates the Temporal worker lifecycle
type WorkerWrapper struct {
	client    client.Client
	worker    worker.Worker
	taskQueue string
}

// NewWorker initializes a Temporal worker and registers the customer onboarding workflow and activities
func NewWorker(
	c client.Client,
	taskQueue string,
	acts *activities.OnboardingActivities,
) *WorkerWrapper {
	w := worker.New(c, taskQueue, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.CustomerOnboardingWorkflow)

	// Register activities
	w.RegisterActivity(acts.UpdateApplicationStatusActivity)
	w.RegisterActivity(acts.ProvisionWalletAccountActivity)
	w.RegisterActivity(acts.SendOnboardingNotificationActivity)
	w.RegisterActivity(acts.ExecuteKycVerificationActivity)

	return &WorkerWrapper{
		client:    c,
		worker:    w,
		taskQueue: taskQueue,
	}
}

// Start begins processing workflow and activity tasks in non-blocking mode
func (w *WorkerWrapper) Start() error {
	slog.Info("Starting Temporal worker", "task_queue", w.taskQueue)
	if err := w.worker.Start(); err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the Temporal worker
func (w *WorkerWrapper) Stop() {
	slog.Info("Stopping Temporal worker", "task_queue", w.taskQueue)
	w.worker.Stop()
}
