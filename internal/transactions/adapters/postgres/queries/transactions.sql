-- name: CreateTransaction :exec
INSERT INTO transactions (
        id,
        user_id,
        transaction_type,
        amount,
        created_at
    )
VALUES (
        @id,
        @user_id,
        @transaction_type,
        @amount,
        @created_at
    );
-- name: ListAllTransactions :many
-- Fetches all transactions, newest first.
-- Cursor: The last 'id' from the previous page (use uuid.Nil for the first page).
SELECT id,
    user_id,
    transaction_type,
    amount,
    created_at
FROM transactions
WHERE -- 1. Handle Pagination: If cursor is Nil, ignore the ID filter.
    -- Otherwise, only show records older (smaller) than the cursor.
    (
        @cursor_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid
        OR id < @cursor_id::uuid
    ) -- 2. Optional User ID filter
    AND (
        sqlc.narg('user_id')::varchar IS NULL
        OR user_id = sqlc.narg('user_id')::varchar
    ) -- 3. Optional Transaction Type filter
    AND (
        sqlc.narg('transaction_type')::varchar IS NULL
        OR transaction_type = sqlc.narg('transaction_type')::varchar
    )
ORDER BY id DESC
LIMIT @page_size;