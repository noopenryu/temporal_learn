package temporal

import (
	"fmt"
	"log/slog"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"ms-eligiblity/internal/temporal/activities"
)

// WorkerWrapper encapsulates the Temporal worker lifecycle
type WorkerWrapper struct {
	client    client.Client
	worker    worker.Worker
	taskQueue string
}

// NewWorker initializes a Temporal worker and registers all eligibility activities
func NewWorker(
	c client.Client,
	taskQueue string,
	acts *activities.EligibilityActivities,
) *WorkerWrapper {
	w := worker.New(c, taskQueue, worker.Options{})

	// Register eligibility activities
	w.RegisterActivity(acts.EvaluateEligibilityActivity)
	w.RegisterActivity(acts.QueryEligibilityStatusActivity)

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
