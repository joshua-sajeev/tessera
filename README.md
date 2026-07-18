# Tessera

> An asset processing platform built in Go.

## Overview

Tessera is a backend service that enables applications to upload, process, and manage digital assets through a REST API.

The platform stores original assets in object storage, processes them asynchronously using background workers, and serves optimized variants while maintaining metadata and processing history.

The project is being built to explore backend engineering concepts including Hexagonal Architecture, asynchronous processing, object storage, caching, and event-driven systems.

---
## Documentation

Detailed architecture, design decisions, and development guidelines are available in the [Architecture Documentation](./docs/architecture/README.md).

---
## Goals

- Build a production-style backend in Go
- Practice Hexagonal Architecture
- Learn asynchronous job processing
- Integrate object storage using MinIO
- Use PostgreSQL for metadata
- Implement Redis-backed background workers
- Explore event-driven architecture with Kafka
- Deploy services using Kubernetes

---

## Planned Features

- Asset upload and management
- Image transformations
  - Resize
  - Compress
  - Convert formats
  - Generate thumbnails
  - Watermarking
- Background processing
- Processing status tracking
- REST API
- Object storage
- Authentication

---

## Planned Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go |
| HTTP | chi |
| Database | PostgreSQL |
| Object Storage | MinIO |
| Cache / Queue | Redis |
| Migrations | Goose |
| Containers | Docker |

Future

- Kafka
- Kubernetes
- Prometheus
- Grafana

---

## Project Status

This project is currently in the planning and design phase.

The first milestone is to establish a clean architecture and development environment before implementing application features.

---

## Roadmap

- [x] Initialize repository
- [ ] Design system architecture
- [ ] Design database schema
- [ ] Setup Docker development environment
- [ ] Implement asset upload
- [ ] Integrate MinIO
- [ ] Add background workers
- [ ] Implement image processing
- [ ] Add Redis queue
- [ ] Introduce Kafka
- [ ] Deploy using Kubernetes

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
