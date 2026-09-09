-- +goose Up
-- Seed account_balances for all Asset Level 3 (posting) accounts
-- Balance awal = 0, akan dihitung oleh function recalculate_account_balance

INSERT INTO account_balances (account_id, balance)
SELECT id, 0
FROM accounts
WHERE user_id = (SELECT id FROM users WHERE username = 'admin')
  AND type = 'ASSET'
  AND level = 3
ON CONFLICT (account_id) DO NOTHING;

SELECT * FROM recalculate_account_balance(NULL, ARRAY[(SELECT id FROM users WHERE username = 'admin')]);

-- +goose Down
DELETE FROM account_balances
WHERE account_id IN (
    SELECT id FROM accounts
    WHERE user_id = (SELECT id FROM users WHERE username = 'admin')
      AND type = 'ASSET'
      AND level = 3
);
