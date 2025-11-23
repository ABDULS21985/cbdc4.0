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
