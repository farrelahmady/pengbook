package account

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// accountExportHeaderStyle returns the header style for the account export
// (bold, blue fill, centered). Kept local to this package; the journal
// package has its own copy.
func accountExportHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
}

// GenerateAccountExport creates an Excel file listing the user's accounts,
// ordered by code (the input slice from FindByUserID is already ordered).
// Columns: Kode | Nama | Tipe | Level | Posting | Kode Induk.
// Parent codes are resolved from the slice itself via an id->code map,
// so no extra query is needed. No balance column.
func GenerateAccountExport(accounts []Account) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Daftar Akun"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("failed to rename sheet: %w", err)
	}

	headers := []string{"Kode", "Nama", "Tipe", "Level", "Posting", "Kode Induk"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, err
		}
	}

	headerStyle, err := accountExportHeaderStyle(f)
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A1", "F1", headerStyle); err != nil {
		return nil, err
	}

	idToCode := make(map[int64]string, len(accounts))
	for _, a := range accounts {
		idToCode[a.ID] = a.Code
	}

	for i, a := range accounts {
		row := i + 2
		posting := "Tidak"
		if a.Level == 3 {
			posting = "Ya"
		}
		parentCode := ""
		if a.ParentID != nil {
			parentCode = idToCode[*a.ParentID]
		}
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.Code)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), string(a.Type))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), a.Level)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), posting)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), parentCode)
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 8)
	f.SetColWidth(sheet, "E", "E", 10)
	f.SetColWidth(sheet, "F", "F", 15)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate account Excel export: %w", err)
	}
	return buffer.Bytes(), nil
}
