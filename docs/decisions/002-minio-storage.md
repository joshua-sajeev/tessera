# ADR 002: Choosing MinIO as the Object Storage Implementation

**Status:** Accepted  
**Date:** 2026-07-23

## Context

Tessera requires object storage for uploaded assets and generated variants. Since the application follows Hexagonal Architecture, storage is accessed through the `Storage` port, allowing different implementations without affecting the domain or application layers.

I wanted an object storage solution that is easy to run locally while also giving me hands-on experience with technologies commonly used in production.

## Decision

I have decided to use **MinIO** as the initial implementation of the `Storage` port.

## Consequences

**Why I made this choice:**

- **Learning Objective:** I wanted practical experience with object storage and the S3 API.
- **Production-like Development:** MinIO provides an S3-compatible API while running locally with Docker.
- **Infrastructure Independence:** The storage implementation can later be replaced with AWS S3 or another provider by adding a new adapter.
- **Developer Experience:** No cloud account or external infrastructure is required for local development.

**Accepting the trade-offs:**

- Running MinIO adds another service to the development environment.
- Migrating to another provider requires implementing a new adapter.

## Compliance

- All storage operations must go through the `Storage` interface in `internal/ports/`.
- The domain and application layers must not depend on MinIO SDK types.
- MinIO-specific code must remain isolated within the adapter.
