package journal_test

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"

	"pengbook/api/internal/module/journal"
)

// createTestExcel creates a test Excel file with journal data.
func createTestExcel(t *testing.T, rows [][]interface{}) []byte {
	t.Helper()

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// Header
	headers := []string{"Tanggal", "Deskripsi", "Kode Akun", "Debit", "Kredit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Data rows
	for r, row := range rows {
		for c, val := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create test Excel: %v", err)
	}

	return buffer.Bytes()
}

func TestParseExcel_Success(t *testing.T) {
	rows := [][]interface{}{
		{"2026-04-21", "Pembelian perlengkapan", "5.01.01.04", 500000, 0},
		{"2026-04-21", "Pembelian perlengkapan", "1.01.02.01", 0, 500000},
	}
	excelData := createTestExcel(t, rows)

	entries, err := journal.ParseExcel(bytes.NewReader(excelData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Description != "Pembelian perlengkapan" {
		t.Errorf("expected description 'Pembelian perlengkapan', got %s", entry.Description)
	}
	if len(entry.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(entry.Lines))
	}
	if entry.Lines[0].AccountCode != "5.01.01.04" {
		t.Errorf("expected account code '5.01.01.04', got %s", entry.Lines[0].AccountCode)
	}
	if entry.Lines[0].Debit != 500000 {
		t.Errorf("expected debit 500000, got %f", entry.Lines[0].Debit)
	}
	if entry.Lines[1].Credit != 500000 {
		t.Errorf("expected credit 500000, got %f", entry.Lines[1].Credit)
	}
}

func TestParseExcel_MultipleEntries(t *testing.T) {
	rows := [][]interface{}{
		{"2026-04-21", "Entry 1", "5.01.01.04", 100000, 0},
		{"2026-04-21", "Entry 1", "1.01.02.01", 0, 100000},
		{"2026-04-22", "Entry 2", "5.01.01.04", 200000, 0},
		{"2026-04-22", "Entry 2", "1.01.02.01", 0, 200000},
	}
	excelData := createTestExcel(t, rows)

	entries, err := journal.ParseExcel(bytes.NewReader(excelData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Description != "Entry 1" {
		t.Errorf("expected first entry description 'Entry 1', got %s", entries[0].Description)
	}
	if entries[1].Description != "Entry 2" {
		t.Errorf("expected second entry description 'Entry 2', got %s", entries[1].Description)
	}
}

func TestParseExcel_EmptySheet(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create test Excel: %v", err)
	}

	_, err = journal.ParseExcel(bytes.NewReader(buffer.Bytes()))
	if err == nil {
		t.Fatal("expected error for empty sheet")
	}
}

func TestParseExcel_HeaderOnly(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// Only header, no data
	headers := []string{"Tanggal", "Deskripsi", "Kode Akun", "Debit", "Kredit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create test Excel: %v", err)
	}

	_, err = journal.ParseExcel(bytes.NewReader(buffer.Bytes()))
	if err == nil {
		t.Fatal("expected error for header-only sheet")
	}
}

func TestParseExcel_MissingDate(t *testing.T) {
	rows := [][]interface{}{
		{"", "Test", "5.01.01.04", 100000, 0},
		{"2026-04-21", "Test", "1.01.02.01", 0, 100000},
	}
	excelData := createTestExcel(t, rows)

	_, err := journal.ParseExcel(bytes.NewReader(excelData))
	if err == nil {
		t.Fatal("expected error for missing date")
	}
}

func TestParseExcel_MissingAccountCode(t *testing.T) {
	rows := [][]interface{}{
		{"2026-04-21", "Test", "", 100000, 0},
		{"2026-04-21", "Test", "1.01.02.01", 0, 100000},
	}
	excelData := createTestExcel(t, rows)

	_, err := journal.ParseExcel(bytes.NewReader(excelData))
	if err == nil {
		t.Fatal("expected error for missing account code")
	}
}

func TestParseExcel_TooFewLines(t *testing.T) {
	rows := [][]interface{}{
		{"2026-04-21", "Test", "5.01.01.04", 100000, 0},
	}
	excelData := createTestExcel(t, rows)

	_, err := journal.ParseExcel(bytes.NewReader(excelData))
	if err == nil {
		t.Fatal("expected error for entry with only 1 line")
	}
}

func TestParseExcel_WithHeaderRow(t *testing.T) {
	// Excel with explicit header row that contains "Tanggal"
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Template Jurnal"
	f.SetSheetName("Sheet1", sheet)

	// Header row
	f.SetCellValue(sheet, "A1", "Tanggal")
	f.SetCellValue(sheet, "B1", "Deskripsi")
	f.SetCellValue(sheet, "C1", "Kode Akun")
	f.SetCellValue(sheet, "D1", "Debit")
	f.SetCellValue(sheet, "E1", "Kredit")

	// Data rows
	f.SetCellValue(sheet, "A2", "2026-04-21")
	f.SetCellValue(sheet, "B2", "Test")
	f.SetCellValue(sheet, "C2", "5.01.01.04")
	f.SetCellValue(sheet, "D2", 100000)
	f.SetCellValue(sheet, "E2", 0)

	f.SetCellValue(sheet, "A3", "2026-04-21")
	f.SetCellValue(sheet, "B3", "Test")
	f.SetCellValue(sheet, "C3", "1.01.02.01")
	f.SetCellValue(sheet, "D3", 0)
	f.SetCellValue(sheet, "E3", 100000)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("failed to create test Excel: %v", err)
	}

	entries, err := journal.ParseExcel(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestGenerateTemplate(t *testing.T) {
	template, err := journal.GenerateTemplate("1.01.01.01", "Kas - Main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(template) == 0 {
		t.Fatal("expected non-empty template")
	}

	// Verify it's a valid Excel file
	f, err := excelize.OpenReader(bytes.NewReader(template))
	if err != nil {
		t.Fatalf("failed to open generated template: %v", err)
	}
	defer f.Close()

	// Check sheets
	sheets := f.GetSheetList()
	if len(sheets) != 2 {
		t.Fatalf("expected 2 sheets, got %d", len(sheets))
	}

	// Check Sheet 1 name
	if sheets[0] != "Template Jurnal" {
		t.Errorf("expected first sheet 'Template Jurnal', got %s", sheets[0])
	}

	// Check Sheet 2 name
	if sheets[1] != "Daftar Akun" {
		t.Errorf("expected second sheet 'Daftar Akun', got %s", sheets[1])
	}
}

func TestGenerateTemplateWithAccounts(t *testing.T) {
	accounts := []journal.AccountInfo{
		{Code: "1.01.01.01", Name: "Kas - Main", Type: "ASSET", Level: 3},
		{Code: "4.01.01.01", Name: "Pendapatan Jasa", Type: "REVENUE", Level: 3},
	}

	template, err := journal.GenerateTemplateWithAccounts(accounts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(template) == 0 {
		t.Fatal("expected non-empty template")
	}

	// Verify it's a valid Excel file
	f, err := excelize.OpenReader(bytes.NewReader(template))
	if err != nil {
		t.Fatalf("failed to open generated template: %v", err)
	}
	defer f.Close()

	// Check Sheet 2 has accounts
	sheet2 := "Daftar Akun"
	rows, err := f.GetRows(sheet2)
	if err != nil {
		t.Fatalf("failed to read sheet: %v", err)
	}

	// Header + 2 accounts = 3 rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows in Daftar Akun, got %d", len(rows))
	}

	// Check first account
	if rows[1][0] != "1.01.01.01" {
		t.Errorf("expected first account code '1.01.01.01', got %s", rows[1][0])
	}
}
