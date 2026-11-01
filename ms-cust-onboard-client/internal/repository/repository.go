package repository

import (
	"context"

	"github.com/google/uuid"
	"ms-cust-onboard-client/internal/domain"
)

// CredentialRepository interface
type CredentialRepository interface {
	Save(ctx context.Context, cred *domain.Credential) error
	GetByPhoneNumber(ctx context.Context, phone string) (*domain.Credential, error)
	GetByApplicationID(ctx context.Context, appID uuid.UUID) (*domain.Credential, error)
}
