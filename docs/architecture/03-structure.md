# Repository Structure

```
tessera/
├── cmd/                           # Entrypoints
│   ├── api/
│   │   └── main.go               # HTTP API server
│   └── worker/
│       └── main.go               # Background worker
│
├── internal/                       # Private application code
│   ├── domain/                    # Core business logic
│   │   ├── asset/
│   │   │   ├── asset.go          # Asset entity
│   │   │   └── errors.go         # Domain errors
│   │   ├── processing/
│   │   │   ├── job.go            # ProcessingJob entity
│   │   │   ├── variant.go        # AssetVariant entity
│   │   │   └── status.go         # Job status enum
│   │   └── errors.go             # Domain error types
│   │
│   ├── application/               # Use cases & orchestration
│   │   ├── upload.go             # UploadAsset use case
│   │   ├── process.go            # ProcessAsset use case
│   │   ├── download.go           # DownloadAsset use case
│   │   └── errors.go             # Application errors
│   │
│   ├── ports/                     # Interfaces (contracts)
│   │   ├── repository.go         # Data access interface
│   │   ├── storage.go            # File storage interface
│   │   └── queue.go              # Job queue interface
│   │
│   └── adapters/                  # Infrastructure implementations
│       ├── http/
│       │   ├── handler/
│       │   │   ├── upload.go     # POST /assets
│       │   │   ├── download.go   # GET /assets/{id}/variants/{type}
│       │   │   └── errors.go     # HTTP error responses
│       │   ├── router.go         # Route definitions
│       │   └── middleware.go     # Logging, validation, etc.
│       │
│       ├── postgres/
│       │   ├── repository.go     # RepositoryPort implementation
│       │   ├── schema.go         # Database schema
│       │   └── migrations/       # SQL migrations
│       │
│       ├── minio/
│       │   ├── storage.go        # StoragePort implementation
│       │   └── config.go         # MinIO configuration
│       │
│       └── redis/
│           ├── queue.go          # QueuePort implementation
│           └── config.go         # Redis configuration
│
├── configs/                       # Configuration files
│   ├── app.yaml                  # Application config
│   ├── database.yaml             # Database config
│   ├── storage.yaml              # MinIO config
│   └── queue.yaml                # Redis config
│
├── deployments/                   # Deployment & infrastructure
│   ├── docker/
│   │   ├── Dockerfile.api        # API server image
│   │   ├── Dockerfile.worker     # Worker image
│   │   └── docker-compose.yaml   # Local development
│   │
│   └── k8s/                       # Kubernetes manifests (future)
│       ├── api-deployment.yaml
│       ├── worker-deployment.yaml
│       └── services.yaml
│
├── docs/                          # Documentation
│   ├── architecture.md            # This file
│   ├── api.md                     # API reference
│   ├── development.md             # Development guide
│   ├── deployment.md              # Deployment guide
│   └── decisions/                 # Architecture Decision Records (ADRs)
│       └── 001-hexagonal-arch.md
│
├── scripts/                       # Utility scripts
│   ├── migrate.sh                # Run database migrations
│   ├── generate-mocks.sh         # Generate test mocks
│   └── dev-setup.sh              # Local environment setup
│
├── pkg/                           # Public/shareable packages (if needed)
│   └── types/                     # Shared types (v2+)
│
├── go.mod                         # Go dependencies
├── go.sum
├── Makefile                       # Build & dev commands
└── README.md                      # Project overview
```

## Folder Breakdown

| Folder                  | Purpose                     | What It Contains                                        |
| ----------------------- | --------------------------- | ------------------------------------------------------- |
| `cmd/`                  | **Application Entrypoints** | Go `main()` functions for API and Worker                |
| `internal/`             | **Private Code**            | All business logic, never imported by external packages |
| `internal/domain/`      | **Core Logic**              | Entities, value objects, business rules                 |
| `internal/application/` | **Orchestration**           | Use cases, transaction handling                         |
| `internal/ports/`       | **Contracts**               | Interfaces that adapters must implement                 |
| `internal/adapters/`    | **Infrastructure**          | Concrete implementations (DB, Storage, Queue)           |
| `configs/`              | **Configuration**           | YAML/env config files for all services                  |
| `deployments/`          | **Deployment**              | Docker, Kubernetes, Docker Compose                      |
| `docs/`                 | **Documentation**           | Architecture, API reference, guides, ADRs               |
| `scripts/`              | **Automation**              | Migration, setup, code generation scripts               |
| `pkg/`                  | **Public Packages**         | Shared libraries (used in v2+)                          |

## MinIO Storage Structure

MinIO is object storage (like AWS S3). Images are stored in **buckets** with **paths**, not folders.

**Bucket Structure:**

```
tessera-assets/                    # Main bucket
├── originals/
│   ├── asset-001/
│   │   └── image.jpg            # Original upload
│   └── asset-002/
│       └── document.pdf
│
└── variants/
    ├── asset-001/
    │   ├── thumbnail.jpg        # Generated variant
    │   ├── optimized.jpg        # Generated variant
    │   └── webp.jpg             # Generated variant
    └── asset-002/
        └── preview.pdf
```

**In Code:**

```go
// StoragePort interface (no implementation details)
type StoragePort interface {
    Save(ctx context.Context, path string, data []byte) error
    Read(ctx context.Context, path string) ([]byte, error)
}

// MinIO Adapter (implementation)
func (m *MinIOStorage) Save(ctx, path, data) error {
    _, err := m.client.PutObject(ctx, "tessera-assets", path, data, -1, minio.PutObjectOptions{})
    return err
}
```

---

## Navigation

Previous: **[02 - Flows](02-flows.md)**

Next: **[04 - Guidelines](04-guidelines.md)**

Learn the architectural rules and conventions followed throughout the project.
