# Repository Structure

```text
tessera/
├── cmd/                                  # Application entrypoints
│   ├── api/
│   │   └── main.go                       # Starts the HTTP API
│   └── worker/
│       └── main.go                       # Starts the background worker
│
├── internal/                             # Private application code
│   ├── domain/                           # Business entities and rules
│   │   ├── asset/
│   │   │   ├── asset.go                  # Asset entity
│   │   │   └── errors.go                 # Domain errors
│   │   └── processing/
│   │       ├── job.go                    # Processing job entity
│   │       ├── variant.go                # Asset variant entity
│   │       └── status.go                 # Job status definitions
│   │
│   ├── application/                      # Application use cases
│   │   ├── upload.go                     # Upload asset workflow
│   │   ├── process.go                    # Process asset workflow
│   │   ├── download.go                   # Download asset workflow
│   │   └── errors.go                     # Application errors
│   │
│   ├── ports/                            # Interfaces used by the application
│   │   ├── repository.go                 # Persistence contracts
│   │   ├── storage.go                    # Object storage contracts
│   │   └── queue.go                      # Job queue contracts
│   │
│   ├── adapters/                         # Infrastructure implementations
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── upload.go             # Upload endpoint
│   │   │   │   ├── download.go           # Download endpoint
│   │   │   │   └── errors.go             # HTTP error mapping
│   │   │   ├── middleware.go             # HTTP middleware
│   │   │   └── router.go                 # Route registration
│   │   │
│   │   ├── postgres/
│   │   │   └── repository.go             # PostgreSQL implementation
│   │   │
│   │   ├── minio/
│   │   │   └── storage.go                # MinIO implementation
│   │   │
│   │   └── redis/
│   │       └── queue.go                  # Redis implementation
│   │
│   └── config/
│       └── config.go                     # Loads application configuration
│
├── migrations/                           # Goose SQL migrations
│
├── deployments/                          # Deployment resources
│   ├── docker/
│   │   ├── Dockerfile.api
│   │   ├── Dockerfile.worker
│   │   └── docker-compose.yml
│   └── k8s/                              # Kubernetes manifests (planned)
│
├── architecture/                         # Architecture documentation
│   ├── README.md
│   ├── 00-overview.md
│   ├── 01-layers.md
│   ├── 02-flows.md
│   ├── 03-structure.md
│   ├── 04-guidelines.md
│   └── decisions/
│       └── 001-hexagonal-arch.md
│
├── docs/                                 # User and operational documentation
│   ├── api.md
│   ├── development.md
│   └── deployment.md
│
├── scripts/                              # Development automation
│   ├── dev-setup.sh
│   ├── migrate.sh
│   └── generate-mocks.sh
│
├── .env.example                          # Example environment variables
├── Makefile                              # Common development commands
├── go.mod                                # Go module definition
├── go.sum                                # Dependency checksums
├── README.md                             # Project overview
└── LICENSE                               # Project license
```

## Folder Breakdown

| Folder | Purpose |
|---------|---------|
| `cmd/` | Executable applications. Each subdirectory builds a separate binary. |
| `internal/` | Private application code following the Hexagonal Architecture. |
| `internal/domain/` | Core business entities and business rules with no infrastructure dependencies. |
| `internal/application/` | Use cases that orchestrate the domain through ports. |
| `internal/ports/` | Interfaces defining the application's required external capabilities. |
| `internal/adapters/` | Infrastructure implementations of the ports (HTTP, PostgreSQL, MinIO, Redis). |
| `internal/config/` | Application configuration loading and validation. |
| `migrations/` | Database schema migrations managed by Goose. |
| `deployments/` | Docker and Kubernetes deployment resources. |
| `architecture/` | Architecture documentation, design decisions, and ADRs. |
| `docs/` | API reference, development guides, and operational documentation. |
| `scripts/` | Helper scripts for local development and automation. |

---

## MinIO Object Layout

Assets are stored in MinIO using object paths within a single bucket.

```text
tessera-assets/
├── originals/
│   ├── asset-001/
│   │   └── image.jpg
│   └── asset-002/
│       └── document.pdf
│
└── variants/
    ├── asset-001/
    │   ├── thumbnail.jpg
    │   ├── optimized.jpg
    │   └── webp.jpg
    └── asset-002/
        └── preview.pdf
```

The application interacts with object storage only through the `Storage` port. The MinIO adapter provides the concrete implementation.

---

## Navigation

**Previous:** [02 - Flows](02-flows.md)

**Next:** [04 - Guidelines](04-guidelines.md)

Learn the architectural rules and development conventions followed throughout the project.
