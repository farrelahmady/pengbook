-- +goose Up
CREATE TABLE IF NOT EXISTS accounts_audit_logs (
    id            BIGSERIAL   PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id    BIGINT      REFERENCES accounts(id) ON DELETE SET NULL,
    action        TEXT        NOT NULL,
    old_values    JSONB,
    new_values    JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_audit_logs_user_id ON accounts_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_audit_logs_account_id ON accounts_audit_logs(account_id);

-- +goose Down
DROP TABLE IF EXISTS accounts_audit_logs;
