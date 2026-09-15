package journal_test

import (
	"bytes"
	"testing"
	"time"

	"pengbook/api/internal/module/journal"
)

func TestGenerateExport_RoundTrip(t *testing.T) {
	date := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	rows := []journal.ExportRow{
		{Datetime: date, Description: "Pembelian perlengkapan", AccountCode: "5.01.01.04", Debit: 500000, Credit: 0},
		{Datetime: date, Description: "Pembelian perlengkapan", AccountCode: "1.01.01.01", Debit: 0, Credit: 500000},
		{Datetime: date.AddDate(0, 0, 1), Description: "Penjualan", AccountCode: "1.01.01.01", Debit: 200000, Credit: 0},
		{Datetime: date.AddDate(0, 0, 1), Description: "Penjualan", AccountCode: "4.01.01.01", Debit: 0, Credit: 200000},
	}
	accounts := []journal.AccountInfo{
		{Code: "5.01.01.04", Name: "Beban Perlengkapan", Type: "EXPENSE", Level: 3},
		{Code: "1.01.01.01", Name: "Kas", Type: "ASSET", Level: 3},
	}

	exported, err := journal.GenerateExport(rows, accounts)
	if err != nil {
		t.Fatalf("GenerateExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export")
	}

	// The export must be directly re-uploadable.
	parsed, err := journal.ParseExcel(bytes.NewReader(exported))
	if err != nil {
		t.Fatalf("ParseExcel(exported): %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 entries after round-trip, got %d", len(parsed))
	}
	for _, e := range parsed {
		if len(e.Lines) != 2 {
			t.Fatalf("expected 2 lines per entry, got %d", len(e.Lines))
		}
	}

	first := parsed[0]
	if first.Description != "Pembelian perlengkapan" {
		t.Errorf("expected first description kept, got %q", first.Description)
	}
	if first.Lines[0].AccountCode != "5.01.01.04" || first.Lines[0].Debit != 500000 {
		t.Errorf("unexpected first line: %+v", first.Lines[0])
	}
	if first.Lines[1].AccountCode != "1.01.01.01" || first.Lines[1].Credit != 500000 {
		t.Errorf("unexpected second line: %+v", first.Lines[1])
	}
}

func TestGenerateExport_Empty(t *testing.T) {
	exported, err := journal.GenerateExport(nil, nil)
	if err != nil {
		t.Fatalf("GenerateExport: %v", err)
	}
	if len(exported) == 0 {
		t.Fatal("expected non-empty export even without rows")
	}
}
