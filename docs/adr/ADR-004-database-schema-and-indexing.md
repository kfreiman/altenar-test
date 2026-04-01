# ADR-004: Database Schema and Indexing

**Status:** Accepted  
**Date:** 2026-03-26  
**Deciders:** kfreiman  

## Context

The system must handle high-volume transaction data (bets/wins). We require a schema that ensures financial precision, supports efficient filtering by `user_id` and `transaction_type`, and scales through data partitioning.

## Decision

We will use **PostgreSQL** with a partitioned schema and UUIDv7 identifiers.

### Schema and Types

```sql
CREATE TABLE transactions (
    id UUID NOT NULL DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Stored in minor units (e.g., cents)
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('bet', 'win')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    PRIMARY KEY (id)
) PARTITION BY RANGE (id);

-- Partition-local indexes for cursor-based pagination
CREATE INDEX idx_transactions_user_id ON transactions (user_id, id DESC);
CREATE INDEX idx_transactions_type_id ON transactions (transaction_type, id DESC);
```

### Partitioning Strategy

We use **UUIDv7-based range partitioning** for time-based data distribution:

* **UUIDv7 Advantage:** Naturally sortable by time. Native in **PostgreSQL 18+**; generated at the application level for older versions.
* **Retention:** Partitions created for **current month + 3 months ahead**.
* **Safety:** A `default` partition handles overflow or edge-case timestamps.
* **Automation:** Initial partitions via migrations; future partitions managed via cron.

### Query Patterns and Index Design

Queries utilize **cursor-based pagination** using the `id` as the cursor.

| Query | Index Used | Strategy |
| :--- | :--- | :--- |
| `ListTransactionsByUser` | `idx_transactions_user_id` | Filter by `user_id`, sort by `id DESC` |
| `ListTransactionsByType` | `idx_transactions_type_id` | Filter by `type`, sort by `id DESC` |
| `GetTransactionByID` | Primary Key | Direct lookup |

**Rationale for `id DESC` over `created_at`:**

1. UUIDv7 embeds time, making `ORDER BY id` identical to `ORDER BY created_at`.
2. Using `id` as the cursor covers both the filter and the sort in a single index, reducing storage and write overhead.

### Migrations

Use [dbmate](https://github.com/amacneil/dbmate) for version control. Migrations are stored in `internal/transactions/adapters/postgres/migrations`. It provides a simple SQL-first approach; [dbmate](https://github.com/amacneil/dbmate) automatically maintains the `schema.sql` file after applying migrations, which is subsequently used by [sqlc](https://sqlc.dev/) for Go code generation.

## Alternatives Considered

### Option 1: BIGINT + UUIDv7 Partitioning (Chosen)

* **Pros:** High precision, efficient time-based pruning, avoids floating-point errors, and simplifies indexing.
* **Cons:** Requires partition maintenance and PG18+ or app-side ID generation.

### Option 2: NUMERIC Type & Flat Table

* **Pros:** Infinite decimal precision.
* **Cons:** Slower performance than `BIGINT`; single-table growth leads to index bloat and maintenance downtime.

### Option 3: Composite Index on `created_at`

* **Pros:** Standard approach for many systems.
* **Cons:** Redundant indexing. Since UUIDv7 is sequential, indexing both `created_at` and `id` wastes I/O.

## Tradeoffs

* **Optimizing for:** Query speed, financial accuracy, and long-term scalability.
* **Sacrificing:** Simplicity (partitioning adds minor operational overhead).

## Consequences

### Positive

* **Pruning:** Queries for recent data only scan relevant partitions.
* **Performance:** `BIGINT` math is faster than `NUMERIC`.
* **Cleanliness:** Cursor-based pagination avoids the "offset" performance trap.

### Negative

* **Maintenance:** Requires a cron job or worker to ensure partitions exist ahead of time.
* **Strictness:** Application must ensure UUIDv7 format if using older Postgres versions.

## Risks

* **Partition Gap:** If the cron job fails, writes to the default partition may degrade performance.
  * *Mitigation:* Monitoring and alerts on partition coverage.
* **Precision:** `BIGINT` maxes at ~92 quadrillion units.
  * *Mitigation:* Sufficient for standard currencies; high-inflation tokens may require `NUMERIC`.

## Implementation Notes

* **Financials:** Never use `float` or `real`.
* **Postgres:** Prefer v18 for native `uuidv7()`.
* **Automation:** Partition creation must be idempotent.

## Follow-up Actions

* [x] Define schema (kfreiman)
* [x] Configure dbmate (kfreiman)
* [ ] Implement partition management cron job (kfreiman, TBD)
* [ ] Verify PG18 environment compatibility (kfreiman)

## References

* [UUIDv7 RFC 9562](https://www.rfc-editor.org/rfc/rfc9562)
* [dbmate Docs](https://github.com/amacneil/dbmate)
* [ADR-001: PostgreSQL Selection](ADR-001-choose-postgresql-for-transaction-storage.md)

## Revision History

* 2026-03-26: Initial decision.
* 2026-03-29: Updated for UUIDv7 partitioning and PG18 specifics.
