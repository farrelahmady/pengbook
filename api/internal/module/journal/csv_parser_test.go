package journal_test

import (
	"strings"
	"testing"
	"time"

	"pengbook/api/internal/module/journal"
)

func TestParseCSV_Success(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Pembelian perlengkapan,5.01.01.04,500000,0
2026-04-21,Pembelian perlengkapan,1.01.02.01,0,500000`

	entries, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
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

func TestParseCSV_MultipleEntries(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Entry 1,5.01.01.04,100000,0
2026-04-21,Entry 1,1.01.02.01,0,100000
2026-04-22,Entry 2,5.01.01.04,200000,0
2026-04-22,Entry 2,1.01.02.01,0,200000`

	entries, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
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

func TestParseCSV_WithCommasInAmount(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,5.01.01.04,"1,000,000",0
2026-04-21,Test,1.01.02.01,0,"1,000,000"`

	entries, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries[0].Lines[0].Debit != 1000000 {
		t.Errorf("expected debit 1000000, got %f", entries[0].Lines[0].Debit)
	}
}

func TestParseCSV_EmptyFile(t *testing.T) {
	csv := ``

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseCSV_HeaderOnly(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for header-only file")
	}
}

func TestParseCSV_MissingDate(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
,Test,5.01.01.04,100000,0
2026-04-21,Test,1.01.02.01,0,100000`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for missing date")
	}
}

func TestParseCSV_MissingAccountCode(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,,100000,0
2026-04-21,Test,1.01.02.01,0,100000`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for missing account code")
	}
}

func TestParseCSV_InvalidDate(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
not-a-date,Test,5.01.01.04,100000,0
2026-04-21,Test,1.01.02.01,0,100000`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestParseCSV_InvalidAmount(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,5.01.01.04,abc,0
2026-04-21,Test,1.01.02.01,0,100000`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for invalid amount")
	}
}

func TestParseCSV_TooFewLines(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,5.01.01.04,100000,0`

	_, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err == nil {
		t.Fatal("expected error for entry with only 1 line")
	}
}

func TestParseCSV_WithBOM(t *testing.T) {
	// BOM: 0xEF 0xBB 0xBF
	csv := "\xEF\xBB\xBFTanggal,Deskripsi,Kode Akun,Debit,Kredit\n2026-04-21,Test,5.01.01.04,100000,0\n2026-04-21,Test,1.01.02.01,0,100000"

	entries, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestParseCSV_SkipEmptyRows(t *testing.T) {
	csv := `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Test,5.01.01.04,100000,0
2026-04-21,Test,1.01.02.01,0,100000
2026-04-21,,5.01.01.04,0,0`

	entries, err := journal.ParseCSV(strings.NewReader(csv), time.UTC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 entry (empty row skipped)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}
