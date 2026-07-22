# Request Flow

## Upload Asset Flow

```
1. Client sends POST /api/v1/assets with file
   ↓
2. HTTP Adapter validates request
   ↓
3. Application calls UploadAsset use case
   ↓
4. Domain creates Asset entity
   ↓
5. Application calls Storage.Save() → MinIO adapter stores original
   ↓
6. Application calls AssetRepository.Create() → PostgreSQL stores metadata
   ↓
7. Application calls Queue.Enqueue() → Redis queues processing job
   ↓
8. HTTP Adapter returns 202 Accepted with asset ID
   ↓
9. Client receives response
```

## Processing Flow

```
1. Worker polls Redis for jobs (Queue)
   ↓
2. Worker receives ProcessingJob
   ↓
3. Application calls ProcessAsset use case
   ↓
4. Application calls Storage.Read() → MinIO retrieves original
   ↓
5. Domain/Worker generates variants (resize, optimize)
   ↓
6. Application calls Storage.Save() → MinIO stores each variant
   ↓
7. Application calls ProcessingRepository.Update() → Update processing job status
   ↓
8. Application calls AssetRepository.Update() → Update asset status
   ↓
9. Processing complete
```

## Download Asset Flow

```
1. Client sends GET /api/v1/assets/{id}/variants/{variant_type}
   ↓
2. HTTP Adapter validates request
   ↓
3. Application calls DownloadAsset use case
   ↓
4. Application calls AssetRepository.Get() → Fetch asset metadata
   ↓
5. Application calls Storage.Read() → MinIO retrieves variant
   ↓
6. HTTP Adapter returns file with appropriate headers
   ↓
7. Client receives asset
```

---

# Sequence Diagram

```mermaid
sequenceDiagram
    Client->>HTTP API: POST /assets (file)
    HTTP API->>Application: UploadAsset(file)
    Application->>Domain: Create Asset
    Application->>Storage: Save original
    Storage->>MinIO: Upload file
    Application->>AssetRepository: Save metadata
    AssetRepository->>PostgreSQL: INSERT asset
    Application->>Queue: Enqueue job
    Queue->>Redis: RPUSH processing_queue
    HTTP API-->>Client: 202 Accepted {assetId}
    
    Worker->>Queue: Dequeue job
    Queue->>Redis: LPOP processing_queue
    Worker->>Application: ProcessAsset(jobId)
    Application->>Storage: Read original
    Storage->>MinIO: Download file
    Application->>Domain: Generate variants
    Application->>Storage: Save variant_thumb
    Storage->>MinIO: Upload variant
    Application->>Storage: Save variant_optimized
    Storage->>MinIO: Upload variant
    Application->>ProcessingRepository: Update job status
    ProcessingRepository->>PostgreSQL: UPDATE processing_job
    Application->>AssetRepository: Update asset status
    AssetRepository->>PostgreSQL: UPDATE assets
```

---

# Architecture Diagram

```mermaid
graph TD
    A[Client] --> B[HTTP Handler]
    B --> C[Upload Use Case]

    C --> D[AssetRepository]
    D --> E[(PostgreSQL)]

    C --> F[Storage]
    F --> G[(MinIO)]

    C --> H[Queue]
    H --> I[(Redis)]

    B --> A

    J[Worker] --> H
    J --> F
    J --> K[ProcessingRepository]
    K --> E
```

---

## Navigation

Previous: **[01 - Layers](01-layers.md)**

Next: **[03 - Structure](03-structure.md)**

Explore how the project is organized on disk and how each package maps to the architecture.
