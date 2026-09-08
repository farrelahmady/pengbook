-- +goose Up
CREATE TABLE IF NOT EXISTS journal_entry_lines (
    id                BIGSERIAL    PRIMARY KEY,
    journal_entry_id  BIGINT       NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id        BIGINT       NOT NULL REFERENCES accounts(id),
    debit             NUMERIC(19,4) NOT NULL DEFAULT 0,
    credit            NUMERIC(19,4) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    -- Setiap baris harus ada di debit ATAU credit (bukan keduanya, bukan nol)
    CONSTRAINT chk_journal_entry_lines_debit_credit CHECK (
        (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_journal_entry_lines_entry_id ON journal_entry_lines(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_entry_lines_account_id ON journal_entry_lines(account_id);

-- +goose Down
DROP TABLE IF EXISTS journal_entry_lines;
