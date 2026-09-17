package account

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// GenerateAssetExport creates an Excel file listing the user's asset posting
// accounts, ordered by code (the input groups preserve code order from
// FindAssetWithBalances). Columns: Kode | Nama | Kelompok | Saldo.
// Kelompok is the level-2 ancestor name (e.g. Bank, Kas), Saldo is the cached
// balance from account_balances. Balances are calculated server-side; the
// frontend only downloads the finished file.
func GenerateAssetExport(groups []AssetGroup) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Daftar Aset"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, fmt.Errorf("failed to rename sheet: %w", err)
	}

	headers := []string{"Kode", "Nama", "Kelompok", "Saldo"}
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
	if err := f.SetCellStyle(sheet, "A1", "D1", headerStyle); err != nil {
		return nil, err
	}

	row := 2
	for _, g := range groups {
		for _, a := range g.Accounts {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.Code)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.Name)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), g.Name)
			f.SetCellValue(sheet, fmt.Sprintf("D%d", row), a.Balance)
			row++
		}
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 18)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to generate asset Excel export: %w", err)
	}
	return buffer.Bytes(), nil
}
