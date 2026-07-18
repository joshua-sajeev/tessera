# ADR 001: Choosing Hexagonal Architecture for Tessera

Status: Accepted
Date: 2026-04-18

## Context

I’m building Tessera to sharpen my backend engineering skills and create a robust, production-grade asset processor. I initially considered using a simple Repository Pattern to manage data access. However, I realized that for my goal of professional growth, I wanted an architecture that enforces a stricter separation of concerns. I wanted to move beyond simple repository-service layering and truly isolate my business logic from infrastructure (PostgreSQL, MinIO, Redis, HTTP).

## Decision

I have decided to implement the Hexagonal Architecture (Ports and Adapters) pattern.
The system is now divided into:
- Domain: Core logic and entities (strictly zero external dependencies).
- Application: Orchestration of use cases.
- Ports: Interfaces defining required business capabilities.
- Adapters: Concrete implementations (PostgreSQL, MinIO, Redis, HTTP).

## Consequences

Why I made this choice:
- Learning Objective: My primary driver is to master Hexagonal Architecture. Implementing this pattern in Go is a key learning milestone for me.
- True Testability: By isolating business logic from infrastructure, I can write fast, reliable unit tests using mocks without needing to spin up a full environment.
- Infrastructure Independence: I can experiment with different tools (e.g., swapping MinIO for AWS S3) by simply adding a new Adapter, without touching my business code.
- Clean Code: It prevents "spaghetti code" by forcing me to design clear interfaces (Ports).

**Accepting the trade-offs:**

- Boilerplate: I accept that this pattern introduces more setup and interface overhead compared to a standard repository approach.
- Complexity: I recognize that for a small project, this might feel like "over-engineering," but I believe the clarity and long-term maintainability make it the right choice for my learning.

## Compliance

- Domain Purity: internal/domain/ must contain no database drivers, web frameworks, or I/O logic.
- Interface-First: Every interaction with the "outside world" must go through an interface defined in internal/ports/.
