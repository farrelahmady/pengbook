-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_journal_entries_user_datetime ON journal_entries(user_id, datetime DESC, id DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_journal_entries_user_datetime;
-- +goose StatementEnd
