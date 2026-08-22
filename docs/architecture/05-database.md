# Database Design

Tessera uses **PostgreSQL** for metadata and relationships. Multi-user isolation is enforced through explicit `user_id` foreign keys on all tables.

---

## Schema Overview

```mermaid
erDiagram
    USERS ||--o{ ASSETS : owns
    USERS ||--o{ PROCESSING_JOBS : owns
    ASSETS ||--o{ PROCESSING_JOBS : processes
    ASSETS ||--o{ ASSET_VARIANTS : generates

    USERS {
        UUID id PK
        TEXT username UK
        TEXT email UK
        TEXT api_key_id UK
        TEXT api_key_hash
        BIGINT storage_quota
        BIGINT storage_used
        TEXT status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    ASSETS {
        UUID id PK
        UUID user_id FK
        TEXT original_filename
        TEXT content_type
        BIGINT size
        TEXT storage_path
        TEXT status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
    }

    PROCESSING_JOBS {
        UUID id PK
        UUID user_id FK
        UUID asset_id FK
        TEXT status
        TIMESTAMPTZ created_at
        TIMESTAMPTZ started_at
        TIMESTAMPTZ completed_at
    }

    ASSET_VARIANTS {
        UUID id PK
        UUID asset_id FK
        TEXT type
        TEXT content_type
        BIGINT size
        TEXT storage_path
        TIMESTAMPTZ created_at
    }
````

---

## Tables

### USERS

Central identity table for multi-tenant isolation.

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `username` | TEXT | UNIQUE, login identifier |
| `email` | TEXT | UNIQUE, contact |
| `storage_quota` | BIGINT | Max bytes (default: 10GB) |
| `storage_used` | BIGINT | Current usage in bytes |
| `status` | TEXT | active, suspended, deleted |
| `created_at` | TIMESTAMPTZ | Account creation |
| `updated_at` | TIMESTAMPTZ | Last update |
| `api_key_id`    | TEXT        | UNIQUE, API key lookup identifier |
| `api_key_hash`  | TEXT        | Argon2id-hashed API key secret    |

### ASSETS

Asset metadata owned by users.

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK → USERS(id), enforces ownership |
| `original_filename` | TEXT | Filename |
| `content_type` | TEXT | MIME type |
| `size` | BIGINT | File size in bytes |
| `storage_path` | TEXT | MinIO path |
| `status` | TEXT | uploaded, processing, ready, failed |
| `created_at` | TIMESTAMPTZ | Upload time |
| `updated_at` | TIMESTAMPTZ | Last modified |

### PROCESSING_JOBS

Async job tracking with denormalized `user_id` for fairness.

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK → USERS(id), enables per-user queries |
| `asset_id` | UUID | FK → ASSETS(id) |
| `status` | TEXT | queued, processing, completed, failed |
| `created_at` | TIMESTAMPTZ | Job created |
| `started_at` | TIMESTAMPTZ | Processing started (nullable) |
| `completed_at` | TIMESTAMPTZ | Processing finished (nullable) |

### ASSET_VARIANTS

Generated variants (thumbnails, optimized copies).

| Column | Type | Notes |
|--------|------|-------|
| `id` | UUID | Primary key |
| `asset_id` | UUID | FK → ASSETS(id) |
| `type` | TEXT | thumbnail, optimized, preview, etc. |
| `content_type` | TEXT | MIME type of variant |
| `size` | BIGINT | File size in bytes |
| `storage_path` | TEXT | MinIO path |
| `created_at` | TIMESTAMPTZ | Generated at |

---

## Multi-Tenancy Enforcement

### Why user_id in Every Table?

**Database-level isolation:**

```sql
-- User A cannot access User B's assets (enforced by database)
SELECT * FROM assets WHERE id = 'asset-123' AND user_id = 'user-b';
-- Returns empty → No leak possible
```

**Repository method signatures:**

```go
// CORRECT: user_id parameter is mandatory
func (r *AssetRepository) GetByID(ctx context.Context, assetID, userID uuid.UUID) (*Asset, error)

// WRONG: This signature doesn't exist (compiler won't allow)
// func (r *AssetRepository) GetByID(ctx context.Context, assetID uuid.UUID) (*Asset, error)
```

**Why this works:**

1. Developer cannot forget `user_id` parameter (won't compile)
2. Database prevents cross-user access (foreign key + WHERE clause)
3. Clear audit trail of ownership

---

## Indices

Composite indices on `(user_id, status)` for fast user-scoped queries:

```sql
-- User isolation
CREATE INDEX idx_assets_user_status ON assets(user_id, status);
CREATE INDEX idx_processing_jobs_user_status ON processing_jobs(user_id, status);

-- Asset lookups
CREATE INDEX idx_processing_jobs_asset_id ON processing_jobs(asset_id);
CREATE INDEX idx_asset_variants_asset_id ON asset_variants(asset_id);

-- Job cleanup (find old jobs)
CREATE INDEX idx_processing_jobs_created_at ON processing_jobs(created_at);

-- Authentication (API key lookup)
-- api_key_id already has a unique index from its UNIQUE constraint.
```

---

## Migrations

All schema changes are version-controlled with **Goose**.

### Naming Convention

```
migrations/NNNN_description.sql
```

- **NNNN**: 4-digit sequence (ordering)
- **description**: snake_case, describes change

### Example: 0002_add_multi_user_support.sql

```sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    api_key_id TEXT UNIQUE NOT NULL,
    api_key_hash TEXT NOT NULL,
    storage_quota BIGINT NOT NULL DEFAULT 10737418240,
    storage_used BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE assets ADD COLUMN user_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE assets ADD CONSTRAINT fk_assets_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE assets ALTER COLUMN user_id DROP DEFAULT;

ALTER TABLE processing_jobs ADD COLUMN user_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
ALTER TABLE processing_jobs ADD CONSTRAINT fk_processing_jobs_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE processing_jobs ALTER COLUMN user_id DROP DEFAULT;

CREATE INDEX idx_assets_user_status ON assets(user_id, status);
CREATE INDEX idx_processing_jobs_user_status ON processing_jobs(user_id, status);

-- +goose Down
DROP INDEX idx_processing_jobs_user_status;
DROP INDEX idx_assets_user_status;
ALTER TABLE processing_jobs DROP CONSTRAINT fk_processing_jobs_user_id;
ALTER TABLE processing_jobs DROP COLUMN user_id;
ALTER TABLE assets DROP CONSTRAINT fk_assets_user_id;
ALTER TABLE assets DROP COLUMN user_id;
DROP TABLE users;
```

---

## Running Migrations

```bash
make migrate           # Apply pending migrations
make goose-status      # Check status
goose up              # Apply next
goose down            # Rollback one
```

---

## Data Access Pattern

All queries must include `user_id` to prevent cross-user access:

```go
// CORRECT: Isolation enforced
asset, err := repo.GetByID(ctx, assetID, userID)
// Executes: SELECT ... FROM assets WHERE id = $1 AND user_id = $2

// CORRECT: List user's assets only
assets, err := repo.ListByUser(ctx, userID)
// Executes: SELECT ... FROM assets WHERE user_id = $1

// WRONG: Missing user_id check
// func GetByID(ctx, assetID) { ... }  // Won't compile
```

---

## Setup

### PostgreSQL Connection

```bash
# Environment variables
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=tessera
```

### Start Database

```bash
make up          # Start PostgreSQL in Docker
make migrate     # Apply migrations
```

---

## Related Documents

- **[ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md)** — Design rationale
- **[ADR 002 - Database & Storage Separation](../decisions/002-database-and-storage-separation.md)** — Why PostgreSQL + MinIO
- **[04 - Guidelines](04-guidelines.md)** — Data access patterns and conventions

---

## Navigation

**Previous:** [04 - Guidelines](04-guidelines.md)  
**Return:** [Architecture Index](README.md)
