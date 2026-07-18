# Layer Responsibilities

## Domain Layer

**Responsibility:** Define and enforce business rules.

**Should:**

- Define entities and value objects
- Implement business logic
- Validate invariants
- Return domain errors

**Should NOT:**

- Access databases
- Make HTTP calls
- Know about frameworks
- Handle I/O

**Example:**

```go
func (a *Asset) IsProcessable() bool {
    return a.Status == StatusUploaded
}
```

## Application Layer

**Responsibility:** Orchestrate use cases using domain logic and ports.

**Should:**

- Implement use cases (workflows)
- Call domain logic
- Coordinate multiple ports
- Handle transactions
- Map errors

**Should NOT:**

- Contain business logic
- Access infrastructure directly
- Return infrastructure types
- Handle HTTP/database details

**Example:**

```go
func (u *UploadUseCase) Execute(ctx, file) (*Asset, error) {
    asset := domain.NewAsset(file)
    if err := u.storage.Save(ctx, path, file); err != nil {
        return nil, fmt.Errorf("storage failed: %w", err)
    }
    return asset, nil
}
```

## Ports (Interfaces)

**Responsibility:** Define contracts for external dependencies.

**Should:**

- Be simple, focused interfaces
- Represent business capabilities
- Have application-level semantics
- Be implementation-agnostic

**Example:**

```go
type StoragePort interface {
    Save(ctx context.Context, path string, data []byte) error
    Read(ctx context.Context, path string) ([]byte, error)
}
```

## Adapters (Infrastructure)

**Responsibility:** Implement ports using concrete tools.

**Should:**

- Implement exactly one port
- Handle tool-specific details
- Manage connections/lifecycles
- Translate to/from infrastructure types

**Should NOT:**

- Contain business logic
- Import domain types
- Violate port contracts

**Example:**

```go
type MinIOAdapter struct {
    client *minio.Client
}

func (m *MinIOAdapter) Save(ctx, path, data) error {
    _, err := m.client.PutObject(ctx, "bucket", path, data, -1, opts)
    return err // Must match StoragePort interface
}
```

## HTTP Adapter

**Responsibility:** Translate HTTP to/from application.

**Should:**

- Parse HTTP requests
- Validate input
- Call application use cases
- Format responses
- Handle HTTP-specific concerns (headers, status codes)

**Should NOT:**

- Contain business logic
- Access infrastructure directly
- Bypass application layer

---
# Configuration Management

**Pattern:** Each adapter has its own config file.

```yaml
# configs/app.yaml
server:
  port: 8080
  timeout: 30s

# configs/database.yaml
postgres:
  host: localhost
  port: 5432
  database: tessera
  
# configs/storage.yaml
minio:
  endpoint: localhost:9000
  bucket: tessera-assets
  
# configs/queue.yaml
redis:
  host: localhost
  port: 6379
```

Load via environment variables or config files (use Viper or similar).

---

# Development Workflow

## Local Setup

```bash
make dev-setup          # Run docker-compose, init databases
make migrate            # Run migrations
make run-api            # Start API server
make run-worker         # Start background worker
make test               # Run all tests
make test-coverage      # Coverage report
```

## Adding a New Feature

1. **Define domain entity** → `internal/domain/`
2. **Write domain tests** → `_test.go` files
3. **Create use case** → `internal/application/`
4. **Test use case** → Mock ports
5. **Implement port** → Update `internal/ports/`
6. **Implement adapters** → Add to `internal/adapters/`
7. **Add HTTP handler** → `internal/adapters/http/handler/`
8. **Test end-to-end** → Integration tests
9. **Update docs** → API docs, examples

---

## Navigation

Previous: **[03 - Structure](03-structure.md)**

You've reached the end of the architecture documentation.

Return to the **[Architecture Index](README.md)** or continue exploring the source code.
