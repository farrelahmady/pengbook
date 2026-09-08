package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/journal"
)

type journalRepository struct {
	pool *pgxpool.Pool
}

func NewJournalRepository(pool *pgxpool.Pool) journal.Repository {
	return &journalRepository{pool: pool}
}

func (r *journalRepository) db(ctx context.Context) database.DBTX {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

const createEntryQuery = `
	INSERT INTO journal_entries (user_id, date, description)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
`

const createLineQuery = `
	INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at
`

func (r *journalRepository) CreateEntry(ctx context.Context, entry *journal.JournalEntry) error {
	return r.db(ctx).QueryRow(ctx, createEntryQuery,
		entry.UserID, entry.Date, entry.Description,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
}

func (r *journalRepository) createLines(ctx context.Context, entryID int64, lines []journal.JournalEntryLine) error {
	for i := range lines {
		lines[i].JournalEntryID = entryID
		err := r.db(ctx).QueryRow(ctx, createLineQuery,
			entryID, lines[i].AccountID, lines[i].Debit, lines[i].Credit,
		).Scan(&lines[i].ID, &lines[i].CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

const findEntryByIDQuery = `
	SELECT id, user_id, date, description, created_at, updated_at
	FROM journal_entries
	WHERE id = $1
`

const findLinesByEntryIDQuery = `
	SELECT l.id, l.journal_entry_id, l.account_id, l.debit, l.credit, l.created_at,
	       a.id, a.code, a.name, a.type, a.level, a.parent_id
	FROM journal_entry_lines l
	JOIN accounts a ON a.id = l.account_id
	WHERE l.journal_entry_id = $1
	ORDER BY l.id
`

func (r *journalRepository) FindEntryByID(ctx context.Context, id int64) (*journal.JournalEntry, error) {
	var entry journal.JournalEntry
	err := r.db(ctx).QueryRow(ctx, findEntryByIDQuery, id).
		Scan(&entry.ID, &entry.UserID, &entry.Date, &entry.Description, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Fetch lines with account info
	rows, err := r.db(ctx).Query(ctx, findLinesByEntryIDQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var line journal.JournalEntryLine
		var accountID int64
		var accountCode, accountName, accountType string
		var accountLevel int8
		var accountParentID *int64
		if err := rows.Scan(
			&line.ID, &line.JournalEntryID, &line.AccountID, &line.Debit, &line.Credit, &line.CreatedAt,
			&accountID, &accountCode, &accountName, &accountType, &accountLevel, &accountParentID,
		); err != nil {
			return nil, err
		}
		entry.Lines = append(entry.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *journalRepository) FindEntriesByUserID(ctx context.Context, userID int64, filter journal.EntryFilter) ([]journal.JournalEntry, error) {
	query, args := r.buildListQuery(userID, filter)
	rows, err := r.db(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []journal.JournalEntry
	for rows.Next() {
		var entry journal.JournalEntry
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Date, &entry.Description, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			return nil, err
		}

		// Fetch lines for each entry
		lineRows, err := r.db(ctx).Query(ctx, findLinesByEntryIDQuery, entry.ID)
		if err != nil {
			return nil, err
		}

		for lineRows.Next() {
			var line journal.JournalEntryLine
			var accountID int64
			var accountCode, accountName, accountType string
			var accountLevel int8
			var accountParentID *int64
			if err := lineRows.Scan(
				&line.ID, &line.JournalEntryID, &line.AccountID, &line.Debit, &line.Credit, &line.CreatedAt,
				&accountID, &accountCode, &accountName, &accountType, &accountLevel, &accountParentID,
			); err != nil {
				lineRows.Close()
				return nil, err
			}
			entry.Lines = append(entry.Lines, line)
		}
		lineRows.Close()

		if err := lineRows.Err(); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

const deleteEntryLinesQuery = `DELETE FROM journal_entry_lines WHERE journal_entry_id = $1`
const deleteEntryQuery = `DELETE FROM journal_entries WHERE id = $1`

func (r *journalRepository) DeleteEntry(ctx context.Context, id int64) error {
	_, err := r.db(ctx).Exec(ctx, deleteEntryLinesQuery, id)
	if err != nil {
		return err
	}
	_, err = r.db(ctx).Exec(ctx, deleteEntryQuery, id)
	return err
}

const updateEntryQuery = `
	UPDATE journal_entries
	SET date = $2, description = $3, updated_at = NOW()
	WHERE id = $1
	RETURNING updated_at
`

func (r *journalRepository) UpdateEntry(ctx context.Context, entry *journal.JournalEntry) error {
	// Delete existing lines
	_, err := r.db(ctx).Exec(ctx, deleteEntryLinesQuery, entry.ID)
	if err != nil {
		return err
	}

	// Update entry
	err = r.db(ctx).QueryRow(ctx, updateEntryQuery,
		entry.ID, entry.Date, entry.Description,
	).Scan(&entry.UpdatedAt)
	if err != nil {
		return err
	}

	// Create new lines
	return r.createLines(ctx, entry.ID, entry.Lines)
}

const journalCountByUserIDQuery = `SELECT COUNT(*) FROM journal_entries WHERE user_id = $1`

func (r *journalRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db(ctx).QueryRow(ctx, journalCountByUserIDQuery, userID).Scan(&count)
	return count, err
}

const sumDebitByUserIDQuery = `
	SELECT COALESCE(SUM(l.debit), 0)
	FROM journal_entry_lines l
	JOIN journal_entries e ON e.id = l.journal_entry_id
	WHERE e.user_id = $1
`

func (r *journalRepository) SumDebitByUserID(ctx context.Context, userID int64) (float64, error) {
	var sum float64
	err := r.db(ctx).QueryRow(ctx, sumDebitByUserIDQuery, userID).Scan(&sum)
	return sum, err
}

const sumCreditByUserIDQuery = `
	SELECT COALESCE(SUM(l.credit), 0)
	FROM journal_entry_lines l
	JOIN journal_entries e ON e.id = l.journal_entry_id
	WHERE e.user_id = $1
`

func (r *journalRepository) SumCreditByUserID(ctx context.Context, userID int64) (float64, error) {
	var sum float64
	err := r.db(ctx).QueryRow(ctx, sumCreditByUserIDQuery, userID).Scan(&sum)
	return sum, err
}

func (r *journalRepository) buildListQuery(userID int64, filter journal.EntryFilter) (string, []interface{}) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if len(filter.AccountIDs) > 0 {
		placeholders := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			"e.id IN (SELECT journal_entry_id FROM journal_entry_lines WHERE account_id IN (%s))",
			strings.Join(placeholders, ","),
		))
	}

	query := fmt.Sprintf(`
		SELECT e.id, e.user_id, e.date, e.description, e.created_at, e.updated_at
		FROM journal_entries e
		WHERE %s
		ORDER BY e.date DESC, e.id DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), argIdx, argIdx+1)

	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	return query, args
}

func (r *journalRepository) CountByUserIDWithFilter(ctx context.Context, userID int64, filter journal.EntryFilter) (int64, error) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if len(filter.AccountIDs) > 0 {
		placeholders := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			"e.id IN (SELECT journal_entry_id FROM journal_entry_lines WHERE account_id IN (%s))",
			strings.Join(placeholders, ","),
		))
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM journal_entries e
		WHERE %s
	`, strings.Join(where, " AND "))

	var count int64
	err := r.db(ctx).QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *journalRepository) SumDebitByUserIDWithFilter(ctx context.Context, userID int64, filter journal.EntryFilter) (float64, error) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if len(filter.AccountIDs) > 0 {
		placeholders := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			"e.id IN (SELECT journal_entry_id FROM journal_entry_lines WHERE account_id IN (%s))",
			strings.Join(placeholders, ","),
		))
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(l.debit), 0)
		FROM journal_entry_lines l
		JOIN journal_entries e ON e.id = l.journal_entry_id
		WHERE %s
	`, strings.Join(where, " AND "))

	var sum float64
	err := r.db(ctx).QueryRow(ctx, query, args...).Scan(&sum)
	return sum, err
}

func (r *journalRepository) SumCreditByUserIDWithFilter(ctx context.Context, userID int64, filter journal.EntryFilter) (float64, error) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.date >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.date <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if len(filter.AccountIDs) > 0 {
		placeholders := make([]string, len(filter.AccountIDs))
		for i, id := range filter.AccountIDs {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, id)
			argIdx++
		}
		where = append(where, fmt.Sprintf(
			"e.id IN (SELECT journal_entry_id FROM journal_entry_lines WHERE account_id IN (%s))",
			strings.Join(placeholders, ","),
		))
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(l.credit), 0)
		FROM journal_entry_lines l
		JOIN journal_entries e ON e.id = l.journal_entry_id
		WHERE %s
	`, strings.Join(where, " AND "))

	var sum float64
	err := r.db(ctx).QueryRow(ctx, query, args...).Scan(&sum)
	return sum, err
}
