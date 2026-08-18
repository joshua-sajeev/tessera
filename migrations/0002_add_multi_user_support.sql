-- +goose Up

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    storage_quota BIGINT DEFAULT 10737418240,
    storage_used BIGINT DEFAULT 0,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Add user_id to assets table
ALTER TABLE assets
ADD COLUMN user_id UUID NOT NULL,
ADD CONSTRAINT fk_assets_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- Add user_id to processing_jobs table
ALTER TABLE processing_jobs
ADD COLUMN user_id UUID NOT NULL,
ADD CONSTRAINT fk_processing_jobs_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- Create indices for user isolation
CREATE INDEX idx_users_api_key ON users(api_key);
CREATE INDEX idx_assets_user_id ON assets(user_id);
CREATE INDEX idx_assets_user_status ON assets(user_id, status);
CREATE INDEX idx_processing_jobs_user_id ON processing_jobs(user_id);
CREATE INDEX idx_processing_jobs_user_status ON processing_jobs(user_id, status);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP INDEX IF EXISTS idx_processing_jobs_user_status;
DROP INDEX IF EXISTS idx_processing_jobs_user_id;
DROP INDEX IF EXISTS idx_assets_user_status;
DROP INDEX IF EXISTS idx_assets_user_id;
DROP INDEX IF EXISTS idx_users_api_key;

ALTER TABLE processing_jobs DROP CONSTRAINT fk_processing_jobs_user_id;
ALTER TABLE processing_jobs DROP COLUMN user_id;

ALTER TABLE assets DROP CONSTRAINT fk_assets_user_id;
ALTER TABLE assets DROP COLUMN user_id;

DROP TABLE IF EXISTS users;
