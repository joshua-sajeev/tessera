# System Components

## 1. Domain Layer

**Location:** `internal/domain/`

Core business entities with no external dependencies.

**Entities:**

- `Asset` - Represents an uploaded asset
- `ProcessingJob` - Represents an async processing task
- `AssetVariant` - Represents a processed variant
- `User` _(v2)_ - User accounts
- `APIKey` _(v2)_ - Authentication
- `Webhook` _(v2)_ - Event notifications
- `Folder` _(v2)_ - Asset organization

**Responsibilities:**

- Define business rules
- Validate domain invariants
- Model relationships
- Zero external imports (no gorm, no redis, no http packages)

---

## 2. Application Layer (Planned)

**Location:** `internal/application/`

Use cases and orchestration. Will coordinate between domain and ports.

**Planned Responsibilities:**

- Implement use cases (UploadAsset, ProcessAsset, DownloadAsset)
- Call domain logic
- Orchestrate dependencies via ports
- Handle errors and transaction boundaries
- Application-specific validation

**Key Pattern:**

```
UseCase(input) -> Domain Logic -> Ports -> Response
```

---

## 3. Ports (Interfaces)

**Location:** `internal/ports/`

Define contracts for external dependencies. No implementation.

**Port Types:**

| Port | Purpose | Implementation Status |
|------|---------|----------------------|
| AssetRepository | Asset persistence | PostgreSQL adapter ✅ |
| ProcessingRepository | Processing job persistence | PostgreSQL adapter ✅ |
| Storage | Object storage | Planned (MinIO) |
| Queue | Job queue | Planned (Redis) |

**Key Principle:**

- Ports are interfaces in `internal/ports/`
- Adapters implement these interfaces
- Application depends on ports, not adapters

Current implemented ports:

- AssetRepository
- ProcessingRepository

---

## 4. Adapters (Infrastructure)

**Location:** `internal/adapters/`

Concrete implementations of ports. Can be replaced without affecting core logic.

**Adapter Types:**

### PostgreSQL Adapter (`adapters/postgres/`) ✅ Implemented

- AssetRepository implementation
- ProcessingRepository implementation
- Database schema management
- Query logic
- Connection management
- Integration tests

### HTTP Adapter (`adapters/http/`) 🔄 Planned

- REST endpoints
- Request/response handling
- Route definitions
- Error responses

### Planned: MinIO Adapter (`adapters/minio/`) 🔄 Planned

- Storage interface implementation
- Upload/download logic
- Variant storage
- Bucket management

### Planned: Redis Adapter (`adapters/redis/`) 🔄 Planned

- Queue interface implementation
- Job enqueueing
- Job dequeuing

---

## Navigation

Previous: **[00 - Overview](00-overview.md)**

Next: **[02 - Flows](02-flows.md)**

See how a request moves through the architecture from upload to processing.
