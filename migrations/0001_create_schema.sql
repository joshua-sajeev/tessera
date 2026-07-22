-- +goose Up

CREATE TABLE assets (
    id UUID PRIMARY KEY,
    original_filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE processing_jobs (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE TABLE asset_variants (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processing_jobs_asset_id
    ON processing_jobs(asset_id);

CREATE INDEX idx_processing_jobs_status
    ON processing_jobs(status);

CREATE INDEX idx_asset_variants_asset_id
    ON asset_variants(asset_id);

CREATE UNIQUE INDEX idx_asset_variants_asset_type
    ON asset_variants(asset_id, type);

-- +goose Down

DROP TABLE IF EXISTS asset_variants;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS assets;
