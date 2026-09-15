package journal

import (
	"testing"
	"time"
)

func TestIsDebit(t *testing.T) {
	tests := []struct {
		name     string
		line     JournalEntryLine
		expected bool
	}{
		{
			name:     "debit line",
			line:     JournalEntryLine{Debit: 100, Credit: 0},
			expected: true,
		},
		{
			name:     "credit line",
			line:     JournalEntryLine{Debit: 0, Credit: 100},
			expected: false,
		},
		{
			name:     "zero amounts",
			line:     JournalEntryLine{Debit: 0, Credit: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.IsDebit(); got != tt.expected {
				t.Errorf("IsDebit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsCredit(t *testing.T) {
	tests := []struct {
		name     string
		line     JournalEntryLine
		expected bool
	}{
		{
			name:     "credit line",
			line:     JournalEntryLine{Debit: 0, Credit: 100},
			expected: true,
		},
		{
			name:     "debit line",
			line:     JournalEntryLine{Debit: 100, Credit: 0},
			expected: false,
		},
		{
			name:     "zero amounts",
			line:     JournalEntryLine{Debit: 0, Credit: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.line.IsCredit(); got != tt.expected {
				t.Errorf("IsCredit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTotalDebit(t *testing.T) {
	tests := []struct {
		name     string
		lines    []JournalEntryLine
		expected float64
	}{
		{
			name:     "empty lines",
			lines:    []JournalEntryLine{},
			expected: 0,
		},
		{
			name: "single debit line",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
			},
			expected: 100,
		},
		{
			name: "multiple lines",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
				{Debit: 0, Credit: 50},
				{Debit: 200, Credit: 0},
			},
			expected: 300,
		},
		{
			name: "no debit lines",
			lines: []JournalEntryLine{
				{Debit: 0, Credit: 100},
				{Debit: 0, Credit: 200},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalDebit(tt.lines); got != tt.expected {
				t.Errorf("TotalDebit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTotalCredit(t *testing.T) {
	tests := []struct {
		name     string
		lines    []JournalEntryLine
		expected float64
	}{
		{
			name:     "empty lines",
			lines:    []JournalEntryLine{},
			expected: 0,
		},
		{
			name: "single credit line",
			lines: []JournalEntryLine{
				{Debit: 0, Credit: 100},
			},
			expected: 100,
		},
		{
			name: "multiple lines",
			lines: []JournalEntryLine{
				{Debit: 0, Credit: 100},
				{Debit: 50, Credit: 0},
				{Debit: 0, Credit: 200},
			},
			expected: 300,
		},
		{
			name: "no credit lines",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
				{Debit: 200, Credit: 0},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalCredit(tt.lines); got != tt.expected {
				t.Errorf("TotalCredit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsBalanced(t *testing.T) {
	tests := []struct {
		name     string
		lines    []JournalEntryLine
		expected bool
	}{
		{
			name:     "empty lines",
			lines:    []JournalEntryLine{},
			expected: true,
		},
		{
			name: "balanced entry",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
				{Debit: 0, Credit: 100},
			},
			expected: true,
		},
		{
			name: "unbalanced entry - more debit",
			lines: []JournalEntryLine{
				{Debit: 200, Credit: 0},
				{Debit: 0, Credit: 100},
			},
			expected: false,
		},
		{
			name: "unbalanced entry - more credit",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
				{Debit: 0, Credit: 200},
			},
			expected: false,
		},
		{
			name: "balanced multi-line entry",
			lines: []JournalEntryLine{
				{Debit: 100, Credit: 0},
				{Debit: 200, Credit: 0},
				{Debit: 0, Credit: 150},
				{Debit: 0, Credit: 150},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBalanced(tt.lines); got != tt.expected {
				t.Errorf("IsBalanced() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestJournalEntryLine_methods(t *testing.T) {
	// Test IsDebit and IsCredit on same line
	debitLine := JournalEntryLine{ID: 1, AccountID: 1, Debit: 100, Credit: 0}
	if !debitLine.IsDebit() {
		t.Error("expected IsDebit() to be true for debit line")
	}
	if debitLine.IsCredit() {
		t.Error("expected IsCredit() to be false for debit line")
	}

	creditLine := JournalEntryLine{ID: 2, AccountID: 2, Debit: 0, Credit: 100}
	if creditLine.IsDebit() {
		t.Error("expected IsDebit() to be false for credit line")
	}
	if !creditLine.IsCredit() {
		t.Error("expected IsCredit() to be true for credit line")
	}
}

func TestJournalEntry_timestamps(t *testing.T) {
	now := time.Now()
	entry := JournalEntry{
		ID:          1,
		UserID:      100,
		Date:        now,
		Description: "Test entry",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if entry.ID != 1 {
		t.Errorf("expected ID 1, got %d", entry.ID)
	}
	if entry.UserID != 100 {
		t.Errorf("expected UserID 100, got %d", entry.UserID)
	}
	if entry.Description != "Test entry" {
		t.Errorf("expected Description 'Test entry', got %s", entry.Description)
	}
}
