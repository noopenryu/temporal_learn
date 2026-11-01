-- Eligibility checks and audit logs
CREATE TABLE IF NOT EXISTS eligibility_checks (
    id UUID PRIMARY KEY,
    customer_id VARCHAR(100) NOT NULL,
    national_id VARCHAR(100) NOT NULL,
    id_type VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    nationality VARCHAR(10) NOT NULL,
    residence_country VARCHAR(10) NOT NULL,
    email VARCHAR(255),
    phone_number VARCHAR(50),
    declared_pep BOOLEAN DEFAULT FALSE,
    wallet_tier_requested VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    risk_score INT NOT NULL,
    risk_level VARCHAR(50) NOT NULL,
    approved_tier VARCHAR(50) NOT NULL,
    rejection_reasons JSONB DEFAULT '[]'::jsonb,
    rule_results JSONB DEFAULT '[]'::jsonb,
    evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_eligibility_checks_customer_id ON eligibility_checks(customer_id);
CREATE INDEX IF NOT EXISTS idx_eligibility_checks_national_id ON eligibility_checks(national_id);
CREATE INDEX IF NOT EXISTS idx_eligibility_checks_status ON eligibility_checks(status);

-- Sanctions, PEP, and Fraud blacklist table
CREATE TABLE IF NOT EXISTS sanctions_blacklist (
    id SERIAL PRIMARY KEY,
    national_id VARCHAR(100),
    full_name VARCHAR(255),
    country VARCHAR(10),
    reason TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'SANCTION', -- SANCTION, AML_FRAUD, PEP
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sanctions_blacklist_national_id ON sanctions_blacklist(national_id);
CREATE INDEX IF NOT EXISTS idx_sanctions_blacklist_full_name ON sanctions_blacklist(full_name);

-- Existing customer identities registry (for duplicate checks in onboarding)
CREATE TABLE IF NOT EXISTS existing_identities (
    id SERIAL PRIMARY KEY,
    customer_id VARCHAR(100) UNIQUE NOT NULL,
    national_id VARCHAR(100) NOT NULL,
    id_type VARCHAR(50) NOT NULL,
    phone_number VARCHAR(50),
    email VARCHAR(255),
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_existing_identities_national_id ON existing_identities(national_id, id_type);

-- Seed initial test sanctions / fraud entries if empty
INSERT INTO sanctions_blacklist (national_id, full_name, country, reason, category)
SELECT 'BL-999-001', 'Test Fraudster', 'US', 'Known fraudulent identity and synthetic identity theft', 'AML_FRAUD'
WHERE NOT EXISTS (SELECT 1 FROM sanctions_blacklist WHERE national_id = 'BL-999-001');

INSERT INTO sanctions_blacklist (national_id, full_name, country, reason, category)
SELECT 'BL-999-002', 'Sanctioned Entity Person', 'RU', 'Specially Designated Nationals (SDN) list match', 'SANCTION'
WHERE NOT EXISTS (SELECT 1 FROM sanctions_blacklist WHERE national_id = 'BL-999-002');
