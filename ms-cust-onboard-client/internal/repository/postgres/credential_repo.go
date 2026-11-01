package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ms-cust-onboard-client/internal/domain"
	"ms-cust-onboard-client/internal/repository"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) repository.CredentialRepository {
	return &CredentialRepository{pool: pool}
}

// Connect establishes a connection pool to PostgreSQL
func Connect(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// Migrate executes raw SQL migration script
func Migrate(ctx context.Context, pool *pgxpool.Pool, sqlScript string) error {
	_, err := pool.Exec(ctx, sqlScript)
	if err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}
	return nil
}

// Save inserts or updates customer credentials
func (r *CredentialRepository) Save(ctx context.Context, cred *domain.Credential) error {
	if r.pool == nil {
		return errors.New("database pool is not initialized")
	}

	query := `
		INSERT INTO onboarding_credentials (phone_number, application_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (phone_number) DO UPDATE
		SET application_id = EXCLUDED.application_id,
		    password_hash = EXCLUDED.password_hash,
		    updated_at = EXCLUDED.updated_at
	`
	now := time.Now().UTC()
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = now
	}
	cred.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query, cred.PhoneNumber, cred.ApplicationID, cred.PasswordHash, cred.CreatedAt, cred.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save credential: %w", err)
	}
	return nil
}

// GetByPhoneNumber retrieves credential by phone number
func (r *CredentialRepository) GetByPhoneNumber(ctx context.Context, phone string) (*domain.Credential, error) {
	if r.pool == nil {
		return nil, errors.New("database pool is not initialized")
	}

	query := `
		SELECT phone_number, application_id, password_hash, created_at, updated_at
		FROM onboarding_credentials
		WHERE phone_number = $1
	`
	var cred domain.Credential
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&cred.PhoneNumber,
		&cred.ApplicationID,
		&cred.PasswordHash,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get credential by phone: %w", err)
	}
	return &cred, nil
}

// GetByApplicationID retrieves credential by application ID
func (r *CredentialRepository) GetByApplicationID(ctx context.Context, appID uuid.UUID) (*domain.Credential, error) {
	if r.pool == nil {
		return nil, errors.New("database pool is not initialized")
	}

	query := `
		SELECT phone_number, application_id, password_hash, created_at, updated_at
		FROM onboarding_credentials
		WHERE application_id = $1
	`
	var cred domain.Credential
	err := r.pool.QueryRow(ctx, query, appID).Scan(
		&cred.PhoneNumber,
		&cred.ApplicationID,
		&cred.PasswordHash,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get credential by application id: %w", err)
	}
	return &cred, nil
}
