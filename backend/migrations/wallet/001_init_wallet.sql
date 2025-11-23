CREATE SCHEMA IF NOT EXISTS wallet_db;

CREATE TABLE IF NOT EXISTS wallet_db.wallets (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) REFERENCES wallet_db.users(id),
    address VARCHAR(255) NOT NULL, -- On-chain ID
    encrypted_keys TEXT, -- For custodial wallets
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
