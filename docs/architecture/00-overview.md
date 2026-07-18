# Overview

Tessera is a backend service that stores uploaded assets, processes them asynchronously, and serves optimized variants through a REST API.

## What Tessera Is

- An asset processing backend
- Async-first architecture
- Modular and testable
- Infrastructure-agnostic

## What Tessera Is NOT

- An image editor
- A frontend application
- A CDN
- An authentication provider

---

# Version 1 Features

1. **Upload Asset** - Accept asset uploads via HTTP
2. **Store Original** - Persist original asset to object storage
3. **Save Metadata** - Record asset metadata in database
4. **Create Processing Job** - Queue asset processing work
5. **Return 202 Accepted** - Immediate client response
6. **Worker Processes Asset** - Async job processing
7. **Generate Variants** - Create optimized copies
8. **Update Status** - Track processing progress
9. **Download Asset** - Serve assets and variants

---

# Architectural Style: Hexagonal Architecture

## Why Hexagonal?

| Benefit                        | Why It Matters                                  |
| ------------------------------ | ----------------------------------------------- |
| **Isolated Business Logic**    | Domain logic independent of frameworks/tools    |
| **Easier Testing**             | Core logic testable without infrastructure      |
| **Infrastructure Replaceable** | Swap PostgreSQL for MongoDB, MinIO for S3, etc. |
| **Learning Objective**         | Learn Hexagonal Architecture                    |

## Core Principle

Dependencies point inward. The domain knows nothing about HTTP, databases, or queues.

```
External World (HTTP, DB, Storage, Queues)
        ↓
    Adapters (Ports implementation)
        ↓
    Application (Use Cases)
        ↓
    Domain (Core Business Logic)
        ↑
    Ports (Interfaces/Contracts)
```

---

## Continue

Next: **[01 - Layers](01-layers.md)**

Learn how Tessera separates responsibilities between the Domain, Application, Ports, and Adapters layers.
