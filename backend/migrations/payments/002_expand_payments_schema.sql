ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'NGN';
ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS fee BIGINT DEFAULT 0;
ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'P2P';
ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS channel VARCHAR(50) DEFAULT 'MOBILE';
ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE payments_db.transactions ADD COLUMN IF NOT EXISTS description TEXT;
