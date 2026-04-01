-- migrate:up
-- 1. Create the parent table
-- Partitioning by 'id' allows 'id' to be the sole PRIMARY KEY.
CREATE TABLE transactions (
    id UUID NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('bet', 'win')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    PRIMARY KEY (id)
) PARTITION BY RANGE (id);
-- 2. Create indexes on the parent table
-- Note: Local indexes in partitions will automatically include 'id' logic
CREATE INDEX idx_transactions_user_id ON transactions (user_id, id DESC);
CREATE INDEX idx_transactions_type_id ON transactions (transaction_type, id DESC);
-- 3. Create the Default partition (Safety bucket)
CREATE TABLE transactions_default PARTITION OF transactions DEFAULT;
-- 4. Dynamic Partition Generation
-- This block creates partitions for the current month + the next 3 months (4 months total coverage)
-- TODO: create partitions using queries (not migrations) to allow execution via cron job
DO $$
DECLARE start_date TIMESTAMP;
end_date TIMESTAMP;
partition_name TEXT;
-- Helper to turn a timestamp into a 'start-of-period' UUIDv7
-- UUIDv7: [48 bits timestamp][4 bits version 7][12 bits 0][2 bits variant 10][62 bits 0]
-- Hex:    [12 chars time]   [7]              [000]   [8]              [000...]
make_uuid_boundary CONSTANT TEXT := 'lpad(to_hex((extract(epoch from %L::timestamp) * 1000)::bigint), 12, ''0'') || ''70008000000000000000''';
start_uuid UUID;
end_uuid UUID;
BEGIN FOR i IN 0..3 LOOP start_date := date_trunc('month', now() + (i * interval '1 month'));
end_date := start_date + interval '1 month';
-- Generate boundaries
EXECUTE format(
    'SELECT (' || make_uuid_boundary || ')::uuid',
    start_date
) INTO start_uuid;
EXECUTE format(
    'SELECT (' || make_uuid_boundary || ')::uuid',
    end_date
) INTO end_uuid;
partition_name := 'transactions_y' || to_char(start_date, 'YYYY') || '_m' || to_char(start_date, 'MM');
EXECUTE format(
    'CREATE TABLE IF NOT EXISTS %I PARTITION OF transactions FOR VALUES FROM (%L) TO (%L)',
    partition_name,
    start_uuid,
    end_uuid
);
END LOOP;
END $$;
-- migrate:down
DROP TABLE IF EXISTS transactions CASCADE;