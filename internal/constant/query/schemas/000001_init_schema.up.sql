CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50),
    tier VARCHAR(50) NOT NULL DEFAULT 'BASIC',
    password VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tier ON users(tier);


CREATE TABLE wallets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
    CONSTRAINT balance_non_negative CHECK (balance >= 0)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);



CREATE TABLE transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id VARCHAR(36) NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL,
    fee DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    transaction_type VARCHAR(50) NOT NULL CHECK (
        transaction_type IN ('deposit', 'withdrawal', 'transfer', 'bill_payment')
    ),
    description VARCHAR(255) NOT NULL,
    reference UUID NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'completed', 'failed', 'reversed')
    ),
    metadata JSONB,
    user_tier VARCHAR(50) NOT NULL, -- Denormalized for reporting
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_reference ON transactions(reference);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_type_status ON transactions(transaction_type, status);
CREATE INDEX idx_transactions_provider ON transactions(provider);
CREATE INDEX idx_transactions_user_tier ON transactions(user_tier);

CREATE TABLE bills (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL, -- 'ethiopian_electric', 'ethio_telecom', etc.
    account_number VARCHAR(100) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    fee DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    due_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL, -- 'pending', 'paid', 'failed'
    transaction_id VARCHAR(36) REFERENCES transactions(id),
    metadata JSONB, -- Additional provider-specific data
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bills_user_id ON bills(user_id);
CREATE INDEX idx_bills_provider ON bills(provider);
CREATE INDEX idx_bills_status ON bills(status);
CREATE INDEX idx_bills_transaction_id ON bills(transaction_id);


CREATE TABLE tier_configs (
    tier VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    monthly_fee DECIMAL(10, 2) NOT NULL,
    max_transactions INTEGER NOT NULL,
    max_amount DECIMAL(15, 2) NOT NULL,
    features JSONB NOT NULL, -- {"csv_export": true, "advanced_reports": true, ...}
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Initial data for tiers
INSERT INTO tier_configs (tier, name, monthly_fee, max_transactions, max_amount, features) VALUES
('basic', 'Basic Tier', 0.00, 100, 10000.00, '{"csv_export": false, "advanced_reports": false}'),
('premium', 'Premium Tier', 50.00, 500, 50000.00, '{"csv_export": true, "advanced_reports": true}'),
('enterprise', 'Enterprise Tier', 200.00, 5000, 500000.00, '{"csv_export": true, "advanced_reports": true, "api_access": true}');


CREATE TABLE fee_rules (
    id SERIAL PRIMARY KEY,
    transaction_type VARCHAR(50) NOT NULL,
    base_percentage DECIMAL(5, 2) NOT NULL,
    tier_overrides JSONB NOT NULL, -- {"basic": 3.0, "premium": 1.5, "enterprise": 1.0}
    peak_hours JSONB NOT NULL, -- [{"start_hour": 16, "end_hour": 20, "surcharge": 0.5}]
    fee_cap DECIMAL(15, 2),
    fee_floor DECIMAL(15, 2),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fee_rules_transaction_type ON fee_rules(transaction_type);
CREATE INDEX idx_fee_rules_is_active ON fee_rules(is_active);

-- Initial fee rules
INSERT INTO fee_rules (transaction_type, base_percentage, tier_overrides, peak_hours, fee_cap, fee_floor) VALUES
('bill_payment', 2.5, '{"basic": 3.0, "premium": 1.5, "enterprise": 1.0}', '[{"start_hour": 16, "end_hour": 20, "surcharge": 0.5}]', 100.00, 5.00),
('transfer', 1.0, '{"basic": 1.5, "premium": 0.8, "enterprise": 0.5}', '[]', 50.00, 2.00),
('withdrawal', 0.5, '{"basic": 1.0, "premium": 0.3, "enterprise": 0.1}', '[]', 30.00, 1.00);


CREATE TABLE webhook_events (
    id VARCHAR(36) PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL UNIQUE, -- External event ID
    event_type VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processed', 'failed'
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    next_attempt_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhook_events_event_id ON webhook_events(event_id);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_next_attempt ON webhook_events(next_attempt_at);


CREATE TABLE session_logs (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    login_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    logout_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL -- 'active', 'expired', 'revoked'
);

CREATE INDEX idx_session_logs_user_id ON session_logs(user_id);
CREATE INDEX idx_session_logs_session_id ON session_logs(session_id);
CREATE INDEX idx_session_logs_status ON session_logs(status);


