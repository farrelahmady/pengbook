package journal

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ParseExcel parses an Excel (.xlsx) file and returns journal entry requests.
//
// Expected format (Sheet 1 - "Template Jurnal"):
//
//	| Tanggal    | Deskripsi          | Kode Akun  | Debit   | Kredit  |
//	|------------|--------------------|------------|---------|---------|
//	| 2026-04-21 | Pembelian perlengkapan | 5.01.01.04 | 500000 | 0       |
//	| 2026-04-21 | Pembelian perlengkapan | 1.01.02.01 | 0      | 500000  |
func ParseExcel(reader io.Reader) ([]CreateJournalRequest, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get first sheet name
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel file has no sheets")
	}

	sheetName := sheets[0]

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %q: %w", sheetName, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("Excel sheet is empty")
	}

	// Find header row (skip if it looks like a header)
	startIdx := 0
	if len(rows) > 0 {
		firstCell := strings.TrimSpace(rows[0][0])
		lowerFirst := strings.ToLower(firstCell)
		if strings.Contains(lowerFirst, "tanggal") || strings.Contains(lowerFirst, "date") {
			startIdx = 1
		}
	}

	if startIdx >= len(rows) {
		return nil, fmt.Errorf("Excel file has no data rows")
	}

	// Group lines by date+description
	type entryKey struct {
		date        string
		description string
	}
	entryMap := make(map[entryKey]*CreateJournalRequest)
	entryOrder := make([]entryKey, 0)

	for i, row := range rows[startIdx:] {
		lineNum := startIdx + i + 1 // 1-indexed for error messages

		// Ensure row has enough columns
		if len(row) < 5 {
			return nil, fmt.Errorf("row %d: expected at least 5 columns, got %d", lineNum, len(row))
		}

		dateStr := strings.TrimSpace(row[0])
		description := strings.TrimSpace(row[1])
		accountCode := strings.TrimSpace(row[2])
		debitStr := strings.TrimSpace(row[3])
		creditStr := strings.TrimSpace(row[4])

		// Validate required fields
		if dateStr == "" {
			return nil, fmt.Errorf("row %d: date is required", lineNum)
		}
		if accountCode == "" {
			return nil, fmt.Errorf("row %d: account code is required", lineNum)
		}

		// Parse date
		parsedDate, err := parseExcelDate(dateStr)
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
		return nil, fmt.Errorf("no valid journal entries found in Excel")
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

// parseExcelDate parses dates from Excel, handling both string dates and Excel serial numbers.
func parseExcelDate(s string) (time.Time, error) {
	// Try common date formats first
	formats := []string{
		"2006-01-02",              // YYYY-MM-DD
		"02/01/2006",              // DD/MM/YYYY
		"01/02/2006",              // MM/DD/YYYY
		"2006-01-02T15:04:05Z07:00", // RFC3339
		time.RFC3339,
		"02-01-2006",              // DD-MM-YYYY
		"01-02-2006",              // MM-DD-YYYY
		"2006/01/02",              // YYYY/MM/DD
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	// Try parsing as Excel serial number
	var serial float64
	if _, err := fmt.Sscanf(s, "%f", &serial); err == nil && serial > 0 {
		// Excel serial number: days since 1900-01-01 (with leap year bug)
		baseDate := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		days := int(serial)
		return baseDate.AddDate(0, 0, days), nil
	}

	return time.Time{}, fmt.Errorf("unsupported date format, expected YYYY-MM-DD")
}

// GenerateTemplate creates an Excel template with 2 sheets:
//   - Sheet 1: "Template Jurnal" - journal entry template
//   - Sheet 2: "Daftar Akun" - list of posting accounts
func GenerateTemplate(accountCode string, accountName string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Template Jurnal
	sheet1 := "Template Jurnal"
	f.SetSheetName("Sheet1", sheet1)

	// Header
	headers := []string{"Tanggal", "Deskripsi", "Kode Akun", "Debit", "Kredit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet1, cell, h)
	}

	// Style header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet1, "A1", "E1", headerStyle)

	// Example rows
	if accountCode != "" {
		// Find a second account code (use first 2 characters + different suffix)
		secondCode := "4.01.01.01" // Default revenue account
		if strings.HasPrefix(accountCode, "4.") {
			secondCode = "1.01.01.01" // Use asset account if first is revenue
		}

		exampleRows := [][]interface{}{
			{"2026-04-21", "Pembelian perlengkapan", accountCode, 500000, 0},
			{"2026-04-21", "Pembelian perlengkapan", secondCode, 0, 500000},
		}

		for r, row := range exampleRows {
			for c, val := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
				f.SetCellValue(sheet1, cell, val)
			}
		}
	}

	// Set column widths
	f.SetColWidth(sheet1, "A", "A", 15) // Tanggal
	f.SetColWidth(sheet1, "B", "B", 25) // Deskripsi
	f.SetColWidth(sheet1, "C", "C", 15) // Kode Akun
	f.SetColWidth(sheet1, "D", "D", 15) // Debit
	f.SetColWidth(sheet1, "E", "E", 15) // Kredit

	// Sheet 2: Daftar Akun
	sheet2 := "Daftar Akun"
	f.NewSheet(sheet2)

	// Header
	accHeaders := []string{"Kode Akun", "Nama Akun", "Tipe", "Level", "Posting"}
	for i, h := range accHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet2, cell, h)
	}

	// Style header
	f.SetCellStyle(sheet2, "A1", "E1", headerStyle)

	// Column widths
	f.SetColWidth(sheet2, "A", "A", 15) // Kode Akun
	f.SetColWidth(sheet2, "B", "B", 30) // Nama Akun
	f.SetColWidth(sheet2, "C", "C", 12) // Tipe
	f.SetColWidth(sheet2, "D", "D", 8)  // Level
	f.SetColWidth(sheet2, "E", "E", 10) // Posting

	// Save to buffer
	var buf []byte
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Excel: %w", err)
	}
	buf = buffer.Bytes()

	return buf, nil
}

// GenerateTemplateWithAccounts creates an Excel template with user's accounts.
func GenerateTemplateWithAccounts(accounts []AccountInfo) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Template Jurnal
	sheet1 := "Template Jurnal"
	f.SetSheetName("Sheet1", sheet1)

	// Header
	headers := []string{"Tanggal", "Deskripsi", "Kode Akun", "Debit", "Kredit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet1, cell, h)
	}

	// Style header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet1, "A1", "E1", headerStyle)

	// Example rows with first available account
	if len(accounts) > 0 {
		firstCode := accounts[0].Code
		secondCode := "4.01.01.01" // Default
		if len(accounts) > 1 {
			secondCode = accounts[1].Code
		}

		exampleRows := [][]interface{}{
			{"2026-04-21", "Pembelian perlengkapan", firstCode, 500000, 0},
			{"2026-04-21", "Pembelian perlengkapan", secondCode, 0, 500000},
		}

		for r, row := range exampleRows {
			for c, val := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
				f.SetCellValue(sheet1, cell, val)
			}
		}
	}

	// Set column widths
	f.SetColWidth(sheet1, "A", "A", 15)
	f.SetColWidth(sheet1, "B", "B", 25)
	f.SetColWidth(sheet1, "C", "C", 15)
	f.SetColWidth(sheet1, "D", "D", 15)
	f.SetColWidth(sheet1, "E", "E", 15)

	// Sheet 2: Daftar Akun
	sheet2 := "Daftar Akun"
	f.NewSheet(sheet2)

	// Header
	accHeaders := []string{"Kode Akun", "Nama Akun", "Tipe", "Level", "Posting"}
	for i, h := range accHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet2, cell, h)
	}

	// Style header
	f.SetCellStyle(sheet2, "A1", "E1", headerStyle)

	// Add accounts
	for i, acc := range accounts {
		row := i + 2
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", row), acc.Code)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", row), acc.Name)
		f.SetCellValue(sheet2, fmt.Sprintf("C%d", row), acc.Type)
		f.SetCellValue(sheet2, fmt.Sprintf("D%d", row), acc.Level)
		f.SetCellValue(sheet2, fmt.Sprintf("E%d", row), "Ya")
	}

	// Set column widths
	f.SetColWidth(sheet2, "A", "A", 15)
	f.SetColWidth(sheet2, "B", "B", 30)
	f.SetColWidth(sheet2, "C", "C", 12)
	f.SetColWidth(sheet2, "D", "D", 8)
	f.SetColWidth(sheet2, "E", "E", 10)

	// Save to buffer
	var buf []byte
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Excel: %w", err)
	}
	buf = buffer.Bytes()

	return buf, nil
}

// AccountInfo is a minimal account info for template generation.
type AccountInfo struct {
	Code  string
	Name  string
	Type  string
	Level int8
}
