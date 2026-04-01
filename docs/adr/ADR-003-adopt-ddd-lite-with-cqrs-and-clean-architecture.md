# ADR-003: Adopt DDD-Lite with CQRS and Clean Architecture

**Status:** Accepted
**Date:** 2026-03-26
**Deciders:** kfreiman

## Context

The Casino Transaction Management System needs to process high-volume bet and win events via Kafka and serve them via an API. Per `docs/requirements.md`, the API must return data in JSON format; therefore, we will implement an HTTP handler instead of gRPC or GraphQL. We need an architecture that ensures financial data integrity, supports complex business rules, and makes it easy to achieve the required 85% test coverage.

## Decision

We will adopt **DDD-Lite with CQRS and Clean Architecture**.

### Structure

```
internal/transactions/
├── domain/           # Transaction entity, value objects, repository interface
├── app/
│   ├── command/     # Handlers for Kafka messages (write path)
│   └── query/        # Handlers for API requests (read path)
├── ports/
│   ├── http/         # HTTP handlers
│   └── kafka/        # Consumer group handlers
└── adapters/
    └── postgres/    # sqlc-generated type-safe repository
```

## Alternatives Considered

### Option 1: DDD-Lite with CQRS and Clean Architecture (CHOSEN)

**Pros:**

- High testability; domain logic can be unit-tested without DB or Kafka
- Clear separation of concerns
- Prevents leaky abstractions (no ORM tags in domain)
- Separates read and write models for independent optimization

**Cons:**

- Higher initial boilerplate
- Steeper learning curve for developers unfamiliar with DDD/CQRS
- Requires maintaining SQL query files and running sqlc code generation

### Option 2: Standard Layered Architecture

**Pros:**

- Simpler and faster to start
- Fewer files and directories

**Cons:**

- Often leads to "Fat Services" and "Anemic Domain Models"
- Testing usually requires mocking the database, making 85% coverage harder

## Tradeoffs

**What we're optimizing for:**

- Testability (domain logic independent of infrastructure)
- Maintainability of bet/win business rules
- 85% code coverage requirement

**What we're sacrificing:**

- Initial development speed (more structure)
- Simplicity (higher learning curve)

## Consequences

### Positive

- Domain logic highly testable without DB or Kafka
- Kafka consumer only cares about Commands, API only cares about Queries
- Prevents `json` or `db` tags in domain layer
- Easy to achieve 85%+ coverage

### Negative

- Higher initial boilerplate (more files and directories)
- Steeper learning curve for DDD/CQRS unfamiliar developers
- Requires sqlc code generation workflow

## Implementation Notes

- Use **Constructors** (`NewTransaction`) for validation
- Use **Functional Update** pattern in repositories for atomic state changes
- Keep `domain` package dependency-free (no imports of external frameworks or DB drivers)
- Generate type-safe SQL access layer with **sqlc**
- Do not use public fields in domain entities
- Do not bypass the application layer from ports
- Do not use ORM libraries (e.g., GORM) in repository layer

## Follow-up Actions

- [x] Select DDD-Lite with CQRS (kfreiman, 2026-03-26)
- [ ] Create domain layer structure (kfreiman, 2026-03-26)
- [ ] Configure sqlc for type-safe repository (kfreiman, 2026-03-26)
- [ ] Verify domain layer achieves >90% coverage without database (kfreiman, 2026-03-26)

## References

- Skill followed: `ddd-lite-go` (based on Three Dots Labs patterns)
- [sqlc documentation](https://sqlc.dev/)
- Related ADR: [ADR-001: Choose PostgreSQL for Transaction Storage](ADR-001-choose-postgresql-for-transaction-storage.md)
- Related ADR: [ADR-002: Choose Kafka and Idempotent Message Processing](ADR-002-choose-kafka-and-idempotent-message-processing.md)
- Related ADR: [ADR-004: Database Schema and Indexing](ADR-004-database-schema-and-indexing.md)

## Revision History

- 2026-03-26: Initial decision (kfreiman)
- 2026-03-29: Renumbered from ADR-004 to ADR-003 to fix ordering (kfreiman)
