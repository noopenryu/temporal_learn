package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ms-eligiblity/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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
func (r *Repository) Migrate(ctx context.Context, sqlScript string) error {
	_, err := r.pool.Exec(ctx, sqlScript)
	if err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}
	return nil
}

// Ping checks database connectivity
func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// SaveCheck persists the eligibility check result and applicant details
func (r *Repository) SaveCheck(ctx context.Context, check *domain.EligibilityCheckResult, req *domain.EligibilityCheckRequest) error {
	rejectionJSON, err := json.Marshal(check.RejectionReasons)
	if err != nil {
		return fmt.Errorf("failed to marshal rejection reasons: %w", err)
	}

	rulesJSON, err := json.Marshal(check.RuleResults)
	if err != nil {
		return fmt.Errorf("failed to marshal rule results: %w", err)
	}

	query := `
		INSERT INTO eligibility_checks (
			id, customer_id, national_id, id_type, first_name, last_name,
			date_of_birth, nationality, residence_country, email, phone_number,
			declared_pep, wallet_tier_requested, status, risk_score, risk_level,
			approved_tier, rejection_reasons, rule_results, evaluated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20
		)
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		check.ID,
		req.CustomerID,
		req.NationalID,
		req.IDType,
		req.FirstName,
		req.LastName,
		req.DateOfBirth,
		req.Nationality,
		req.ResidenceCountry,
		req.Email,
		req.PhoneNumber,
		req.DeclaredPEP,
		req.WalletTierRequested,
		string(check.Status),
		check.RiskScore,
		string(check.RiskLevel),
		check.ApprovedTier,
		rejectionJSON,
		rulesJSON,
		check.EvaluatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert eligibility check: %w", err)
	}

	return nil
}

// GetCheckByID retrieves an evaluation by ID
func (r *Repository) GetCheckByID(ctx context.Context, id uuid.UUID) (*domain.EligibilityCheckResult, error) {
	query := `
		SELECT 
			id, customer_id, status, risk_score, risk_level,
			approved_tier, rejection_reasons, rule_results, evaluated_at
		FROM eligibility_checks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	return scanCheckResult(row)
}

// GetLatestCheckByCustomerID retrieves the most recent check for a customer ID
func (r *Repository) GetLatestCheckByCustomerID(ctx context.Context, customerID string) (*domain.EligibilityCheckResult, error) {
	query := `
		SELECT 
			id, customer_id, status, risk_score, risk_level,
			approved_tier, rejection_reasons, rule_results, evaluated_at
		FROM eligibility_checks
		WHERE customer_id = $1
		ORDER BY evaluated_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, customerID)
	return scanCheckResult(row)
}

func scanCheckResult(row pgx.Row) (*domain.EligibilityCheckResult, error) {
	var (
		check       domain.EligibilityCheckResult
		statusStr   string
		levelStr    string
		rejectionRaw []byte
		rulesRaw    []byte
	)

	err := row.Scan(
		&check.ID,
		&check.CustomerID,
		&statusStr,
		&check.RiskScore,
		&levelStr,
		&check.ApprovedTier,
		&rejectionRaw,
		&rulesRaw,
		&check.EvaluatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan eligibility check: %w", err)
	}

	check.Status = domain.EligibilityStatus(statusStr)
	check.RiskLevel = domain.RiskLevel(levelStr)

	if len(rejectionRaw) > 0 {
		_ = json.Unmarshal(rejectionRaw, &check.RejectionReasons)
	}
	if len(rulesRaw) > 0 {
		_ = json.Unmarshal(rulesRaw, &check.RuleResults)
	}

	return &check, nil
}

// IsNationalIDRegistered checks if national ID already has an account
func (r *Repository) IsNationalIDRegistered(ctx context.Context, nationalID, idType string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM existing_identities 
			WHERE national_id = $1 AND id_type = $2
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nationalID, idType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existing identity: %w", err)
	}
	return exists, nil
}

// IsSanctionedOrBlacklisted checks blacklist by national ID or full name match
func (r *Repository) IsSanctionedOrBlacklisted(ctx context.Context, nationalID, fullName, country string) (bool, string, error) {
	query := `
		SELECT reason FROM sanctions_blacklist
		WHERE national_id = $1
		   OR (LOWER(full_name) = LOWER($2) AND ($3 = '' OR country = $3))
		LIMIT 1
	`
	var reason string
	err := r.pool.QueryRow(ctx, query, nationalID, strings.TrimSpace(fullName), strings.TrimSpace(country)).Scan(&reason)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to check sanctions blacklist: %w", err)
	}
	return true, reason, nil
}

// RegisterExistingIdentity registers a customer identity (useful for testing or syncing)
func (r *Repository) RegisterExistingIdentity(ctx context.Context, customerID, nationalID, idType, phoneNumber, email string) error {
	query := `
		INSERT INTO existing_identities (customer_id, national_id, id_type, phone_number, email)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (customer_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, customerID, nationalID, idType, phoneNumber, email)
	if err != nil {
		return fmt.Errorf("failed to register existing identity: %w", err)
	}
	return nil
}
