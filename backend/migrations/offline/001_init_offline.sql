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
