# Tessera Architecture

This directory contains the design patterns, system flows, and technical guidelines for the project.

## Documentation Index

- [00-Overview](00-overview.md): High-level goals, multi-user design, and hexagonal architecture concepts.
- [01-Layers](01-layers.md): Detailed breakdown of the Domain, Ports, Application, and Adapter layers.
- [02-Flows](02-flows.md): Sequence diagrams and system request flows.
- [03-Structure](03-structure.md): Repository layout, folder responsibilities, and current architecture.
- [04-Guidelines](04-guidelines.md): Development workflow, layer responsibilities, and naming conventions.
- [05-Database](05-database.md): Schema design, multi-user isolation, entity relationships, and database strategy.

---

## Architecture Decision Records (ADRs)

- [001-Hexagonal-Arch](../decisions/001-hexagonal-arch.md) — Why hexagonal architecture
- [002-MinIO-Storage](../decisions/002-minio-storage.md) — Database vs object storage separation
- [003-Separate-Storage](../decisions/003-separate-storage.md) — Why MinIO for S3-compatible storage
- [004-Multi-Tenancy-Strategy](../decisions/004-multi-tenancy-strategy.md) — ⭐ User isolation and multi-tenant design
- [005-API-Key-Auth](../decisions/005-api-key-auth.md) — Why API keys over JWT
- [006-HTTP-Auth-Middleware](../decisions/006-http-auth-middleware.md) — Bearer token middleware implementation

---

## Current Status

Phase: Authentication & User Management (Stable) | Asset Management & Processing (In Progress)

### ✅ Implemented (v1 Foundation)

**Persistence & Storage:**
- Domain models (Asset, ProcessingJob, AssetVariant, User)
- Repository ports and interfaces (AssetRepository, ProcessingRepository, UserRepository)
- PostgreSQL adapters with multi-tenant schema
- MinIO object storage adapter with full test coverage
- PostgreSQL integration tests (including isolation verification)
- Goose migrations (schema, user_id foreign keys, indices)

**Authentication & Multi-User (NEW):**
- User domain model with status management (Active, Suspended, Deleted)
- API key generation with cryptographic randomness
- API key hashing with Argon2id (never stored plaintext)
- Authenticator port interface
- PostgreSQL authenticator adapter with user lookup
- Application services: AuthService, UserService
- HTTP authentication middleware with Bearer token support
- User context injection across request handlers
- User management endpoints (create, get, list, update status)
- Complete multi-user isolation at all layers (domain, ports, adapters, HTTP)

### 📋 Planned (v1 Completion)

**Asset Management:**
- Asset upload endpoint with multipart/form-data support
- Asset storage to MinIO with per-user organization
- Asset metadata persistence and retrieval
- Asset access control verification
- Asset deletion with cleanup
- Asset listing with pagination and filtering

**Processing Pipeline:**
- Redis job queue with per-user fairness
- Background worker process
- Asset processing pipeline (variant generation, thumbnails, etc.)
- Processing job endpoints and status tracking
- Job execution monitoring and error handling

**Testing & Polish:**
- End-to-end integration tests
- Per-user storage quota enforcement
- Graceful error handling and input validation
- Performance optimization

---

## Architectural Overview

Tessera uses Hexagonal Architecture to keep business logic independent from infrastructure:

```
External World (HTTP, Database, Storage, Queues, Auth)
               Down
           Adapters
               Down
         Application (Use Cases)
               Down
            Domain
               Up
         Ports (Interfaces)
```

Benefits:
- Core business logic is testable without infrastructure
- Infrastructure components are swappable
- Dependencies point inward
- Each layer has a single responsibility
- Multi-user isolation enforced at all layers

---

## Quick Start

### Reading Guide

**New to the project?**
1. Start with [00 - Overview](00-overview.md) — Project goals and architecture
2. Read [ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md) — User isolation ⭐
3. Read [ADR 005 - API Key Auth](../decisions/005-api-key-auth.md) — Why API keys
4. Read [ADR 006 - HTTP Auth Middleware](../decisions/006-http-auth-middleware.md) — How auth works
5. Review [01 - Layers](01-layers.md) — Component responsibilities
6. Review [03 - Structure](03-structure.md) — Repository layout
7. Follow [04 - Guidelines](04-guidelines.md) — When implementing

**Want to understand request flow?**
- See [02 - Flows](02-flows.md) — Diagrams and sequences

**Working with the database?**
- Read [05 - Database](05-database.md) — Schema, isolation, migrations

**Understanding multi-user design?**
- Read [ADR 004](../decisions/004-multi-tenancy-strategy.md) + [ADR 005](../decisions/005-api-key-auth.md) + [ADR 006](../decisions/006-http-auth-middleware.md)

**Implementing authentication?**
- Study internal/adapters/postgres/user_repository.go for API key storage
- Study internal/adapters/http/middleware/auth.go for middleware pattern
- Study internal/application/auth/service.go for orchestration

---

## Key Concepts

### Ports

Interfaces that define external dependencies:
- AssetRepository - Asset persistence (enforces user_id)
- ProcessingRepository - Job tracking (enforces user_id)
- Storage - Object storage (MinIO)
- Queue - Job queue (Redis)
- Authenticator - API key verification and user lookup

### Adapters

Concrete implementations of ports:
- PostgreSQL Adapter ✅ (stable) - Repositories with user_id enforcement + user storage + API keys
- HTTP Adapter ✅ (stable) - Router, handlers, Bearer token middleware, user context injection
- Auth Adapter ✅ (stable) - API key hashing (Argon2id) and user lookup
- MinIO Adapter ✅ (stable) - Object storage implementation (S3-compatible)
- Asset Handlers 📋 (planned) - Asset upload, download, delete, list endpoints
- Redis Adapter 📋 (planned) - Job queue implementation

### Domain

Core business logic with no external dependencies:
- User - Account entity with API key (id + hash) and storage quota
- User Statuses - Active, Suspended, Deleted with enforcement
- APIKey - Generation and hashing logic (Argon2id)
- Asset - Uploaded asset entity (linked to user, with storage path)
- ProcessingJob - Async job entity (linked to user, with status tracking)
- AssetVariant - Processed variant entity (thumbnail, preview, etc.)
- Business Rules - User isolation enforced at compile time (user_id in all repository methods)

### Application (Implemented & Planned)

Orchestrates use cases using domain logic and ports:

**✅ Implemented:**
- CreateUser - Provision new user with API key generation
- AuthenticateUser - Verify API key and return user
- GetUser - Retrieve user by ID with auth verification
- ListUsers - Query users (admin-scoped, planned to restrict)
- UpdateUserStatus - Change user status (Active/Suspended/Deleted)

**📋 Planned:**
- UploadAsset - Handle asset upload with user isolation
- ProcessAsset - Process variants with per-user fairness
- DownloadAsset - Serve processed assets with auth
- DeleteAsset - Remove asset and clean up variants
- CreateProcessingJob - Queue async variant generation
- UpdateJobStatus - Track job progress

---

## Development Workflow

When adding a new feature, follow this workflow (see [04 - Guidelines](04-guidelines.md) for details):

```
1. Define domain model          -> internal/domain/
2. Create port interface        -> internal/ports/
3. Implement PostgreSQL adapter -> internal/adapters/postgres/
4. Add database migration       -> migrations/
5. Enforce user_id in queries  -> (critical for isolation)
6. Write integration tests      -> _test.go files (verify isolation)
7. Update documentation        -> docs/architecture/
```

Critical: All repository queries must include user_id to maintain isolation.

---

## Local Development

```bash
make up             # Start Docker Compose (PostgreSQL, MinIO, Redis)
make migrate        # Run database migrations
make test           # Run all tests (includes isolation verification)
make test-coverage  # View test coverage report
make run            # Start API server (port 8080)
```

### Making API Requests

```bash
# Create a user (returns API key)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com"}'
# Response: {"id":"...", "api_key":"tsr_v1_..."}

# List users (requires Bearer token)
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer tsr_v1_..."

# Get specific user (requires Bearer token + ownership)
curl -X GET http://localhost:8080/users/{user_id} \
  -H "Authorization: Bearer tsr_v1_..."
```

---

## Multi-User Testing & Isolation Verification

When testing new features, follow this workflow:

1. **Create User A** with API key → `api_key_a` (generates api_key_id + secret hash)
2. **Create User B** with API key → `api_key_b` (generates api_key_id + secret hash)
3. **User A creates resource** (asset, job, etc.) → stored with user_a_id
4. **User A queries resource** with Bearer token → resource returned
5. **User B queries same resource ID** with Bearer token → 404 Not Found or 403 Forbidden
6. **Verify repository method signatures** → All require user_id parameter
7. **Verify database queries** → All include WHERE user_id = $X

**Critical Isolation Points:**
- PostgreSQL enforces user_id in composite indices and foreign keys
- Repository methods force user_id in signature (compile-time safety)
- HTTP handlers verify authenticated user matches resource owner
- All WHERE clauses include user_id filter

See integration tests in `internal/adapters/postgres/` and `internal/adapters/http/handler/` for concrete examples.

---

## Questions?

Each documentation file has a Navigation section for moving between topics.

For implementation examples, see internal/adapters/postgres/ which demonstrates all architectural concepts including multi-user isolation.
