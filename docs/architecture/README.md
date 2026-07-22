# Tessera Architecture

This directory contains the design patterns, system flows, and technical guidelines for the project.

## Documentation Index

- [**00-Overview**](00-overview.md): High-level goals, current status, and hexagonal architecture concepts.
- [**01-Layers**](01-layers.md): Detailed breakdown of the Domain, Ports, Application, and Adapter layers.
- [**02-Flows**](02-flows.md): Sequence diagrams and system request flows.
- [**03-Structure**](03-structure.md): Repository layout, folder responsibilities, and current architecture.
- [**04-Guidelines**](04-guidelines.md): Development workflow, layer responsibilities, and naming conventions.
- [**05-Database**](05-database.md): Schema design, entity relationships, and database strategy.

---

## Architecture Decision Records (ADRs)

- [001-Hexagonal-Arch](../decisions/001-hexagonal-arch.md)

---

## Current Status

**Phase:** Persistence Layer (Stable) ✅

### What's Implemented

- ✅ Domain models (Asset, ProcessingJob, AssetVariant)
- ✅ Repository ports and interfaces
- ✅ PostgreSQL adapters with full test coverage
- ✅ Goose database migrations
- ✅ Integration tests

### What's Planned

- 🔄 Application use cases and orchestration (v1)
- 🔄 HTTP API adapter (v1)
- 🔄 MinIO object storage adapter (v1)
- 🔄 Redis queue adapter (v1)
- 🔄 Worker implementation (v1)

---

## Architectural Overview

Tessera uses **Hexagonal Architecture** to keep business logic independent from infrastructure:

```
External World (HTTP, Database, Storage, Queues)
              ↓
          Adapters
              ↓
        Application (Use Cases)
              ↓
           Domain
              ↑
        Ports (Interfaces)
```

**Benefits:**
- Core business logic is testable without infrastructure
- Infrastructure components are swappable
- Dependencies point inward
- Each layer has a single responsibility

---

## Quick Start

### Reading Guide

**New to the project?**
1. Start with [00 - Overview](00-overview.md)
2. Read [01 - Layers](01-layers.md) to understand components
3. Review [03 - Structure](03-structure.md) for repository layout
4. Follow [04 - Guidelines](04-guidelines.md) when implementing

**Want to understand request flow?**
- See [02 - Flows](02-flows.md) for diagrams and sequences

**Working with the database?**
- Read [05 - Database](05-database.md) for schema and design rationale

---

## Key Concepts

### Ports

Interfaces that define external dependencies:
- `AssetRepository` - Asset persistence
- `ProcessingRepository` - Job tracking
- `Storage` - Object storage (MinIO)
- `Queue` - Job queue (Redis)

### Adapters

Concrete implementations of ports:
- **PostgreSQL Adapter** ✅ - Implements repository ports
- **HTTP Adapter** 🔄 - HTTP request handling (planned)
- **MinIO Adapter** 🔄 - Object storage (planned)
- **Redis Adapter** 🔄 - Job queue (planned)

### Domain

Core business logic with no external dependencies:
- `Asset` - Uploaded asset entity
- `ProcessingJob` - Async job entity
- `AssetVariant` - Processed variant entity
- Business rules and validations

### Application (Planned)

Orchestrates use cases using domain logic and ports:
- `UploadAsset` - Handle asset upload
- `ProcessAsset` - Process variants
- `DownloadAsset` - Serve processed assets

---

## Development Workflow

When adding a new feature, follow this workflow (see [04 - Guidelines](04-guidelines.md) for details):

```
1. Define domain model          → internal/domain/
2. Create port interface        → internal/ports/
3. Implement PostgreSQL adapter → internal/adapters/postgres/
4. Add database migration       → migrations/
5. Write integration tests      → _test.go files
6. Update documentation        → docs/architecture/
```

---

## Local Development

```bash
make dev-setup      # Initialize development environment
make migrate        # Run database migrations
make test           # Run all tests
make test-coverage  # View test coverage
```

---

## Questions?

Each documentation file has a **Navigation** section for moving between topics.

For implementation examples, see `internal/adapters/postgres/` which demonstrates all architectural concepts in practice.
