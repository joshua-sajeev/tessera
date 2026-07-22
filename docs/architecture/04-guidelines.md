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

## Application Layer (Planned)

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
type Storage interface {
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
    return err // Must match Storage interface
}
```

## HTTP Adapter (Planned)

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

# Development Workflow

## Current Project State

The project currently focuses on the persistence layer. The workflow below reflects the current implementation path.

## Adding a New Domain Entity

1. **Define domain model** in `internal/domain/{entity}/`
   - Create entity file (e.g., `asset.go`)
   - Define status types (e.g., `status.go`)
   - Define domain errors (e.g., `errors.go`)

2. **Write domain unit tests** in `{entity}_test.go`

3. **Define port interface** in `internal/ports/`
   - Create repository interface if persistence needed
   - Keep interface focused and application-agnostic

4. **Implement PostgreSQL adapter** in `internal/adapters/postgres/`
   - Implement repository interface
   - Add integration tests

5. **Add Goose migration** if schema changes in `migrations/`
   - Follow naming convention: `NNNN_description.sql`
   - Include UP and DOWN migrations

6. **Update architecture documentation** as necessary

## Example: Adding a New Entity

```
1. Define or extend the domain model
   → Create `internal/domain/newentity/newentity.go`

2. Update the appropriate port interface
   → Add methods to `internal/ports/newentity_repository.go`

3. Implement PostgreSQL adapter
   → Create `internal/adapters/postgres/newentity_repository.go`

4. Add Goose migration if schema changes
   → Create `migrations/NNNN_add_newentity_table.sql`

5. Write integration tests
   → Create `internal/adapters/postgres/newentity_repository_test.go`

6. Update architecture documentation
   → Reflect changes in relevant doc files
```

## Local Development Setup

```bash
make dev-setup          # Run docker-compose, init databases
make migrate            # Run migrations
make test               # Run all tests
make test-coverage      # Coverage report
```

---

# Naming Conventions

## Interface Names

Use clear, descriptive names without "Port" suffix:

- `AssetRepository` ✅
- `ProcessingRepository` ✅
- `Storage` ✅
- `Queue` ✅

Avoid:
- `RepositoryPort` ❌
- `StoragePort` ❌
- `QueuePort` ❌

## File Names

Use consistent, descriptive names in lowercase with underscores:

- `errors.go` ✅ (plural, consistent)
- `asset_repository.go` ✅
- `processing_repository.go` ✅

Avoid:
- `error.go` ❌ (inconsistent)

## Directory Structure

Keep related files organized by concern:

```
internal/domain/asset/
  ├── asset.go
  ├── status.go
  └── errors.go

internal/ports/
  ├── asset_repository.go
  └── processing_repository.go

internal/adapters/postgres/
  ├── asset_repository.go
  └── processing_repository.go
```

---

## Navigation

Previous: **[03 - Structure](03-structure.md)**

You've reached the end of the architecture documentation.

Return to the **[Architecture Index](README.md)** or continue exploring the source code.
