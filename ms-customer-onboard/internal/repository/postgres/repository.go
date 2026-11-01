package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ms-customer-onboard/internal/domain"
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

// Create inserts a new onboarding application record
func (r *Repository) Create(ctx context.Context, app *domain.OnboardingApplication) error {
	completedJSON, err := json.Marshal(app.CompletedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal completed steps: %w", err)
	}
	personalJSON, err := json.Marshal(app.PersonalInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal personal info: %w", err)
	}
	addressJSON, err := json.Marshal(app.AddressInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal address info: %w", err)
	}
	identityJSON, err := json.Marshal(app.IdentityDocs)
	if err != nil {
		return fmt.Errorf("failed to marshal identity docs: %w", err)
	}
	financialJSON, err := json.Marshal(app.FinancialInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal financial info: %w", err)
	}
	var confirmedJSON []byte
	if app.ConfirmedData != nil {
		confirmedJSON, err = json.Marshal(app.ConfirmedData)
		if err != nil {
			return fmt.Errorf("failed to marshal confirmed data: %w", err)
		}
	}
	rejectionJSON, err := json.Marshal(app.RejectionReasons)
	if err != nil {
		return fmt.Errorf("failed to marshal rejection reasons: %w", err)
	}

	query := `
		INSERT INTO onboarding_applications (
			id, customer_id, status, current_step, completed_steps,
			personal_info, address_info, identity_docs, financial_info,
			confirmed_data, wallet_tier_requested, approved_tier, wallet_account_id,
			rejection_reasons, workflow_id, run_id, expires_at,
			submitted_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10,
			$11, $12, $13,
			$14, $15, $16, $17,
			$18, $19, $20
		)
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		app.ID,
		app.CustomerID,
		string(app.Status),
		app.CurrentStep,
		completedJSON,
		personalJSON,
		addressJSON,
		identityJSON,
		financialJSON,
		confirmedJSON,
		app.WalletTierRequested,
		app.ApprovedTier,
		app.WalletAccountID,
		rejectionJSON,
		app.WorkflowID,
		app.RunID,
		app.ExpiresAt,
		app.SubmittedAt,
		app.CreatedAt,
		app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert onboarding application: %w", err)
	}

	return nil
}

// GetByID retrieves an onboarding application by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	query := `
		SELECT 
			id, customer_id, status, current_step, completed_steps,
			personal_info, address_info, identity_docs, financial_info,
			confirmed_data, wallet_tier_requested, approved_tier, wallet_account_id,
			rejection_reasons, workflow_id, run_id, expires_at,
			submitted_at, created_at, updated_at
		FROM onboarding_applications
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanApplication(row)
}

// GetActiveByCustomerID finds the latest non-terminal application for a customer
func (r *Repository) GetActiveByCustomerID(ctx context.Context, customerID string) (*domain.OnboardingApplication, error) {
	query := `
		SELECT 
			id, customer_id, status, current_step, completed_steps,
			personal_info, address_info, identity_docs, financial_info,
			confirmed_data, wallet_tier_requested, approved_tier, wallet_account_id,
			rejection_reasons, workflow_id, run_id, expires_at,
			submitted_at, created_at, updated_at
		FROM onboarding_applications
		WHERE customer_id = $1 AND status NOT IN ('CANCELLED', 'EXPIRED', 'REJECTED')
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, customerID)
	return scanApplication(row)
}

// GetActiveByPhoneNumber finds the latest active or in-progress application for a phone number
func (r *Repository) GetActiveByPhoneNumber(ctx context.Context, phone string) (*domain.OnboardingApplication, error) {
	normalized := domain.NormalizePhoneNumber(phone)
	query := `
		SELECT 
			id, customer_id, status, current_step, completed_steps,
			personal_info, address_info, identity_docs, financial_info,
			confirmed_data, wallet_tier_requested, approved_tier, wallet_account_id,
			rejection_reasons, workflow_id, run_id, expires_at,
			submitted_at, created_at, updated_at
		FROM onboarding_applications
		WHERE (
			personal_info->>'phone_number' = $1
			OR regexp_replace(COALESCE(personal_info->>'phone_number', ''), '[^0-9+]', '', 'g') = $2
		) AND status NOT IN ('CANCELLED', 'EXPIRED')
		ORDER BY created_at DESC
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, phone, normalized)
	return scanApplication(row)
}

// List retrieves a paginated list of applications matching optional status and search terms
func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.OnboardingApplication, int, error) {
	whereClauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if filter.Status != "" && filter.Status != "ALL" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if strings.TrimSpace(filter.Search) != "" {
		searchPattern := "%" + strings.TrimSpace(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			customer_id ILIKE $%d 
			OR personal_info->>'first_name' ILIKE $%d 
			OR personal_info->>'last_name' ILIKE $%d 
			OR personal_info->>'phone_number' ILIKE $%d 
			OR personal_info->>'email' ILIKE $%d 
			OR id::text ILIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx, argIdx, argIdx))
		args = append(args, searchPattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM onboarding_applications WHERE %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count onboarding applications: %w", err)
	}

	// 2. Fetch records
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			id, customer_id, status, current_step, completed_steps,
			personal_info, address_info, identity_docs, financial_info,
			confirmed_data, wallet_tier_requested, approved_tier, wallet_account_id,
			rejection_reasons, workflow_id, run_id, expires_at,
			submitted_at, created_at, updated_at
		FROM onboarding_applications
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query onboarding applications: %w", err)
	}
	defer rows.Close()

	var apps []*domain.OnboardingApplication
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan application row: %w", err)
		}
		if app != nil {
			apps = append(apps, app)
		}
	}

	return apps, total, nil
}

// Update saves changes to an existing onboarding application
func (r *Repository) Update(ctx context.Context, app *domain.OnboardingApplication) error {
	completedJSON, err := json.Marshal(app.CompletedSteps)
	if err != nil {
		return fmt.Errorf("failed to marshal completed steps: %w", err)
	}
	personalJSON, err := json.Marshal(app.PersonalInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal personal info: %w", err)
	}
	addressJSON, err := json.Marshal(app.AddressInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal address info: %w", err)
	}
	identityJSON, err := json.Marshal(app.IdentityDocs)
	if err != nil {
		return fmt.Errorf("failed to marshal identity docs: %w", err)
	}
	financialJSON, err := json.Marshal(app.FinancialInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal financial info: %w", err)
	}
	var confirmedJSON []byte
	if app.ConfirmedData != nil {
		confirmedJSON, err = json.Marshal(app.ConfirmedData)
		if err != nil {
			return fmt.Errorf("failed to marshal confirmed data: %w", err)
		}
	}
	rejectionJSON, err := json.Marshal(app.RejectionReasons)
	if err != nil {
		return fmt.Errorf("failed to marshal rejection reasons: %w", err)
	}

	app.UpdatedAt = time.Now()

	query := `
		UPDATE onboarding_applications SET
			status = $2,
			current_step = $3,
			completed_steps = $4,
			personal_info = $5,
			address_info = $6,
			identity_docs = $7,
			financial_info = $8,
			confirmed_data = $9,
			wallet_tier_requested = $10,
			approved_tier = $11,
			wallet_account_id = $12,
			rejection_reasons = $13,
			workflow_id = $14,
			run_id = $15,
			expires_at = $16,
			submitted_at = $17,
			updated_at = $18
		WHERE id = $1
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		app.ID,
		string(app.Status),
		app.CurrentStep,
		completedJSON,
		personalJSON,
		addressJSON,
		identityJSON,
		financialJSON,
		confirmedJSON,
		app.WalletTierRequested,
		app.ApprovedTier,
		app.WalletAccountID,
		rejectionJSON,
		app.WorkflowID,
		app.RunID,
		app.ExpiresAt,
		app.SubmittedAt,
		app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update onboarding application: %w", err)
	}

	return nil
}

// SaveAuditLog inserts an audit log entry
func (r *Repository) SaveAuditLog(ctx context.Context, log *domain.AuditLog) error {
	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal audit details: %w", err)
	}

	query := `
		INSERT INTO onboarding_audit_logs (
			application_id, action, from_status, to_status, details, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err = r.pool.Exec(ctx, query, log.ApplicationID, log.Action, log.FromStatus, log.ToStatus, detailsJSON, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

// GetAuditLogs retrieves all audit trail entries for an application
func (r *Repository) GetAuditLogs(ctx context.Context, appID uuid.UUID) ([]domain.AuditLog, error) {
	query := `
		SELECT id, application_id, action, from_status, to_status, details, created_at
		FROM onboarding_audit_logs
		WHERE application_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var (
			item        domain.AuditLog
			detailsRaw  []byte
			fromStatus  *string
			toStatus    *string
		)

		err := rows.Scan(
			&item.ID,
			&item.ApplicationID,
			&item.Action,
			&fromStatus,
			&toStatus,
			&detailsRaw,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log row: %w", err)
		}

		if fromStatus != nil {
			item.FromStatus = *fromStatus
		}
		if toStatus != nil {
			item.ToStatus = *toStatus
		}
		if len(detailsRaw) > 0 {
			_ = json.Unmarshal(detailsRaw, &item.Details)
		}
		logs = append(logs, item)
	}

	return logs, nil
}

func scanApplication(row pgx.Row) (*domain.OnboardingApplication, error) {
	var (
		app           domain.OnboardingApplication
		statusStr     string
		completedRaw  []byte
		personalRaw   []byte
		addressRaw    []byte
		identityRaw   []byte
		financialRaw  []byte
		confirmedRaw  []byte
		rejectionRaw  []byte
		approvedTier  *string
		walletAcctID  *string
		workflowID    *string
		runID         *string
	)

	err := row.Scan(
		&app.ID,
		&app.CustomerID,
		&statusStr,
		&app.CurrentStep,
		&completedRaw,
		&personalRaw,
		&addressRaw,
		&identityRaw,
		&financialRaw,
		&confirmedRaw,
		&app.WalletTierRequested,
		&approvedTier,
		&walletAcctID,
		&rejectionRaw,
		&workflowID,
		&runID,
		&app.ExpiresAt,
		&app.SubmittedAt,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan onboarding application: %w", err)
	}

	app.Status = domain.OnboardingStatus(statusStr)
	if approvedTier != nil {
		app.ApprovedTier = *approvedTier
	}
	if walletAcctID != nil {
		app.WalletAccountID = *walletAcctID
	}
	if workflowID != nil {
		app.WorkflowID = *workflowID
	}
	if runID != nil {
		app.RunID = *runID
	}

	if len(completedRaw) > 0 {
		_ = json.Unmarshal(completedRaw, &app.CompletedSteps)
	}
	if len(personalRaw) > 0 {
		_ = json.Unmarshal(personalRaw, &app.PersonalInfo)
	}
	if len(addressRaw) > 0 {
		_ = json.Unmarshal(addressRaw, &app.AddressInfo)
	}
	if len(identityRaw) > 0 {
		_ = json.Unmarshal(identityRaw, &app.IdentityDocs)
	}
	if len(financialRaw) > 0 {
		_ = json.Unmarshal(financialRaw, &app.FinancialInfo)
	}
	if len(confirmedRaw) > 0 && string(confirmedRaw) != "null" {
		var confirmed domain.ConfirmedData
		if err := json.Unmarshal(confirmedRaw, &confirmed); err == nil {
			app.ConfirmedData = &confirmed
		}
	}
	if len(rejectionRaw) > 0 {
		_ = json.Unmarshal(rejectionRaw, &app.RejectionReasons)
	}

	return &app, nil
}
