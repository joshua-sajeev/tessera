# Overview

Tessera is a backend service that stores uploaded assets, processes them asynchronously, and serves optimized variants through a REST API.

## What Tessera Is

- An asset processing backend
- Async-first architecture
- Modular and testable
- Infrastructure-agnostic
- Multi-user capable (v1 includes auth and isolation)

## What Tessera Is NOT

- An image editor
- A frontend application
- A CDN
- An authentication provider (though it manages API keys)

---

## Current Status

The current implementation provides the foundational persistence and object storage layers for Tessera, including:

- Domain models (Asset, ProcessingJob, AssetVariant)
- Ports (AssetRepository, ProcessingRepository, Storage)
- PostgreSQL and MinIO adapters with integration tests
- Goose database migrations (initial schema)

### Planned (Multi-User & Integration)

The following components are designed and planned for near-term implementation:

- User domain model (User entity)
- Authenticator port and API key validation adapter
- Multi-user data isolation in database migrations, repository ports, and adapters

The HTTP API, queueing (Redis), and application use cases will be added in future milestones.

---

# Version 1 Features

1. User Authentication - API key-based auth with Bearer tokens
2. Upload Asset - Accept asset uploads via HTTP with auth
3. Data Isolation - Users only access their own assets
4. Store Original - Persist original asset to object storage
5. Save Metadata - Record asset metadata in database with user_id
6. Create Processing Job - Queue asset processing work
7. Return 202 Accepted - Immediate client response
8. Worker Processes Asset - Async job processing (per-user fairness)
9. Generate Variants - Create optimized copies
10. Update Status - Track processing progress
11. Download Asset - Serve assets and variants (with auth)
12. Storage Quotas - Enforce per-user storage limits

---

# Architectural Style: Hexagonal Architecture

## Why Hexagonal?

| Benefit | Why It Matters |
| --- | --- |
| Isolated Business Logic | Domain logic independent of frameworks/tools |
| Easier Testing | Core logic testable without infrastructure |
| Infrastructure Replaceable | Swap PostgreSQL for MongoDB, MinIO for S3, etc. |
| Learning Objective | Learn Hexagonal Architecture |
| Multi-Tenancy Ready | Auth and isolation baked in from the start |

## Core Principle

Dependencies point inward. The domain knows nothing about HTTP, databases, queues, or users.

```
External World (HTTP, DB, Storage, Queues, Auth)
        Down
    Adapters (Ports implementation)
        Down
    Application (Use Cases)
        Down
    Domain (Core Business Logic)
        Up
    Ports (Interfaces/Contracts)
```

---

## Multi-User Design

Tessera v1 is built with multi-user isolation from the ground up:

- Database Layer: All tables have user_id foreign keys
- Auth Layer: API key verification maps tokens to users
- Repository Layer: All queries include user_id filters
- API Layer: Bearer token middleware enforces auth on every request
- Quota Layer: Users have storage limits and fairness rules

See ADR 004: Multi-Tenancy Strategy for deep dive.

---

## Continue

Next: [01 - Layers](01-layers.md)

Learn how Tessera separates responsibilities between the Domain, Application, Ports, and Adapters layers.
