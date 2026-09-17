package account_test

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"

	"pengbook/api/internal/module/account"
)

func TestGenerateAssetExport(t *testing.T) {
	groups := []account.AssetGroup{
		{
			ID:   10,
			Code: "1.01.01.00",
			Name: "Bank",
			Accounts: []account.AssetAccount{
				{ID: 11, Code: "1.01.01.01", Name: "Mandiri", Balance: 1500000, IsPosting: true},
				{ID: 12, Code: "1.01.01.02", Name: "BCA", Balance: 500000, IsPosting: true},
			},
			TotalBalance: 2000000,
		},
	}

	exported, err := account.GenerateAssetExport(groups)
	if err != nil {
		t.Fatalf("GenerateAssetExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export")
	}

	f, err := excelize.OpenReader(bytes.NewReader(exported))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Daftar Aset")
	if err != nil {
		t.Fatalf("GetRows: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 accounts), got %d", len(rows))
	}

	wantHeaders := []string{"Kode", "Nama", "Kelompok", "Saldo"}
	for i, want := range wantHeaders {
		if rows[0][i] != want {
			t.Errorf("header col %d: expected %q, got %q", i, want, rows[0][i])
		}
	}

	if rows[1][0] != "1.01.01.01" {
		t.Errorf("expected code %q, got %q", "1.01.01.01", rows[1][0])
	}
	if rows[1][2] != "Bank" {
		t.Errorf("expected group %q, got %q", "Bank", rows[1][2])
	}
	if rows[1][3] != "1500000" {
		t.Errorf("expected balance %q, got %q", "1500000", rows[1][3])
	}
}

func TestGenerateAssetExport_Empty(t *testing.T) {
	exported, err := account.GenerateAssetExport(nil)
	if err != nil {
		t.Fatalf("GenerateAssetExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export even without groups")
	}
}
