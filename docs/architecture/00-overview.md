# Overview

Tessera is a backend service that stores uploaded assets, processes them asynchronously, and serves optimized variants through a REST API. It is built with **Hexagonal Architecture** and includes **multi-user isolation from the ground up**.

## What Tessera Is

- An asset processing backend
- Async-first architecture with job queuing
- Modular and testable (Hexagonal Architecture)
- Infrastructure-agnostic (swappable persistence, storage, and queue implementations)
- **Multi-user capable with built-in API key authentication and data isolation** (v1)

## What Tessera Is NOT

- An image editor
- A frontend application
- A CDN
- An authentication provider (though it manages API keys for users)

---

## Current Status

The project has implemented the foundational persistence, storage, and multi-user infrastructure layers:

### ✅ Implemented

- **Domain models** (Asset, ProcessingJob, AssetVariant, User)
- **Domain API key layer** (generation, Argon2id hashing, validation)
- **Repository ports** with user_id enforcement
- **PostgreSQL persistence adapters** with multi-tenant isolation
- **PostgreSQL user repository** with API key storage and retrieval
- **Goose database migrations** (schema with user_id foreign keys, users table)
- **PostgreSQL integration tests** (isolation verification)
- **MinIO object storage adapter** with integration tests
- **Docker Compose development environment**
- **User domain model** with status management (Active, Suspended, Deleted)
- **Authentication foundation** (Authenticator port + PostgreSQL adapter)
- **Application layer services** (auth service, user service)
- **HTTP API** with user management endpoints
- **Bearer token authentication middleware** (RequireAuth wrapper)
- **User context injection** across HTTP handlers
- **Multi-user data isolation** at all layers (domain, ports, adapters, HTTP)

### 📋 Planned (v1 Completion)

- **Asset management endpoints** (upload, download, delete, list)
- **Asset storage and metadata** (MinIO + PostgreSQL)
- **Processing job queue** (Redis-based with per-user fairness)
- **Background worker** (async job processing)
- **Asset processing pipeline** (variant generation and optimization)
- **Job status tracking** and monitoring
- **Per-user storage quotas** enforcement
- **End-to-end integration tests**

---

## Version 1 Features & Capabilities

### Authentication & Authorization (✅ Implemented)
1. **API Key Generation** — Secure key generation with cryptographic randomness
2. **API Key Storage** — Argon2id hashing of secrets, never stored in plaintext
3. **User Authentication** — API key verification against stored hashes
4. **Bearer Token Support** — Standard `Authorization: Bearer token_id` header parsing
5. **User Status Management** — Active, Suspended, Deleted states with enforcement
6. **Request Context Injection** — Authenticated user available throughout request lifecycle
7. **Ownership Verification** — HTTP handlers enforce user can only access their own resources

### Data Isolation (✅ Implemented)
8. **Multi-Tenant Database** — All tables have user_id foreign keys
9. **Repository Enforcement** — All data access methods require user_id parameter
10. **Composite Indices** — Fast queries on (user_id, status) combinations
11. **Database-Level Constraints** — Foreign keys prevent orphaning of user data
12. **No Cross-User Leakage** — WHERE clauses always include user_id filters

### Asset Management (📋 In Progress)
13. **Upload Asset** — Accept file uploads via multipart/form-data with auth
14. **Store Original** — Persist asset to MinIO with per-user bucket organization
15. **Save Metadata** — Record asset info (name, size, mime type, created_at) in PostgreSQL
16. **Download Asset** — Serve assets with access control verification
17. **List Assets** — Query user's assets with pagination and filtering
18. **Delete Asset** — Remove asset from storage and database

### Processing Pipeline (📋 Planned)
19. **Create Processing Job** — Queue async variant generation with user fairness
20. **Job Queue** — Redis-based queue ensuring fair processing across users
21. **Worker Processing** — Background service executing jobs asynchronously
22. **Generate Variants** — Create thumbnails, previews, and optimized versions
23. **Status Tracking** — Queued → Processing → Complete/Failed state transitions
24. **Async Response** — Return 202 Accepted immediately, deliver results asynchronously

---

## Architectural Style: Hexagonal Architecture

Tessera follows **Ports and Adapters** (Hexagonal) architecture to isolate business logic from infrastructure.

### Why Hexagonal?

| Benefit | Why It Matters |
|---------|---|
| **Isolated Business Logic** | Domain logic is independent of frameworks, databases, and storage systems |
| **Easier Testing** | Core domain and use cases are testable without infrastructure |
| **Infrastructure Replaceable** | Swap PostgreSQL for MongoDB, MinIO for S3, etc. without changing domain logic |
| **Learning Objective** | Explore clean architecture principles and design patterns |
| **Multi-Tenancy Ready** | Auth and isolation are baked in from the start, not retrofitted |

### Core Principle

**Dependencies point inward.** The domain knows nothing about HTTP, databases, queues, storage systems, or external frameworks.

```
External World
  (HTTP, PostgreSQL, MinIO, Redis, Auth)
          ↓
    Adapters (Port implementations)
          ↓
  Application (Use Cases)
          ↓
    Domain (Core Business Logic)
          ↑
    Ports (Interfaces/Contracts)
```

---

## Multi-User Architecture

Tessera v1 implements **database-layer user isolation** with enforcement at every level:

### Database Layer
- All tables have `user_id` foreign keys to the `USERS` table
- Composite indices on `(user_id, status)` for fast user-scoped queries
- Foreign key constraints prevent orphaning of user data

### Auth Layer
- **Authenticator port**: Interface for API key verification
- **Hashed API keys**: API keys are stored as Argon2id hashes.
- **User lookup**: Token validation maps Bearer token → User object

### Repository Layer
- **Explicit user_id parameter**: All repository methods require `user_id` in their signature
- **Forced isolation**: Developer cannot query assets without providing user_id
- **Database-enforced**: WHERE clauses include both `user_id` and resource `id`

### API Layer (Planned)
- **Bearer token middleware**: Extracts and validates API key on every request
- **Request context**: Authenticates user and injects into request context
- **Ownership checks**: HTTP handlers verify user owns requested resource

### Example Repository Method

```go
// CORRECT: User_id is explicit and enforced at compile time
func (r *AssetRepository) Get(ctx context.Context, assetID uuid.UUID, userID uuid.UUID) (*Asset, error) {
    // Executes: SELECT ... FROM assets WHERE id = $1 AND user_id = $2
}

// WRONG: This signature cannot exist; it would fail to compile
// func (r *AssetRepository) Get(ctx context.Context, assetID uuid.UUID) (*Asset, error) { ... }
```

For the detailed design rationale and consequences, see **[ADR 004: Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md)**.

---

## Tech Stack

| Category       | Technology      |
|---|---|
| Language       | Go 1.21+        |
| Database       | PostgreSQL 15+  |
| Database Driver| pgx/v5          |
| Migrations     | Goose           |
| Object Storage | MinIO (S3-compatible) |
| Testing        | Go `testing` pkg |
| Containers     | Docker Compose  |

---

## Design Documents

The architecture is documented in `docs/architecture/` and `docs/decisions/`:

### Architecture Guides
- **00 - Overview** — Project goals and architectural style (this document)
- **01 - Layers** — Responsibilities of Domain, Ports, Adapters, and Application layers
- **02 - Flows** — Request/response sequences and processing pipelines
- **03 - Structure** — Repository layout and module organization
- **04 - Guidelines** — Development conventions and patterns
- **05 - Database** — Schema design, indexing strategy, and migration approach

### Architecture Decision Records (ADRs)
- **ADR 001** — Hexagonal Architecture selection
- **ADR 002** — PostgreSQL and object storage separation
- **ADR 003** — MinIO for object storage
- **ADR 004** — Multi-Tenancy Strategy (user isolation)

---

## Roadmap

### ✅ Implemented (v1 Foundation + Auth)
- [x] Domain models (Asset, ProcessingJob, AssetVariant, User)
- [x] API key generation and Argon2id hashing
- [x] Repository ports with user_id enforcement
- [x] PostgreSQL persistence adapters (multi-tenant)
- [x] PostgreSQL user repository with API key support
- [x] Goose migrations (users table, multi-user schema, indices)
- [x] PostgreSQL integration tests (isolation verification)
- [x] MinIO object storage adapter with integration tests
- [x] User domain model with status management
- [x] Authenticator port and PostgreSQL adapter
- [x] Application layer services (auth, user management)
- [x] HTTP API with user management endpoints
- [x] Bearer token authentication middleware
- [x] User context injection across handlers
- [x] Multi-user data isolation (all layers)

### 📋 Planned (v1 Completion)
- [ ] Asset HTTP endpoints (GET, POST, DELETE, LIST)
- [ ] Asset upload with multipart/form-data
- [ ] Asset storage to MinIO with metadata
- [ ] Asset access control verification
- [ ] Redis job queue with user fairness
- [ ] Background worker process
- [ ] Asset processing pipeline (variant generation)
- [ ] Per-user storage quota enforcement
- [ ] Processing job endpoints and status tracking
- [ ] End-to-end integration tests

### 🔮 Future (v2+)
- [ ] Organizations/workspace support with multi-level isolation
- [ ] Role-based access control (RBAC) within organizations
- [ ] Fine-grained API key permissions (scopes)
- [ ] Webhook notifications for job completion
- [ ] Folder/asset organization and tagging
- [ ] Audit logging and compliance
- [ ] Usage metrics and billing integration
- [ ] WebSocket real-time job updates
- [ ] Advanced image processing (filters, effects, formats)

---

## Navigation

**Continue Reading:**
- Next: [01 - Layers](01-layers.md) — Understand the responsibility of each architectural layer
- Related: [ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md) — Deep dive into user isolation design
- Related: [05 - Database](05-database.md) — Schema design, indexing, and migration strategy

Return to [Architecture Index](README.md) for a complete overview.
