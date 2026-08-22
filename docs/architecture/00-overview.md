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
- **Repository ports** with user_id enforcement
- **PostgreSQL persistence adapters** with multi-tenant isolation
- **Goose database migrations** (schema with user_id foreign keys)
- **PostgreSQL integration tests**
- **MinIO object storage adapter** with integration tests
- **Docker Compose development environment**
- **User domain model and authentication foundation**
- **Multi-user data isolation** in database schema, ports, and adapters

### 📋 Planned

- **Application layer** (use cases and business logic)
- **HTTP API** (REST endpoints with Bearer token auth middleware)
- **Redis job queue** (async processing with per-user fairness)
- **Background worker** (process jobs asynchronously)
- **Asset processing pipeline** (variant generation and optimization)

---

## Version 1 Features & Capabilities

### Authentication & Authorization
1. **User Authentication** — API key-based auth with Bearer tokens
2. **Data Isolation** — Users only access their own assets, jobs, and variants
3. **Storage Quotas** — Per-user storage limits enforced at upload time

### Asset Management
4. **Upload Asset** — Accept asset uploads via HTTP with multi-user auth
5. **Store Original** — Persist original asset to MinIO object storage
6. **Save Metadata** — Record asset metadata in PostgreSQL with user_id
7. **Download Asset** — Serve assets and variants (with auth and ownership checks)

### Processing Pipeline
8. **Create Processing Job** — Queue asset processing work with user fairness
9. **Worker Processes Asset** — Async job processing per-user fairness (planned)
10. **Generate Variants** — Create optimized copies (thumbnails, previews, etc.)
11. **Update Status** — Track processing progress in job state machine
12. **Return 202 Accepted** — Immediate client response for async operations

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
func (r *AssetRepository) GetByID(ctx context.Context, assetID uuid.UUID, userID uuid.UUID) (*Asset, error) {
    // Executes: SELECT ... FROM assets WHERE id = $1 AND user_id = $2
}

// WRONG: This signature cannot exist; it would fail to compile
// func (r *AssetRepository) GetByID(ctx context.Context, assetID uuid.UUID) (*Asset, error) { ... }
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

### ✅ Implemented (v1 Foundation)
- [x] Domain models (Asset, ProcessingJob, AssetVariant, User)
- [x] Repository ports with user_id enforcement
- [x] PostgreSQL persistence adapters
- [x] Goose migrations (multi-user schema)
- [x] Integration tests (PostgreSQL)
- [x] MinIO object storage adapter
- [x] MinIO integration tests
- [x] User domain model
- [x] Multi-user data isolation

### 📋 Planned (v1 Completion)
- [ ] Application layer (use cases)
- [ ] HTTP API (REST endpoints)
- [ ] Bearer token authentication middleware
- [ ] Redis job queue
- [ ] Background worker process
- [ ] Asset processing pipeline (variant generation)
- [ ] Per-user storage quota enforcement
- [ ] End-to-end integration tests

### 🔮 Future (v2+)
- [ ] Organizations/workspace support
- [ ] Role-based access control (RBAC)
- [ ] Fine-grained API key permissions
- [ ] Webhook notifications
- [ ] Folder/asset organization
- [ ] Audit logging and compliance
- [ ] Usage metrics and billing
- [ ] WebSocket real-time job updates

---

## Navigation

**Continue Reading:**
- Next: [01 - Layers](01-layers.md) — Understand the responsibility of each architectural layer
- Related: [ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md) — Deep dive into user isolation design
- Related: [05 - Database](05-database.md) — Schema design, indexing, and migration strategy

Return to [Architecture Index](README.md) for a complete overview.
