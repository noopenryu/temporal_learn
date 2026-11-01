package domain

import (
	"context"

	"github.com/google/uuid"
)

// EligibilityRepository defines the persistence contract for eligibility data
type EligibilityRepository interface {
	// SaveCheck saves the complete evaluation result and its rule details
	SaveCheck(ctx context.Context, check *EligibilityCheckResult, req *EligibilityCheckRequest) error

	// GetCheckByID retrieves an evaluation by its unique check ID
	GetCheckByID(ctx context.Context, id uuid.UUID) (*EligibilityCheckResult, error)

	// GetLatestCheckByCustomerID retrieves the most recent evaluation for a customer
	GetLatestCheckByCustomerID(ctx context.Context, customerID string) (*EligibilityCheckResult, error)

	// IsNationalIDRegistered checks if a national ID / passport is already linked to an existing account
	IsNationalIDRegistered(ctx context.Context, nationalID, idType string) (bool, error)

	// IsSanctionedOrBlacklisted checks against sanctions, AML blacklist, or PEP lists
	IsSanctionedOrBlacklisted(ctx context.Context, nationalID, fullName, country string) (bool, string, error)

	// RegisterExistingIdentity registers an identity into the existing accounts registry (for simulation/integration)
	RegisterExistingIdentity(ctx context.Context, customerID, nationalID, idType, phoneNumber, email string) error

	// Ping checks database health
	Ping(ctx context.Context) error
}
