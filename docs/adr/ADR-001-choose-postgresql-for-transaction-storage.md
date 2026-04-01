# ADR-001: Choose PostgreSQL for Transaction Storage

**Status:** Accepted
**Date:** 2026-03-26
**Deciders:** kfreiman

## Context

The Casino Transaction Management System must handle financial data (bets and wins) with absolute reliability. We need a database that ensures data integrity, handles concurrent requests without locking up, and supports efficient chronological sorting.

## Decision

We will use **PostgreSQL** as the primary relational database.

### Technical Details

- **Go Driver**: `pgx/v5` for high-performance, native PostgreSQL support.
- **Identifiers**: **UUIDv7** for primary keys. Since UUIDv7 is time-ordered, it allows PostgreSQL to insert and sort transactions chronologically with high efficiency.
- **Data Types**: `BIGINT` for `amount` (storing values in the smallest unit, e.g., cents) to ensure mathematical precision and speed.

## Rationale: Why PostgreSQL?

### 1. Data Integrity (ACID)

Financial systems cannot afford "eventual consistency." PostgreSQL's strict ACID compliance ensures that every bet and win is recorded permanently and accurately. If a power failure occurs mid-transaction, the data remains uncorrupted.

### 2. Native UUID Support

Unlike many databases that store UUIDs as slow, bulky strings, PostgreSQL has a native 16-byte `UUID` type. This makes our **UUIDv7** primary keys extremely fast to index and search, which is critical as the transaction history grows.

### 3. Concurrency (MVCC)

PostgreSQL uses Multi-Version Concurrency Control. This means users can read their transaction history (Selects) without being blocked by other users placing bets (Inserts). This is vital for a smooth user experience during peak casino activity.

### 4. Integration with Go

The combination of PostgreSQL with the `pgx` driver and `sqlc` provides a type-safe, high-performance environment that catches many errors at compile-time rather than run-time.

## Alternatives Considered

### Option 1: PostgreSQL with pgx (CHOSEN)

**Pros:**

- Robust ACID compliance ensures financial data integrity
- Industry-standard JSONB support for future metadata extensibility
- Native UUID type for efficient UUIDv7 indexing
- Superior performance and Go-idiomatic features
- MVCC for non-blocking concurrent reads

**Cons:**

- Slightly higher operational complexity compared to SQLite (mitigated by Docker)

### Option 2: MySQL

**Pros:**

- Familiar to some team members

**Cons:**

- Inferior UUID support compared to PostgreSQL
- Less efficient for time-ordered inserts
- Inferior JSON support compared to PostgreSQL
- Less type safety for financial precision

### Option 3: SQLite

**Pros:**

- Simpler local development setup

**Cons:**

- Lacks concurrency and performance required for multi-user casino environment
- No native UUID type
- Poor MVCC support for concurrent reads/writes

## Tradeoffs

**What we're optimizing for:**

- Reliability and data correctness
- ACID compliance for financial integrity
- Native support for time-ordered UUIDs
- Non-blocking concurrent reads

**What we're sacrificing:**

- Zero-configuration simplicity of SQLite (mitigated by using Docker for development)

## Consequences

### Positive

- Rock-solid reliability for financial records
- UUIDv7 with PostgreSQL allows stable, deterministic "Newest to Oldest" sorting without extra timestamp logic
- ACID compliance ensures financial data integrity
- JSONB support allows for future metadata extensibility
- Non-blocking reads via MVCC for smooth user experience during peak activity

### Negative

- Requires managing a database server (mitigated by using managed cloud services in production)
- Slightly higher operational complexity for local development (mitigated by Docker)

## Risks

- **Risk:** Operational overhead of managing PostgreSQL
- **Mitigation:** Use Docker for local development; managed cloud DB for production

## Implementation Notes

- Use `pgxpool` for efficient connection management
- Always use the native `UUID` type for the `id` column
- Use `BIGINT` for currency amounts to avoid precision errors common with floating-point numbers
- Use explicit database transactions for operations involving multiple steps
- Avoid long-running transactions that could lock tables

## Follow-up Actions

- [x] Select PostgreSQL and pgx driver (kfreiman, 2026-03-26)
- [x] Set up Docker Compose for local PostgreSQL (kfreiman, 2026-03-26)

## References

- [pgx documentation](https://github.com/jackc/pgx)
- [UUIDv7 Specification](https://www.rfc-editor.org/rfc/rfc9562)
- Related ADR: [ADR-002: Choose Kafka and Idempotent Message Processing](ADR-002-choose-kafka-and-idempotent-message-processing.md)

## Revision History

- 2026-03-26: Initial decision (kfreiman)
- 2026-03-29: Added UUIDv7 rationale and BIGINT data type details (kfreiman)
