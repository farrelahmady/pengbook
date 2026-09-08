-- +goose Up
CREATE TABLE IF NOT EXISTS account_balances (
    account_id      BIGINT        PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    balance         NUMERIC(19,4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS account_balances;
