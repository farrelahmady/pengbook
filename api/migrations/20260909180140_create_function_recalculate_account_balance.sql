-- +goose Up
-- +goose statementbegin
CREATE OR REPLACE FUNCTION recalculate_account_balance(
    p_account_id BIGINT[] DEFAULT NULL,
    p_user_id    BIGINT[] DEFAULT NULL
)
RETURNS TABLE(account_id BIGINT, old_balance NUMERIC(19,4), new_balance NUMERIC(19,4), difference NUMERIC(19,4))
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    rec RECORD;
    calculated NUMERIC(19,4);
    sign_factor INT;
BEGIN
    FOR rec IN
        SELECT
            ab.account_id AS acc_id,
            ab.balance    AS cur_balance
        FROM account_balances ab
        INNER JOIN accounts a ON a.id = ab.account_id
        WHERE a.level = 3
          AND (p_account_id IS NULL OR ab.account_id = ANY(p_account_id))
          AND (p_user_id    IS NULL OR a.user_id    = ANY(p_user_id))
    LOOP
        -- Determine sign factor based on account type
        -- ASSET/EXPENSE/OTHER: debit - credit (positive = normal)
        -- LIABILITY/EQUITY/REVENUE: credit - debit (positive = normal)
        SELECT CASE a.type
            WHEN 'ASSET'     THEN 1
            WHEN 'EXPENSE'   THEN 1
            WHEN 'OTHER'     THEN 1
            WHEN 'LIABILITY' THEN -1
            WHEN 'EQUITY'    THEN -1
            WHEN 'REVENUE'   THEN -1
            ELSE 1
        END INTO sign_factor
        FROM accounts a
        WHERE a.id = rec.acc_id;

        -- Calculate balance from journal_entry_lines
        SELECT COALESCE(SUM(jel.debit - jel.credit) * sign_factor, 0)
        INTO calculated
        FROM journal_entry_lines jel
        WHERE jel.account_id = rec.acc_id;

        -- Update existing account_balances row
        UPDATE account_balances
        SET balance    = calculated,
            updated_at = NOW()
        WHERE account_balances.account_id = rec.acc_id;

        -- Return result row
        account_id   := rec.acc_id;
        old_balance  := rec.cur_balance;
        new_balance  := calculated;
        difference   := calculated - rec.cur_balance;
        RETURN NEXT;
    END LOOP;
END;
$$;
-- +goose statementend

-- +goose Down
DROP FUNCTION IF EXISTS recalculate_account_balance(BIGINT[], BIGINT[]);
