# ADR 004: Multi-Tenancy Strategy

Status: Accepted
Date: 2026-08-17

## Context

Tessera is a learning project for backend engineering. Version 1 will serve multiple users concurrently. Without multi-tenancy design from the start, adding user isolation later requires significant refactoring of the database schema and repository layer.

Key requirements:
- Users should only access their own assets
- Authentication via API keys
- Storage quotas per user
- Fair job processing (no single user starves others)

## Decision

Implement user isolation at the database layer with user_id foreign keys and enforce user_id in all repository queries. Use hashed API keys with Bearer token authentication.

## Solution

Three-layer approach:

1. Database Layer: Add users table, user_id foreign keys on assets and processing_jobs
2. Auth Layer: New Authenticator port that verifies API keys and returns user objects
3. Repository Layer: All queries require user_id parameter, enforced at method signature level

Example repository method:

```go
func (r *AssetRepository) GetByID(ctx context.Context, assetID, userID uuid.UUID) (*Asset, error) {
    // Query includes WHERE user_id = ? AND id = ?
}
```

## Consequences

Why this approach:
- Database-enforced isolation prevents accidental cross-user data leaks
- Queries remain fast with composite indices on (user_id, status)
- Method signatures force developers to provide user_id
- Clear audit trail of which user owns each asset

Trade-offs accepted:
- Schema migration required to add users table and user_id columns
- All existing repositories must be updated to accept user_id parameter
- API key hashing adds small cryptographic overhead per request

## Alternatives Considered

1. Separate databases per user: Complete isolation but massive operational overhead
2. Row-level security (PostgreSQL RLS): Database enforces it automatically, but harder to test and debug
3. No multi-user support: Simpler, but unrealistic and no learning value

Chosen because it balances simplicity, learnability, and production-readiness.

## Compliance

- All repository methods must include user_id parameter
- Tests must verify User A cannot access User B's data
- API key must be hashed before storage in database
- HTTP middleware extracts Bearer token and validates via Authenticator port
- Documentation must explain the isolation model (see 05-database.md)

## References

For detailed schema design, indexing strategy, and migration steps, see [05 - Database](../architecture/05-database.md).
