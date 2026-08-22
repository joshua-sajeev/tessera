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

- [001-Hexagonal-Arch](../decisions/001-hexagonal-arch.md)
- [002-MinIO-Storage](../decisions/002-minio-storage.md)
- [003-Separate-Storage](../decisions/003-separate-storage.md)
- [004-Multi-Tenancy-Strategy](../decisions/004-multi-tenancy-strategy.md) - User isolation, API key auth, quota management

---

## Current Status

Phase: Persistence Layer & Core Storage (Stable)

### What's Implemented

- Domain models (Asset, ProcessingJob, AssetVariant)
- Repository ports and interfaces (AssetRepository, ProcessingRepository)
- PostgreSQL adapters with initial schema
- MinIO object storage adapter with full test coverage
- Integration tests (PostgreSQL, MinIO)

### What's Planned (v1)

- **Multi-User Foundation (Next Step):**
  - User domain model (User entity)
  - Database migrations for users and user_id columns
  - Repository ports & adapters with user_id isolation
  - Authenticator port & API key validation adapter
- Application use cases and orchestration (UploadAsset, ProcessAsset)
- HTTP API adapter with Bearer token auth
- Redis queue adapter
- Worker implementation
- Storage quota enforcement

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

New to the project?
1. Start with [00 - Overview](00-overview.md)
2. Read [ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md) to understand user isolation
3. Review [01 - Layers](01-layers.md) to understand components
4. Review [03 - Structure](03-structure.md) for repository layout
5. Follow [04 - Guidelines](04-guidelines.md) when implementing

Want to understand request flow?
- See [02 - Flows](02-flows.md) for diagrams and sequences

Working with the database?
- Read [05 - Database](05-database.md) for schema, multi-user isolation, and design rationale

Understanding multi-user design?
- Read [ADR 004 - Multi-Tenancy Strategy](../decisions/004-multi-tenancy-strategy.md)

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
- PostgreSQL Adapter (stable) - Implements repository ports with user_id enforcement
- Auth Adapter (stable) - API key hashing and user lookup
- HTTP Adapter (planned) - HTTP request handling and Bearer token extraction
- MinIO Adapter (stable) - Object storage implementation
- Redis Adapter (planned) - Job queue implementation

### Domain

Core business logic with no external dependencies:
- User - Account entity with API key and quota
- Asset - Uploaded asset entity (linked to user)
- ProcessingJob - Async job entity (linked to user)
- AssetVariant - Processed variant entity
- Business rules and validations (user isolation)

### Application (Planned)

Orchestrates use cases using domain logic and ports:
- UploadAsset - Handle asset upload with user isolation
- ProcessAsset - Process variants with user fairness
- DownloadAsset - Serve processed assets with auth
- CreateUser - Provision new user with API key

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
make dev-setup      # Initialize development environment
make goose-up       # Run database migrations (creates users table)
make test           # Run all tests (including isolation verification)
make test-coverage  # View test coverage
```

---

## Multi-User Testing

When testing new features:

1. Create User A with api_key_hash_a
2. Create User B with api_key_hash_b
3. User A creates asset - verify only User A can query it
4. User A queries asset - User B still can't see it
5. Verify repo methods require user_id parameter

See integration tests in internal/adapters/postgres/ for examples.

---

## Questions?

Each documentation file has a Navigation section for moving between topics.

For implementation examples, see internal/adapters/postgres/ which demonstrates all architectural concepts including multi-user isolation.
