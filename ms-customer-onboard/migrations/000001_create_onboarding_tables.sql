-- Customer Onboarding Applications table
CREATE TABLE IF NOT EXISTS onboarding_applications (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step VARCHAR(50) NOT NULL,
    completed_steps JSONB DEFAULT '[]'::jsonb,
    personal_info JSONB DEFAULT '{}'::jsonb,
    address_info JSONB DEFAULT '{}'::jsonb,
    identity_docs JSONB DEFAULT '{}'::jsonb,
    financial_info JSONB DEFAULT '{}'::jsonb,
    confirmed_data JSONB DEFAULT NULL,
    wallet_tier_requested VARCHAR(50) DEFAULT 'TIER_1_BASIC',
    approved_tier VARCHAR(50),
    wallet_account_id VARCHAR(100),
    rejection_reasons JSONB DEFAULT '[]'::jsonb,
    workflow_id VARCHAR(128),
    run_id VARCHAR(128),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_onboarding_customer_id ON onboarding_applications(customer_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_phone_number ON onboarding_applications((personal_info->>'phone_number'));
CREATE INDEX IF NOT EXISTS idx_onboarding_status ON onboarding_applications(status);
CREATE INDEX IF NOT EXISTS idx_onboarding_workflow_id ON onboarding_applications(workflow_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_expires_at ON onboarding_applications(expires_at);

-- Audit log for onboarding state transitions and client patches
CREATE TABLE IF NOT EXISTS onboarding_audit_logs (
    id SERIAL PRIMARY KEY,
    application_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50),
    details JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_onboarding_audit_app_id ON onboarding_audit_logs(application_id);

-- Ensure confirmed_data column exists on existing tables
ALTER TABLE onboarding_applications ADD COLUMN IF NOT EXISTS confirmed_data JSONB DEFAULT NULL;
