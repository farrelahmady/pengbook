-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
    id            BIGSERIAL   PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code          TEXT        NOT NULL,
    name          TEXT        NOT NULL,

    -- Generated columns (derived from code)
    -- type: digit pertama code -> ASSET/LIABILITY/EQUITY/REVENUE/EXPENSE/OTHER
    type          TEXT        NOT NULL GENERATED ALWAYS AS (
        CASE substr(code, 1, 1)
            WHEN '1' THEN 'ASSET'
            WHEN '2' THEN 'LIABILITY'
            WHEN '3' THEN 'EQUITY'
            WHEN '4' THEN 'REVENUE'
            WHEN '5' THEN 'EXPENSE'
            WHEN '6' THEN 'OTHER'
        END
    ) STORED,

    -- level: hitung segmen non-00 setelah digit pertama (0-3)
    level         SMALLINT    NOT NULL GENERATED ALWAYS AS (
        (CASE WHEN split_part(code, '.', 2) != '00' THEN 1 ELSE 0 END) +
        (CASE WHEN split_part(code, '.', 3) != '00' THEN 1 ELSE 0 END) +
        (CASE WHEN split_part(code, '.', 4) != '00' THEN 1 ELSE 0 END)
    ) STORED,

    parent_id     BIGINT      REFERENCES accounts(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Validasi format code: X.YY.ZZ.WW (X=1-6, YY/ZZ/WW=00-99)
    CONSTRAINT chk_accounts_code_format CHECK (code ~ '^[1-6]\.\d{2}\.\d{2}\.\d{2}$'),

    -- Code unik per user (bukan global)
    CONSTRAINT uq_accounts_user_code UNIQUE (user_id, code)
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_parent_id ON accounts(parent_id);

-- +goose Down
DROP TABLE IF EXISTS accounts;
