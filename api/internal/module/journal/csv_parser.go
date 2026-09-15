package journal

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
)

// CSV column indices (0-indexed)
const (
	csvColDate        = 0
	csvColDescription = 1
	csvColAccountCode = 2
	csvColDebit       = 3
	csvColCredit      = 4
	csvMinColumns     = 5
)

// CSV date formats to try (in order of preference)
var csvDateFormats = []string{
	"2006-01-02",              // YYYY-MM-DD
	"02/01/2006",              // DD/MM/YYYY
	"01/02/2006",              // MM/DD/YYYY
	"2006-01-02T15:04:05Z07:00", // RFC3339
	time.RFC3339,
}

// ParseCSV parses a CSV file content and returns journal entry requests.
//
// CSV format:
//
//	Tanggal,Deskripsi,Kode Akun,Debit,Kredit
//	2026-04-21,Pembelian perlengkapan,5.01.01.04,500000,0
//	2026-04-21,Pembelian perlengkapan,1.01.02.01,0,500000
//
// Lines with the same date + description are grouped into one journal entry.
func ParseCSV(reader io.Reader) ([]CreateJournalRequest, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Find header row (skip BOM if present)
	startIdx := 0
	if len(records) > 0 {
		firstCell := strings.TrimSpace(records[0][0])
		// Check for BOM
		if len(firstCell) > 0 && firstCell[0] == 0xEF {
			firstCell = firstCell[1:]
		}
		// Skip header row if it looks like a header (contains "tanggal" or "date")
		lowerFirst := strings.ToLower(firstCell)
		if strings.Contains(lowerFirst, "tanggal") || strings.Contains(lowerFirst, "date") || strings.Contains(lowerFirst, "tanggal") {
			startIdx = 1
		}
	}

	if startIdx >= len(records) {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	// Group lines by date+description
	type entryKey struct {
		date        string
		description string
	}
	entryMap := make(map[entryKey]*CreateJournalRequest)
	entryOrder := make([]entryKey, 0) // preserve order

	for i, record := range records[startIdx:] {
		lineNum := startIdx + i + 1 // 1-indexed for error messages

		if len(record) < csvMinColumns {
			return nil, fmt.Errorf("row %d: expected at least %d columns, got %d", lineNum, csvMinColumns, len(record))
		}

		dateStr := strings.TrimSpace(record[csvColDate])
		description := strings.TrimSpace(record[csvColDescription])
		accountCode := strings.TrimSpace(record[csvColAccountCode])
		debitStr := strings.TrimSpace(record[csvColDebit])
		creditStr := strings.TrimSpace(record[csvColCredit])

		// Validate required fields
		if dateStr == "" {
			return nil, fmt.Errorf("row %d: date is required", lineNum)
		}
		if accountCode == "" {
			return nil, fmt.Errorf("row %d: account code is required", lineNum)
		}

		// Parse date
		parsedDate, err := parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid date %q: %w", lineNum, dateStr, err)
		}
		dateRFC3339 := parsedDate.Format(time.RFC3339)

		// Parse debit/credit
		debit, err := parseAmount(debitStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid debit amount %q: %w", lineNum, debitStr, err)
		}
		credit, err := parseAmount(creditStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid credit amount %q: %w", lineNum, creditStr, err)
		}

		// Skip empty rows
		if debit == 0 && credit == 0 {
			continue
		}

		// Group by date + description
		key := entryKey{date: dateRFC3339, description: description}
		if _, exists := entryMap[key]; !exists {
			entryMap[key] = &CreateJournalRequest{
				Date:        dateRFC3339,
				Description: description,
				Lines:       []JournalLineDto{},
			}
			entryOrder = append(entryOrder, key)
		}

		entry := entryMap[key]
		entry.Lines = append(entry.Lines, JournalLineDto{
			AccountCode: accountCode,
			Debit:       debit,
			Credit:      credit,
		})
	}

	if len(entryMap) == 0 {
		return nil, fmt.Errorf("no valid journal entries found in CSV")
	}

	// Build result in order
	entries := make([]CreateJournalRequest, 0, len(entryMap))
	for _, key := range entryOrder {
		entry := entryMap[key]
		if len(entry.Lines) < 2 {
			return nil, fmt.Errorf("entry on %s must have at least 2 lines, got %d", key.date, len(entry.Lines))
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

// parseDate tries multiple date formats and returns the parsed time.
func parseDate(s string) (time.Time, error) {
	for _, format := range csvDateFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format, expected YYYY-MM-DD")
}

// parseAmount parses a numeric string, handling commas as thousands separators.
func parseAmount(s string) (float64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	// Remove commas (thousands separator)
	s = strings.ReplaceAll(s, ",", "")

	var amount float64
	_, err := fmt.Sscanf(s, "%f", &amount)
	if err != nil {
		return 0, fmt.Errorf("not a valid number")
	}

	return amount, nil
}
