-- +goose Up
CREATE TABLE IF NOT EXISTS journal_audit_logs (
    id               BIGSERIAL   PRIMARY KEY,
    user_id          BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    journal_entry_id BIGINT      REFERENCES journal_entries(id) ON DELETE SET NULL,
    action           TEXT        NOT NULL,
    old_values       JSONB,
    new_values       JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_journal_audit_logs_user_id ON journal_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_journal_audit_logs_entry_id ON journal_audit_logs(journal_entry_id);

-- +goose Down
DROP TABLE IF EXISTS journal_audit_logs;
