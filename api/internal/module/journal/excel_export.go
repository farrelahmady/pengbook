package journal

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// exportHeaderStyle returns the shared header style used by both the template
// and the export (bold, blue fill, centered).
func exportHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
}

// writeAccountSheet fills the "Daftar Akun" sheet with the user's posting
// accounts. Shared layout with GenerateTemplateWithAccounts.
func writeAccountSheet(f *excelize.File, accounts []AccountInfo) error {
	sheet := "Daftar Akun"
	if _, err := f.NewSheet(sheet); err != nil {
		return fmt.Errorf("failed to create sheet %q: %w", sheet, err)
	}

	accHeaders := []string{"Kode Akun", "Nama Akun", "Tipe", "Level", "Posting"}
	for i, h := range accHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return err
		}
	}

	headerStyle, err := exportHeaderStyle(f)
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheet, "A1", "E1", headerStyle); err != nil {
		return err
	}

	for i, acc := range accounts {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), acc.Code)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), acc.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), acc.Type)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), acc.Level)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "Ya")
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 8)
	f.SetColWidth(sheet, "E", "E", 10)
	return nil
}

// GenerateExport creates an Excel file with the user's journal data that can
// be re-uploaded as-is: Sheet 1 "Template Jurnal" has the same columns as the
// upload format (Tanggal, Deskripsi, Kode Akun, Debit, Kredit), Sheet 2
// "Daftar Akun" lists the user's posting accounts.
func GenerateExport(rows []ExportRow, accounts []AccountInfo) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Sheet 1: Template Jurnal (same layout as the upload template)
	sheet1 := "Template Jurnal"
	if err := f.SetSheetName("Sheet1", sheet1); err != nil {
		return nil, fmt.Errorf("failed to rename sheet: %w", err)
	}

	headers := []string{"Tanggal", "Deskripsi", "Kode Akun", "Debit", "Kredit"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet1, cell, h); err != nil {
			return nil, err
		}
	}

	headerStyle, err := exportHeaderStyle(f)
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet1, "A1", "E1", headerStyle); err != nil {
		return nil, err
	}

	// Dates are written as YYYY-MM-DD strings (same as the template examples):
	// unambiguous for parseExcelDate and immune to Excel serial quirks.
	// Times are rendered in Asia/Jakarta (the app's audience) with UTC fallback.
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}
	for r, row := range rows {
		excelRow := r + 2
		f.SetCellValue(sheet1, fmt.Sprintf("A%d", excelRow), row.Datetime.In(loc).Format("2006-01-02"))
		f.SetCellValue(sheet1, fmt.Sprintf("B%d", excelRow), row.Description)
		f.SetCellValue(sheet1, fmt.Sprintf("C%d", excelRow), row.AccountCode)
		f.SetCellValue(sheet1, fmt.Sprintf("D%d", excelRow), row.Debit)
		f.SetCellValue(sheet1, fmt.Sprintf("E%d", excelRow), row.Credit)
	}

	f.SetColWidth(sheet1, "A", "A", 15)
	f.SetColWidth(sheet1, "B", "B", 25)
	f.SetColWidth(sheet1, "C", "C", 15)
	f.SetColWidth(sheet1, "D", "D", 15)
	f.SetColWidth(sheet1, "E", "E", 15)

	// Sheet 2: Daftar Akun
	if err := writeAccountSheet(f, accounts); err != nil {
		return nil, err
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Excel export: %w", err)
	}
	return buffer.Bytes(), nil
}
