-- Onboarding credentials for secure phone and password based resumption
CREATE TABLE IF NOT EXISTS onboarding_credentials (
    phone_number VARCHAR(50) PRIMARY KEY,
    application_id UUID NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_onboarding_cred_app_id ON onboarding_credentials(application_id);
