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
	"pengbook/api/pkg/logger"
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
	INSERT INTO journal_entries (user_id, datetime, description)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
`

const createLineQuery = `
	INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at
`

func (r *journalRepository) CreateEntry(ctx context.Context, entry *journal.JournalEntry) error {
	log := logger.FromContext(ctx)
	err := r.db(ctx).QueryRow(ctx, createEntryQuery,
		entry.UserID, entry.Date, entry.Description,
	).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		log.Error("repo: failed to create journal entry", "user_id", entry.UserID, "error", err)
	}
	return err
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
	SELECT id, user_id, datetime, description, created_at, updated_at
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
	log := logger.FromContext(ctx)
	var entry journal.JournalEntry
	err := r.db(ctx).QueryRow(ctx, findEntryByIDQuery, id).
		Scan(&entry.ID, &entry.UserID, &entry.Date, &entry.Description, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		log.Error("repo: failed to find journal entry by id", "entry_id", id, "error", err)
		return nil, err
	}

	// Fetch lines with account info
	rows, err := r.db(ctx).Query(ctx, findLinesByEntryIDQuery, id)
	if err != nil {
		log.Error("repo: failed to find journal lines by entry id", "entry_id", id, "error", err)
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

func (r *journalRepository) FindEntriesByUserIDWithNetEffect(ctx context.Context, userID int64, filter journal.EntryFilter) ([]journal.JournalEntryListItem, error) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	// Cursor filter
	if filter.CursorDatetime != nil && filter.CursorID != nil {
		where = append(where, fmt.Sprintf("(e.datetime, e.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, *filter.CursorDatetime, *filter.CursorID)
		argIdx += 2
	}

	// Date range filters
	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.datetime >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}
	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.datetime <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	// Account filter
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

	whereClause := strings.Join(where, " AND ")

	// Query 1: Get aggregated entries with net effect
	entryQuery := fmt.Sprintf(`
		SELECT 
			e.id,
			e.datetime,
			e.description,
			COALESCE(SUM(CASE WHEN a.type = 'ASSET' THEN l.debit ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN a.type = 'ASSET' THEN l.credit ELSE 0 END), 0) AS net_effect
		FROM journal_entries e
		LEFT JOIN journal_entry_lines l ON l.journal_entry_id = e.id
		LEFT JOIN accounts a ON a.id = l.account_id
		WHERE %s
		GROUP BY e.id, e.datetime, e.description
		ORDER BY e.datetime DESC, e.id DESC
		LIMIT $%d
	`, whereClause, argIdx)

	log := logger.FromContext(ctx)
	log.Info("repo: FindEntriesByUserIDWithNetEffect query", "query", entryQuery, "args", args)

	args = append(args, filter.Limit)

	rows, err := r.db(ctx).Query(ctx, entryQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]journal.JournalEntryListItem, 0)
	entryIDs := make([]int64, 0)

	for rows.Next() {
		var entry journal.JournalEntryListItem
		if err := rows.Scan(&entry.ID, &entry.Datetime, &entry.Description, &entry.NetEffect); err != nil {
			return nil, err
		}
		entry.Lines = make([]journal.JournalLineItem, 0)
		entries = append(entries, entry)
		entryIDs = append(entryIDs, entry.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(entryIDs) == 0 {
		return entries, nil
	}

	// Query 2: Get lines for all entries
	placeholders := make([]string, len(entryIDs))
	lineArgs := make([]interface{}, len(entryIDs))
	for i, id := range entryIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		lineArgs[i] = id
	}

	lineQuery := fmt.Sprintf(`
		SELECT 
			l.journal_entry_id,
			l.id,
			l.debit,
			l.credit,
			a.code || ' · ' || a.name AS account_display
		FROM journal_entry_lines l
		JOIN accounts a ON a.id = l.account_id
		WHERE l.journal_entry_id IN (%s)
		ORDER BY l.journal_entry_id, l.id
	`, strings.Join(placeholders, ","))

	lineRows, err := r.db(ctx).Query(ctx, lineQuery, lineArgs...)
	if err != nil {
		return nil, err
	}
	defer lineRows.Close()

	// Map entry ID to index
	entryMap := make(map[int64]int)
	for i, entry := range entries {
		entryMap[entry.ID] = i
	}

	for lineRows.Next() {
		var entryID int64
		var line journal.JournalLineItem
		if err := lineRows.Scan(&entryID, &line.ID, &line.Debit, &line.Credit, &line.AccountDisplay); err != nil {
			return nil, err
		}
		if idx, ok := entryMap[entryID]; ok {
			entries[idx].Lines = append(entries[idx].Lines, line)
		}
	}
	if err := lineRows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

const deleteEntryLinesQuery = `DELETE FROM journal_entry_lines WHERE journal_entry_id = $1`
const deleteEntryQuery = `DELETE FROM journal_entries WHERE id = $1`

func (r *journalRepository) DeleteEntry(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx)
	_, err := r.db(ctx).Exec(ctx, deleteEntryLinesQuery, id)
	if err != nil {
		log.Error("repo: failed to delete journal entry lines", "entry_id", id, "error", err)
		return err
	}
	_, err = r.db(ctx).Exec(ctx, deleteEntryQuery, id)
	if err != nil {
		log.Error("repo: failed to delete journal entry", "entry_id", id, "error", err)
	}
	return err
}

const updateEntryQuery = `
	UPDATE journal_entries
	SET datetime = $2, description = $3, updated_at = NOW()
	WHERE id = $1
	RETURNING updated_at
`

func (r *journalRepository) UpdateEntry(ctx context.Context, entry *journal.JournalEntry) error {
	log := logger.FromContext(ctx)

	// Delete existing lines
	_, err := r.db(ctx).Exec(ctx, deleteEntryLinesQuery, entry.ID)
	if err != nil {
		log.Error("repo: failed to delete journal entry lines", "entry_id", entry.ID, "error", err)
		return err
	}

	// Update entry
	err = r.db(ctx).QueryRow(ctx, updateEntryQuery,
		entry.ID, entry.Date, entry.Description,
	).Scan(&entry.UpdatedAt)
	if err != nil {
		log.Error("repo: failed to update journal entry", "entry_id", entry.ID, "error", err)
		return err
	}

	// Create new lines
	if err := r.createLines(ctx, entry.ID, entry.Lines); err != nil {
		log.Error("repo: failed to create journal entry lines", "entry_id", entry.ID, "error", err)
		return err
	}

	return nil
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

	// Cursor filter
	if filter.CursorDatetime != nil && filter.CursorID != nil {
		where = append(where, fmt.Sprintf("(e.datetime, e.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, *filter.CursorDatetime, *filter.CursorID)
		argIdx += 2
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.datetime >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.datetime <= $%d", argIdx))
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
		SELECT e.id, e.user_id, e.datetime, e.description, e.created_at, e.updated_at
		FROM journal_entries e
		WHERE %s
		ORDER BY e.datetime DESC, e.id DESC
		LIMIT $%d
	`, strings.Join(where, " AND "), argIdx)

	args = append(args, filter.Limit)

	return query, args
}

func (r *journalRepository) CountByUserIDWithFilter(ctx context.Context, userID int64, filter journal.EntryFilter) (int64, error) {
	where := []string{"e.user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.CursorDatetime != nil && filter.CursorID != nil {
		where = append(where, fmt.Sprintf("(e.datetime, e.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, *filter.CursorDatetime, *filter.CursorID)
		argIdx += 2
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.datetime >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.datetime <= $%d", argIdx))
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

	if filter.CursorDatetime != nil && filter.CursorID != nil {
		where = append(where, fmt.Sprintf("(e.datetime, e.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, *filter.CursorDatetime, *filter.CursorID)
		argIdx += 2
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.datetime >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.datetime <= $%d", argIdx))
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

	if filter.CursorDatetime != nil && filter.CursorID != nil {
		where = append(where, fmt.Sprintf("(e.datetime, e.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, *filter.CursorDatetime, *filter.CursorID)
		argIdx += 2
	}

	if filter.StartDate != nil {
		where = append(where, fmt.Sprintf("e.datetime >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		where = append(where, fmt.Sprintf("e.datetime <= $%d", argIdx))
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
