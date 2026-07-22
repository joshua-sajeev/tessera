# Repository Structure

The following structure reflects the current implementation of Tessera.

```text
tessera/
├── cmd/                                  # Application entrypoints
│   ├── api/
│   │   └── main.go                       # API application entrypoint (Planned)
│   └── worker/
│       └── main.go                       # Worker application entrypoint (Planned)
│
├── internal/
│   ├── domain/                           # Core business models
│   │   ├── asset/
│   │   │   ├── asset.go                  # Asset entity
│   │   │   ├── variant.go                # Asset variant entity
│   │   │   ├── status.go                 # Asset status definitions
│   │   │   └── errors.go                 # Domain errors
│   │   │
│   │   └── processing/
│   │       ├── job.go                    # Processing job entity
│   │       ├── status.go                 # Processing job status
│   │       └── errors.go                 # Domain errors
│   │
│   ├── ports/                            # Infrastructure contracts
│   │   ├── asset_repository.go           # Asset persistence interface
│   │   ├── processing_repository.go      # Processing persistence interface
│   │   ├── storage.go                    # Object storage contract
│   │   └── queue.go                      # Queue contract
│   │
│   ├── adapters/
│   │   └── postgres/                     # PostgreSQL implementations
│   │       ├── db.go                     # Database connection helper
│   │       ├── asset_repository.go
│   │       ├── processing_repository.go
│   │       ├── asset_repository_test.go
│   │       ├── processing_repository_test.go
│   │       └── main_test.go              # Shared integration test setup
│   │
│   └── config/
│       └── config.go                     # Application configuration
│
├── migrations/
│   └── 0001_create_schema.sql            # Initial PostgreSQL schema
│
├── deployments/
│   └── docker-compose.dev.yml            # Local development services
│
├── docs/
│   ├── architecture/
│   │   ├── README.md
│   │   ├── 00-overview.md
│   │   ├── 01-layers.md
│   │   ├── 02-flows.md
│   │   ├── 03-structure.md
│   │   ├── 04-guidelines.md
│   │   ├── 05-database.md
│   │   └── README.md
│   │
│   └── decisions/
│       └── 001-hexagonal-arch.md
│
├── Makefile
├── README.md
├── go.mod
├── go.sum
└── LICENSE
```

## Folder Breakdown

| Folder                        | Purpose                                                                                                      | Status |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------ | ------ |
| `cmd/`                        | Entry points for the API and worker binaries.                                                                | 🔄 Planned |
| `internal/domain/`            | Core business entities, status types, and domain errors. This layer contains no infrastructure dependencies. | ✅ Implemented |
| `internal/ports/`             | Interfaces that define the application's persistence, storage, and queueing requirements.                    | ✅ Implemented |
| `internal/adapters/postgres/` | PostgreSQL implementations of repository ports using `pgx/v5`, along with integration tests.                 | ✅ Implemented |
| `internal/config/`            | Application configuration loading.                                                                           | ✅ Implemented |
| `migrations/`                 | Goose database migrations that define and evolve the PostgreSQL schema.                                      | ✅ Implemented |
| `deployments/`                | Local development infrastructure, including Docker Compose.                                                  | ✅ Implemented |
| `docs/architecture/`          | Architecture documentation, design decisions, repository layout, and database design.                        | ✅ Implemented |
| `docs/decisions/`             | Architecture Decision Records (ADRs).                                                                        | ✅ Implemented |

---

# Current Architecture

The repository currently contains the foundational layers of the Hexagonal Architecture.

```text
        External Interfaces
               │
               ▼
        Application (Planned)
               │
               ▼
            Ports
               │
               ▼
           Domain
               ▲
               │
        PostgreSQL Adapter (Implemented)
```

**Currently Implemented:**
- Domain models and entities
- Repository ports (interfaces)
- PostgreSQL adapter with full CRUD operations
- Database schema and migrations
- Integration tests

**Future Milestones:**
- HTTP API adapter
- MinIO object storage adapter
- Redis queue adapter
- Application use cases and orchestration
- Worker implementation

---

# Database Layout

The PostgreSQL schema currently consists of three related tables.

```text
assets
   │
   ├──────────────┐
   │              │
   ▼              ▼
processing_jobs   asset_variants
```

* **assets** - Stores metadata for uploaded assets (name, size, content type, storage path, status)
* **processing_jobs** - Tracks asynchronous processing state for each asset
* **asset_variants** - Stores metadata for generated asset variants (thumbnails, optimized versions)

The schema is managed through Goose migrations located in the `migrations/` directory.

---

# Integration Tests

Repository implementations are verified using integration tests against a real PostgreSQL instance.

Current test coverage includes:

* Asset repository CRUD operations
* Processing repository CRUD operations
* Error handling and edge cases
* Shared test database setup and cleanup

Running the integration test suite requires a running PostgreSQL instance.

---

## Navigation

**Previous:** [02 - Flows](02-flows.md)

**Next:** [04 - Guidelines](04-guidelines.md)
