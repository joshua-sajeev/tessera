
Yes. For a personal project, you can make the ADR more direct and focus on **your design decision and learning goals** rather than "clients", "production", etc.

I would use this version:

````markdown
# ADR 005: Choosing API Key Authentication for Tessera

Status: Accepted  
Date: 2026-08-20

## Context

Tessera is an asset-processing API with multi-user support. Each request needs to be associated with a specific user so that assets and processing jobs remain isolated between users.

I considered several authentication approaches, including JWT-based authentication and API keys. Since Tessera is primarily an API and does not currently require browser-based sessions, I decided that API keys provide a simpler authentication mechanism while still allowing me to learn and implement the important authentication concepts.

The authentication mechanism also needs to integrate cleanly with Tessera's Hexagonal Architecture and multi-user isolation.

## Decision

I have decided to use **API key authentication** for Tessera.

Each API key consists of a lookup identifier and a secret:

```text
api_key_id + secret
````

The database stores:

* `api_key_id` — a unique identifier used to locate the user.
* `api_key_hash` — an Argon2id hash of the secret.

The complete API key is only returned when the key is generated and is never stored in plaintext.

Authentication follows this flow:

```text
API Key
   │
   ▼
Auth Middleware
   │
   │ api_key_id
   ▼
Authenticator Port
   │
   ▼
PostgreSQL Adapter
   │
   │ api_key_hash
   ▼
Argon2id Verification
   │
   ▼
Authenticated User
   │
   │ user_id
   ▼
Application / Repository
```

The implementation follows Tessera's Hexagonal Architecture:

* **Domain:** API-key generation and validation.
* **Port:** `Authenticator` interface defining authentication behavior.
* **Adapter:** PostgreSQL implementation for retrieving users.
* **HTTP layer:** Extracts the API key and establishes the authenticated user.

## Consequences

Why I made this choice:

* **Simple:** API keys are straightforward to generate, store, validate, and use.
* **Good fit for an API:** No session management or refresh-token flow is required.
* **Revocation:** An API key can be invalidated without waiting for token expiration.
* **Multi-user integration:** Successful authentication resolves directly to a `user_id`.
* **Secure storage:** Only the Argon2id hash of the API-key secret is stored.
* **Testability:** Authentication is exposed through the `Authenticator` port, allowing the PostgreSQL adapter to be replaced during testing.
* **Learning objective:** Implementing API-key authentication gives me practical experience with credential generation, secure hashing, validation, dependency inversion, and authentication middleware.

**Accepting the trade-offs:**

* **No automatic expiration:** API keys remain valid until revoked or replaced.
* **Credential management:** Anyone who obtains an API key can authenticate as that user.
* **Database lookup:** Authentication requires a database lookup using `api_key_id` followed by Argon2id verification.
* **Less flexible than JWT:** API keys do not provide built-in claims, expiration, or delegated authorization.

## Compliance

* **Hashed Secrets:** API-key secrets must never be stored in plaintext.
* **Argon2id:** API-key secrets are hashed using Argon2id before being stored.
* **No Plaintext Retrieval:** The original API-key secret cannot be recovered from the database.
* **User Resolution:** Successful authentication resolves to a specific `user_id`.
* **Port-Based Authentication:** Application code depends on the `Authenticator` port rather than the PostgreSQL implementation.
* **Tenant Isolation:** Authenticated `user_id` must be supplied to user-scoped repository operations.
* **Domain Purity:** `internal/domain/` must not contain PostgreSQL, HTTP, or other infrastructure dependencies.

```

This fits the style of your **ADR 001** much better because it explains **why you chose it for Tessera and what you're accepting**, rather than pretending you're documenting requirements for a commercial SaaS product.
```
