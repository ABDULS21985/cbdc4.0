-- Create purses table to track offline balances (shadow balance)
CREATE TABLE IF NOT EXISTS offline_db.purses (
    device_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0, -- Shadow balance (funds locked online)
    counter BIGINT NOT NULL DEFAULT 0, -- Last known counter from sync
    last_sync_hash VARCHAR(255),
    last_sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'ACTIVE' -- ACTIVE, LOCKED
);

-- Create used_counters table to prevent double spending
CREATE TABLE IF NOT EXISTS offline_db.used_counters (
    device_id VARCHAR(255) NOT NULL,
    counter BIGINT NOT NULL,
    tx_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (device_id, counter)
);
