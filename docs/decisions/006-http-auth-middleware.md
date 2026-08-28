# ADR 006: Implementing HTTP Authentication Middleware

Status: Accepted  
Date: 2026-08-28  

## Context

Tessera's multi-tenancy strategy (ADR 004) and choice of API key authentication (ADR 005) require HTTP routes to be protected. We need an HTTP middleware that intercepts requests to secure endpoints, extracts the API key, authenticates the key using the `ports.Authenticator`, and makes the resolved user information available to downstream handlers.

The HTTP layer also needs to enforce basic tenant isolation, ensuring that users can only access or modify their own resources.

## Decision

We will implement a route-level HTTP middleware wrapper `RequireAuth`.

1. **Header Format**: The client must supply the API key in the standard `Authorization` header using the `Bearer` scheme:
   ```http
   Authorization: Bearer tsr_v1_...
   ```
2. **Context Propagation**: We will define a custom type for context keys to avoid collisions, and functions to inject and retrieve the authenticated `*user.User` from the request context:
   - `WithUser(ctx context.Context, u *user.User) context.Context`
   - `UserFromContext(ctx context.Context) *user.User`
3. **Authentication Steps**:
   - Extract the token from the header. If missing or malformed, respond with `401 Unauthorized`.
   - Call the `Authenticator` port to get the user.
   - If the key is invalid or not found, respond with `401 Unauthorized`.
   - If the user is suspended or deleted, respond with `403 Forbidden`.
   - Store the user in the context and pass the request downstream.
4. **Selective Guarding**:
   - Apply the middleware to sensitive endpoints (e.g. `GET /users/{id}`, `PUT /users/{id}/status`, `PATCH /users/{id}/status`).
   - Keep registration (`POST /users`) public.
5. **Authorization Verification**:
   - The handler must verify that the authenticated user matches the requested user ID to prevent cross-user data access.

## Consequences

* **Explicit Route Control**: Developers explicitly declare which routes require authentication in the router configuration, which is easy to audit.
* **Standard Context Propagation**: Injected user information is accessible cleanly across handlers.
* **Granular Status Handling**: Users with invalid statuses (Suspended/Deleted) are correctly rejected with `403 Forbidden`.

## Compliance

* **Token Scheme**: Only the `Bearer` scheme is accepted in the `Authorization` header.
* **Type Safety**: Request contexts must use a custom, unexported key type for the user object.
* **Verification**: Unit tests must cover all auth middleware scenarios, including missing header, invalid format, expired/invalid key, suspended user, deleted user, and path-value user authorization mismatch.
