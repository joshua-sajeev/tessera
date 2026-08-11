# ADR 003: Separating Metadata and Binary File Storage

**Status:** Accepted  
**Date:** 2026-07-23

## Context

Tessera manages digital assets, which consist of two parts:
1. **Metadata:** Structural information about the asset (e.g., filename, file size, content type, upload status, creation timestamps, and relationship to processed variants).
2. **Binary Payload:** The raw bytes of the file itself (potentially megabytes or gigabytes in size).

Storing both binary files and metadata in a single database (e.g., using PostgreSQL `bytea` or `OID` types) simplifies the architecture but introduces significant performance, scaling, and cost bottlenecks.

## Decision

We will store asset metadata in PostgreSQL and raw binary file contents separately in MinIO (Object Storage).

## Consequences

**Why I made this choice:**

- **Database Performance:** Large binary objects in PostgreSQL cause table bloat, slow down index scans, and consume significant shared buffer memory, degrading query performance for transactional operations.
- **Cost Efficiency:** Object storage (MinIO/S3) is optimized for high-volume unstructured data and is significantly cheaper per gigabyte than high-performance relational database storage (SSD/NVMe).
- **Scalability:** MinIO/S3 handles parallel uploads/downloads and horizontal scaling of throughput far better than a relational database.
- **Backup and Recovery:** Keeping binary data out of PostgreSQL ensures database backups (`pg_dump` or WAL archiving) remain small, fast, and easy to restore.

**Accepting the trade-offs:**

- **Distributed Transactions:** We lose ACID consistency across database records and binary files. A file might be written to MinIO, but the database update could fail (or vice versa), leading to orphaned files or broken references.
- **Garbage Collection:** We must implement asynchronous cleaning processes (garbage collection) to find and delete orphaned files in MinIO that lack corresponding database records.

## Compliance

- The domain layer must only reference files via abstract paths or URLs stored in the metadata.
- All file operations must be performed using the `Storage` port interface, while database operations must go through the repository ports.
- Database transactions must not block waiting for long-running object storage I/O operations.
