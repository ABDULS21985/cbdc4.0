-- Wallet Service Schema
CREATE SCHEMA IF NOT EXISTS wallet_db;

CREATE TABLE IF NOT EXISTS wallet_db.users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL, -- Added for login
    password_hash VARCHAR(255) NOT NULL,   -- Added for login
    kyc_data JSONB,
    tier VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallet_db.wallets (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) REFERENCES wallet_db.users(id),
    address VARCHAR(255) NOT NULL, -- On-chain ID
    encrypted_keys TEXT, -- For custodial wallets
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Payments Service Schema
CREATE SCHEMA IF NOT EXISTS payments_db;

CREATE TABLE IF NOT EXISTS payments_db.transactions (
    id VARCHAR(255) PRIMARY KEY,
    from_wallet VARCHAR(255) NOT NULL,
    to_wallet VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL, -- Pending, Confirmed, Failed
    tx_hash VARCHAR(255), -- Fabric Transaction ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Offline Service Schema
CREATE SCHEMA IF NOT EXISTS offline_db;

CREATE TABLE IF NOT EXISTS offline_db.devices (
    id VARCHAR(255) PRIMARY KEY,
    public_key TEXT NOT NULL,
    counter BIGINT DEFAULT 0,
    user_id VARCHAR(255), -- Link to wallet_db.users if needed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS offline_db.vouchers (
    id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255) REFERENCES offline_db.devices(id),
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL, -- Active, Redeemed
    signature TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
