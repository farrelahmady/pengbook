package account_test

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"

	"pengbook/api/internal/module/account"
)

func TestGenerateAccountExport(t *testing.T) {
	rootID := int64(1)
	childID := int64(2)
	accounts := []account.Account{
		{ID: rootID, Code: "1.00.00.00", Name: "Aset", Type: account.AccountTypeAsset, Level: 0},
		{ID: childID, Code: "1.01.00.00", Name: "Aset Lancar", Type: account.AccountTypeAsset, Level: 1, ParentID: &rootID},
		{ID: 3, Code: "1.01.01.01", Name: "Kas", Type: account.AccountTypeAsset, Level: 3, ParentID: &childID},
	}

	exported, err := account.GenerateAccountExport(accounts)
	if err != nil {
		t.Fatalf("GenerateAccountExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export")
	}

	f, err := excelize.OpenReader(bytes.NewReader(exported))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer f.Close()

	sheet := "Daftar Akun"
	rows, err := f.GetRows(sheet)
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(rows) != len(accounts)+1 {
		t.Fatalf("expected %d rows, got %d", len(accounts)+1, len(rows))
	}

	wantHeaders := []string{"Kode", "Nama", "Tipe", "Level", "Posting", "Kode Induk"}
	for i, want := range wantHeaders {
		if rows[0][i] != want {
			t.Errorf("header col %d: expected %q, got %q", i, want, rows[0][i])
		}
	}

	// Posting row resolves Ya + parent code from the slice.
	postingRow := rows[3]
	if postingRow[0] != "1.01.01.01" {
		t.Errorf("expected code cell %q, got %q", "1.01.01.01", postingRow[0])
	}
	if postingRow[4] != "Ya" {
		t.Errorf("expected Posting %q, got %q", "Ya", postingRow[4])
	}
	if postingRow[5] != "1.01.00.00" {
		t.Errorf("expected Kode Induk %q, got %q", "1.01.00.00", postingRow[5])
	}

	// Header (non-posting) row resolves Tidak.
	if rows[1][4] != "Tidak" {
		t.Errorf("expected root Posting %q, got %q", "Tidak", rows[1][4])
	}
	// GetRows trims trailing empty cells, so a missing 6th cell means "".
	if cellAt(rows[1], 5) != "" {
		t.Errorf("expected empty Kode Induk for root, got %q", rows[1][5])
	}
}

// cellAt returns the cell value at index i, or "" when GetRows trimmed the
// trailing empty cells.
func cellAt(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return row[i]
}

func TestGenerateAccountExport_Empty(t *testing.T) {
	exported, err := account.GenerateAccountExport(nil)
	if err != nil {
		t.Fatalf("GenerateAccountExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export even without accounts")
	}
}
