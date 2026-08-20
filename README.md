# Tessera

> An asset processing platform built in Go using Hexagonal Architecture.

Tessera is a backend service for storing, processing, and serving digital assets. It implements **multi-user isolation from the ground up** using a clean, modular architecture where business logic is isolated from infrastructure.

The project is being built as a learning-focused backend system exploring asynchronous processing, object storage, and scalable multi-tenant service design.

---

## Current Status

The project has implemented the persistence foundation and multi-user infrastructure for v1.

### ✅ Implemented

- [x] Domain models (Asset, ProcessingJob, AssetVariant, User)
- [x] Repository ports with explicit user_id enforcement
- [x] PostgreSQL persistence adapters (multi-tenant queries)
- [x] Goose database migrations (users table, foreign keys)
- [x] PostgreSQL integration tests (isolation verification)
- [x] MinIO object storage adapter (S3-compatible)
- [x] MinIO integration tests
- [x] Docker Compose development environment
- [x] Multi-user database schema with indices
- [x] User domain model and authentication foundation

### 📋 Planned (v1 Completion)

- [ ] Application layer (use cases)
- [ ] HTTP API with REST endpoints
- [ ] Bearer token authentication middleware
- [ ] Redis job queue with per-user fairness
- [ ] Background worker process
- [ ] Asset processing pipeline
- [ ] End-to-end integration tests

### 🔮 Future (v2+)

- [ ] Organizations/workspace support
- [ ] Role-based access control (RBAC)
- [ ] Fine-grained API key permissions
- [ ] Webhook notifications
- [ ] Folder/asset organization
- [ ] Audit logging and compliance
- [ ] Usage metrics and billing
- [ ] Real-time updates (WebSocket)

---

## Architecture

Tessera follows **Hexagonal Architecture (Ports and Adapters)** with **multi-user isolation at the database layer**.

```text
External World (HTTP, PostgreSQL, MinIO, Redis)
        ↓
    Adapters (Port implementations)
        ↓
  Application (Use Cases)
        ↓
    Domain (Business Logic)
        ↑
    Ports (Interfaces/Contracts)
```

### Multi-Tenant Design

All data access is **explicitly user-scoped**:

```go
// Repository methods require user_id parameter
asset, err := assetRepo.GetByID(ctx, assetID, userID)  // ✓ Correct
asset, err := assetRepo.GetByID(ctx, assetID)          // ✗ Won't compile
```

**Database enforcement:**
- All tables have user_id foreign keys
- Composite indices on (user_id, status) for fast queries
- WHERE clauses always include user_id
- Cross-user data access is prevented at database level

For details, see:
- [Architecture Documentation](./docs/architecture/README.md)
- [ADR 004 - Multi-Tenancy Strategy](./docs/decisions/004-multi-tenancy-strategy.md)
- [05 - Database](./docs/architecture/05-database.md)

---

## Tech Stack

| Category | Technology |
|----------|---|
| Language | Go 1.21+ |
| Database | PostgreSQL 15+ |
| Database Driver | pgx/v5 |
| Migrations | Goose |
| Object Storage | MinIO (S3-compatible) |
| Testing | Go `testing` package |
| Containers | Docker Compose |

---

## Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Goose (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- PostgreSQL 15+ (via Docker)

### Clone & Setup

```bash
git clone https://github.com/joshua-sajeev/tessera.git
cd tessera
```

### Start Development Services

```bash
make up
```

This starts:
- PostgreSQL 15 (port 5432)
- MinIO (port 9000, UI on 9001)
- Redis (port 6379) — optional, for queue implementation

### Run Database Migrations

```bash
make migrate
```

Or manually:
```bash
goose up
```

### Check Migration Status

```bash
make goose-status
```

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/adapters/postgres
```

### Stop Services

```bash
make down
```

---

## Project Structure

```text
tessera/
├── cmd/                          # Application entrypoints
│   └── tessera/
│       └── main.go               # (planned)
├── internal/
│   ├── domain/                   # Business entities
│   │   ├── asset.go
│   │   ├── processing_job.go
│   │   ├── asset_variant.go
│   │   └── user.go
│   ├── ports/                    # Interfaces (contracts)
│   │   ├── asset_repository.go
│   │   ├── processing_repository.go
│   │   ├── storage.go
│   │   └── authenticator.go      # (planned)
│   ├── adapters/                 # Infrastructure implementations
│   │   ├── postgres/
│   │   │   ├── asset_repository.go
│   │   │   ├── processing_repository.go
│   │   │   └── *_test.go
│   │   ├── minio/
│   │   │   ├── storage.go
│   │   │   └── storage_test.go
│   │   └── authenticator/        # (planned)
│   ├── config/                   # Configuration
│   │   └── config.go
│   └── middleware/               # HTTP middleware (planned)
├── migrations/                   # Goose database migrations
│   ├── 0001_create_schema.sql
│   └── 0002_add_multi_user_support.sql
├── deployments/                  # Deployment configs
│   └── docker-compose.yaml
├── docs/                         # Documentation
│   ├── architecture/             # Architecture ADRs and guides
│   │   ├── README.md
│   │   ├── 00-overview.md
│   │   ├── 01-layers.md
│   │   ├── 02-flows.md
│   │   ├── 03-structure.md
│   │   ├── 04-guidelines.md
│   │   └── 05-database.md
│   └── decisions/                # Architecture Decision Records
│       ├── 001-hexagonal-architecture.md
│       ├── 002-database-and-storage-separation.md
│       ├── 003-minio-object-storage.md
│       └── 004-multi-tenancy-strategy.md
├── Makefile
├── go.mod
├── go.sum
└── README.md                     # This file
```

### Detailed Module Breakdown

**See [03 - Structure](./docs/architecture/03-structure.md)** for comprehensive repository layout with responsibilities.

---

## Making Requests

### API Authentication (Planned)

```bash
# Get your API key (from user creation/management)
API_KEY="your-api-key-here"

# Upload an asset with Bearer token
curl -X POST http://localhost:8080/assets \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@image.jpg"

# Expected response (v1.0)
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "status": "uploaded",
#   "created_at": "2026-08-20T12:34:56Z"
# }
```

**Status:** HTTP API not yet implemented. Expected in v1.0 milestone.

---

## Documentation

Architecture and design documentation is located in `docs/`:

### Architecture Guides

- **[00 - Overview](./docs/architecture/00-overview.md)** — Project goals, features, and architectural style
- **[01 - Layers](./docs/architecture/01-layers.md)** — Responsibility of each layer (Domain, Ports, Adapters, Application)
- **[02 - Flows](./docs/architecture/02-flows.md)** — Request sequences and processing pipelines
- **[03 - Structure](./docs/architecture/03-structure.md)** — Repository layout and module organization
- **[04 - Guidelines](./docs/architecture/04-guidelines.md)** — Development conventions and patterns
- **[05 - Database](./docs/architecture/05-database.md)** — Schema design, indexing, migrations, isolation patterns

### Architecture Decision Records (ADRs)

- **[001 - Hexagonal Architecture](./docs/decisions/001-hexagonal-architecture.md)** — Why this architectural style
- **[002 - Database & Storage Separation](./docs/decisions/002-database-and-storage-separation.md)** — Why PostgreSQL and MinIO
- **[003 - MinIO for Object Storage](./docs/decisions/003-minio-object-storage.md)** — Why S3-compatible storage
- **[004 - Multi-Tenancy Strategy](./docs/decisions/004-multi-tenancy-strategy.md)** — How user isolation works ⭐

**Most important for v1:** Start with [ADR 004](./docs/decisions/004-multi-tenancy-strategy.md) to understand multi-user design.

---

## Development Workflow

### Adding a New Feature

1. **Start from domain** — Define domain entities and business rules
2. **Define ports** — Write interface contracts
3. **Implement adapter** — Add PostgreSQL/MinIO implementation
4. **Write tests** — Integration tests verify isolation
5. **Update docs** — Document design decisions

**Example:** Adding user quotas

```go
// 1. Domain: Add quota field
type User struct {
    StorageQuota  int64 // bytes
    StorageUsed   int64 // bytes
}

// 2. Port: Define interface
type UserRepository interface {
    GetByID(ctx, userID uuid.UUID) (*User, error)
    UpdateStorageUsed(ctx, userID uuid.UUID, delta int64) error
}

// 3. Adapter: Implement in PostgreSQL
func (r *UserRepository) UpdateStorageUsed(ctx, userID uuid.UUID, delta int64) error {
    _, err := r.conn.Exec(ctx,
        `UPDATE users SET storage_used = storage_used + $1 WHERE id = $2`,
        delta, userID,
    )
    return err
}

// 4. Tests: Verify isolation
func TestStorageQuotaEnforcement(t *testing.T) { ... }
```

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# With coverage report
go test -cover ./...

# Specific test
go test -run TestAssetIsolation ./internal/adapters/postgres
```

### Database Migrations

```bash
# Create a new migration
goose create add_feature_x sql

# This creates: migrations/NNNN_add_feature_x.sql

# Apply it
goose up

# Rollback
goose down
```

---

## Contributing

Contributions are welcome! Before opening a pull request:

1. **Ensure all tests pass** — `go test ./...`
2. **Update documentation** — Especially if changing architecture or schema
3. **Follow conventions** — See [04 - Guidelines](./docs/architecture/04-guidelines.md)
4. **Enforce isolation** — All data access must include user_id
5. **Write integration tests** — Verify multi-user isolation

### PR Checklist

- [ ] Tests pass locally (`go test ./...`)
- [ ] Code follows Go conventions (gofmt, golint)
- [ ] Documentation updated for schema/API changes
- [ ] Multi-user isolation is maintained (if applicable)
- [ ] Integration tests cover new functionality

---

## Roadmap

### v1.0 (Target: Sept 2026)

**Goal:** Complete REST API with multi-user auth and isolation

- [x] Domain models and ports
- [x] PostgreSQL adapter (multi-tenant)
- [x] MinIO object storage
- [ ] Application layer (use cases)
- [ ] HTTP API with routes
- [ ] Bearer token authentication
- [ ] Redis job queue
- [ ] Background worker
- [ ] Asset processing pipeline
- [ ] End-to-end tests

**Status:** Core infrastructure complete; API and worker in progress

### v1.1 (Target: Oct 2026)

- [ ] Admin API (user management, quotas)
- [ ] Webhook notifications
- [ ] Activity audit log
- [ ] API rate limiting

### v2.0 (Target: 2027)

- [ ] Organizations/workspaces
- [ ] Role-based access control
- [ ] Fine-grained API keys
- [ ] File browser/folders
- [ ] Usage metrics and billing
- [ ] Real-time updates

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contact & Support

- **GitHub Issues:** Report bugs or request features
- **Discussions:** Ask questions and share ideas
- **Email:** joshua@example.com (for questions)

---

## Acknowledgments

This project was built as a learning exercise in:
- Hexagonal Architecture
- Multi-tenant system design
- Go backend development
- Cloud storage integration

Special thanks to:
- Hexagonal Architecture community
- PostgreSQL documentation
- MinIO team

---

## Quick Links

| Resource | Link |
|----------|------|
| **Architecture** | [docs/architecture/README.md](./docs/architecture/README.md) |
| **Multi-Tenancy ADR** | [docs/decisions/004-multi-tenancy-strategy.md](./docs/decisions/004-multi-tenancy-strategy.md) |
| **Database Schema** | [docs/architecture/05-database.md](./docs/architecture/05-database.md) |
| **GitHub** | [joshua-sajeev/tessera](https://github.com/joshua-sajeev/tessera) |
| **Issues** | [GitHub Issues](https://github.com/joshua-sajeev/tessera/issues) |

---


## Contributing

Contributions are welcome.

Before opening a pull request:

* Ensure all tests pass.
* Update documentation when architecture or behavior changes.
* Follow the project's architectural guidelines.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

**Last Updated:** 2026-08-20  
**Version:** v1.0 (in progress)
---
