# Architecture Decision Records

An Architecture Decision Record (ADR) captures an important architecture decision along with its context and consequences.

## Conventions

- Directory: `docs/adr`
- Naming: `ADR-XXX-title.md` with sequential numbering
- Status values: `Proposed`, `Accepted`, `Rejected`, `Deprecated`, `Superseded`

## Workflow

```
Proposed → Accepted → Implemented
    ↓
Rejected

Accepted → Deprecated → Superseded by ADR-XXX
```

## ADRs

| Number | Title | Status | Date |
|--------|-------|--------|------|
| [ADR-001](ADR-001-choose-postgresql-for-transaction-storage.md) | Choose PostgreSQL for Transaction Storage | Accepted | 2026-03-26 |
| [ADR-002](ADR-002-choose-kafka-and-idempotent-message-processing.md) | Choose Kafka for Message Processing with Idempotent Design | Accepted | 2026-03-26 |
| [ADR-003](ADR-003-adopt-ddd-lite-with-cqrs-and-clean-architecture.md) | Adopt DDD-Lite with CQRS and Clean Architecture | Accepted | 2026-03-26 |
| [ADR-004](ADR-004-database-schema-and-indexing.md) | Database Schema and Indexing | Accepted | 2026-03-26 |
