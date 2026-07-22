# Tessera

> An asset processing platform built in Go using Hexagonal Architecture.

Tessera is a backend service for storing, processing, and serving digital assets. It is designed around a clean, modular architecture where business logic is isolated from infrastructure, making the system easier to test, maintain, and extend.

The project is being built as a learning-focused backend system exploring asynchronous processing, object storage, and scalable service design.

---

## Current Status

The project currently provides the persistence foundation for the platform.

### Implemented

* Domain models
* Repository ports
* PostgreSQL persistence adapters
* Goose database migrations
* PostgreSQL integration tests
* Docker Compose development environment

### Planned

* Application layer
* HTTP API
* MinIO object storage
* Redis job queue
* Background worker
* Asset processing pipeline

---

## Architecture

Tessera follows **Hexagonal Architecture (Ports and Adapters)**.

```text
External Interfaces
        │
        ▼
Application (Planned)
        │
        ▼
Ports (Interfaces)
        │
        ▼
Domain
        ▲
        │
Infrastructure Adapters
(PostgreSQL, MinIO, Redis)
```

For a detailed explanation of the architecture and design decisions, see the documentation:

* [Architecture Documentation](./docs/architecture/README.md)
* [Architecture Decision Records](./docs/decisions/)

---

## Tech Stack

| Category   | Technology     |
| ---------- | -------------- |
| Language   | Go             |
| Database   | PostgreSQL     |
| Driver     | pgx/v5         |
| Migrations | Goose          |
| Testing    | Go testing     |
| Containers | Docker Compose |

---

## Getting Started

### Prerequisites

* Go 1.21+
* Docker & Docker Compose
* Goose
* PostgreSQL (or Docker)

### Clone

```bash
git clone https://github.com/joshua-sajeev/tessera.git
cd tessera
```

### Start Development Services

```bash
make up
```

### Run Database Migrations

```bash
make goose-up
```

### Migration Status

```bash
make goose-status
```

### Run Tests

```bash
go test ./...
```

---

## Project Structure

```text
tessera/
├── cmd/                    # Application entrypoints
├── internal/
│   ├── domain/             # Business entities
│   ├── ports/              # Interfaces
│   ├── adapters/           # Infrastructure implementations
│   └── config/             # Configuration
├── migrations/             # Goose migrations
├── deployments/            # Docker Compose
└── docs/                   # Documentation
```

A more detailed repository breakdown is available in the architecture documentation.

---

## Documentation

The project documentation is located under `docs/architecture`.

* **00 - Overview** — Project goals and architecture
* **01 - Layers** — Domain, Ports, Adapters, and Application
* **02 - Flows** — Request flows and sequence diagrams
* **03 - Structure** — Repository layout
* **04 - Guidelines** — Development conventions
* **05 - Database** — Database schema and persistence

---

## Roadmap

* [x] Domain models
* [x] Repository ports
* [x] PostgreSQL persistence
* [x] Goose migrations
* [x] Integration tests
* [ ] Application layer
* [ ] HTTP API
* [ ] MinIO integration
* [ ] Redis queue
* [ ] Background worker
* [ ] Asset processing pipeline

---

## Contributing

Contributions are welcome.

Before opening a pull request:

* Ensure all tests pass.
* Update documentation when architecture or behavior changes.
* Follow the project's architectural guidelines.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
