package domain

import (
	"context"

	"github.com/google/uuid"
)

// ListFilter defines parameters for filtering and searching applications
type ListFilter struct {
	Status string `json:"status,omitempty"`
	Search string `json:"search,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// OnboardingRepository specifies the persistence operations for customer onboarding
type OnboardingRepository interface {
	Create(ctx context.Context, app *OnboardingApplication) error
	GetByID(ctx context.Context, id uuid.UUID) (*OnboardingApplication, error)
	GetActiveByCustomerID(ctx context.Context, customerID string) (*OnboardingApplication, error)
	GetActiveByPhoneNumber(ctx context.Context, phone string) (*OnboardingApplication, error)
	List(ctx context.Context, filter ListFilter) ([]*OnboardingApplication, int, error)
	Update(ctx context.Context, app *OnboardingApplication) error
	SaveAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, appID uuid.UUID) ([]AuditLog, error)
	Ping(ctx context.Context) error
}
