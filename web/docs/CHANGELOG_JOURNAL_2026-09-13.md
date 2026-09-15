# Changelog — Journal Module

Dokumentasi lengkap perubahan pada modul Jurnal (Backend + Frontend).

---

## Table of Contents

- [Ringkasan](#ringkasan)
- [Backend Changes](#backend-changes)
- [Frontend Changes](#frontend-changes)
- [Testing](#testing)
- [Bug Fixes](#bug-fixes)
- [Files Reference](#files-reference)

---

## Ringkasan

Modul Jurnal mengalami perubahan besar dari fitur dasar menjadi fitur lengkap dengan:

| Fitur | Status |
|-------|--------|
| Create & Edit Journal | ✅ |
| Advanced Mode (multi-line) | ✅ |
| Upload CSV (Backend Parse) | ✅ |
| Upload Excel (Backend Parse) | ✅ |
| Download Template Excel | ✅ |
| Dynamic Template (user's accounts) | ✅ |
| Balance Recalculation | ✅ |
| Comprehensive Testing | ✅ |
| i18n Support | ✅ |

---

## Backend Changes

### 1. Domain Layer (`domain.go`)

**Added:**
- `IsDebit()`, `IsCredit()` — Check line type
- `TotalDebit()`, `TotalCredit()` — Calculate totals
- `IsBalanced()` — Check if entry balances

**Unit Tests (`domain_test.go`):**
- 7 tests for domain logic

### 2. Repository Layer

**`repository.go` — Interface Added:**
```go
FindByCodes(ctx context.Context, userID int64, codes []string) ([]domain.Account, error)
SumDebitByUserID(ctx context.Context, userID int64) (float64, error)
SumCreditByUserID(ctx context.Context, userID int64) (float64, error)
RecalculateAccountBalance(ctx context.Context, accountID int64) error
CreateEntries(ctx context.Context, userID int64, entries []domain.JournalEntry) error
```

**`postgres/journal_repository.go` — Implementation:**
- `FindByCodes` — Find accounts by codes for validation
- `CreateEntries` — Batch insert for upload
- `RecalculateAccountBalance` — Call PostgreSQL function

**`postgres/account_repository.go` — Added:**
- `FindByCodes` — Find accounts by codes

### 3. Service Layer (`service.go`)

**Added:**
```go
CreateBulk(ctx context.Context, userID int64, req dto.CreateBulkJournalRequest) (*BulkResult, error)
GenerateTemplate(ctx context.Context, userID int64) ([]byte, error)
```

**CreateBulk Logic:**
1. Validate account codes exist
2. Check all accounts are posting accounts (level 3)
3. Check debit = credit (balanced)
4. Batch insert entries
5. Recalculate affected account balances

**GenerateTemplate Logic:**
1. Fetch user's posting accounts
2. Create Excel with 2 sheets:
   - Sheet 1: Template Jurnal (with example data)
   - Sheet 2: Daftar Akun (user's accounts)

### 4. Handler Layer (`handler.go`)

**Endpoints Added:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/journals/upload` | Upload CSV/Excel file |
| `GET` | `/api/v1/journals/template` | Download Excel template |

**Upload Handler:**
- Accept multipart form (CSV or Excel)
- Detect file type by extension
- Parse file (CSV or Excel)
- Validate & save to database
- Return count of created entries

**Download Template Handler:**
- Generate Excel with user's posting accounts
- Return as binary Excel file
- Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`

### 5. CSV Parser (`csv_parser.go`)

**Function:**
```go
func ParseCSV(reader io.Reader) ([]CreateJournalRequest, error)
```

**Features:**
- Position-based parsing (column indices 0-4)
- Auto-detect header row (skip if contains "tanggal" or "date")
- Support BOM (Byte Order Mark)
- Group lines by date + description
- Multiple date formats: YYYY-MM-DD, DD/MM/YYYY, MM/DD/YYYY, RFC3339
- Number formatting: handles commas (1,000,000)

**CSV Format:**
```csv
Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Pembelian perlengkapan,5.01.01.04,500000,0
2026-04-21,Pembelian perlengkapan,1.01.02.01,0,500000
```

### 6. Excel Parser (`excel_parser.go`)

**Functions:**
```go
func ParseExcel(reader io.Reader) ([]CreateJournalRequest, error)
func GenerateTemplate() ([]byte, error)
func GenerateTemplateWithAccounts(accounts []domain.Account) ([]byte, error)
```

**Features:**
- Parse Excel (.xlsx) files
- Auto-detect header row
- Group lines by date + description
- Generate Excel with 2 sheets
- Use `excelize/v2` library

**Template Structure:**
```
Sheet 1: "Template Jurnal"
├── Tanggal | Deskripsi | Kode Akun | Debit | Kredit
├── 2026-04-21 | Contoh | 1.01.01.01 | 500000 | 0
└── 2026-04-21 | Contoh | 4.01.01.01 | 0 | 500000

Sheet 2: "Daftar Akun"
├── Kode Akun | Nama Akun | Tipe | Level | Posting
└── (user's posting accounts)
```

### 7. Middleware (`middleware/auth.go`)

**Added:**
```go
func ContextWithUserID(ctx context.Context, userID int64) context.Context
```
- Export for testing purposes
- Used in handler tests to mock authenticated user

---

## Frontend Changes

### 1. Upload Form (`components/journal/upload-form.tsx`)

**Features:**
- Accept CSV and Excel files (.csv, .xlsx, .xls)
- Drag & drop support
- File validation (type & size)
- Download template from API
- Two options section (Excel + CSV)
- Format rules display
- i18n support

**UI Sections:**
```
┌─────────────────────────────────────────────────────────────────┐
│  Two Options                                                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │ Excel (Recommended) │  │ CSV Manual          │              │
│  │ [Download Template] │  │ Format: CSV...      │              │
│  └─────────────────────┘  └─────────────────────┘              │
├─────────────────────────────────────────────────────────────────┤
│  Format Rules:                                                  │
│  • Date format: YYYY-MM-DD                                      │
│  • Account code: 4 segments                                     │
│  • Minimum 2 lines per entry                                    │
│  • Debit must equal Credit                                      │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Drag & drop file here                                  │    │
│  │  or Choose File                                         │    │
│  │  Format: CSV, XLSX (Max 5MB)                            │    │
│  └─────────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────────┤
│  [Upload & Process]                                             │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Journal Service (`services/journal.ts`)

**Added:**
```typescript
uploadFile: async (file: File): Promise<{ count: number }>
downloadTemplate: async (): Promise<Blob>
```

**Features:**
- Upload file via multipart form
- Download template as Blob
- Logger integration

### 3. i18n (`messages/en.json`, `messages/id.json`)

**Added Keys:**

| Key | English | Indonesian |
|-----|---------|------------|
| `optionExcelTitle` | Excel (Recommended) | Excel (Disarankan) |
| `optionExcelDesc` | Download template, fill in data... | Download template, isi data... |
| `optionExcelButton` | Download Template | Download Template |
| `optionCSVTitle` | CSV Manual | CSV Manual |
| `optionCSVDesc` | Create your own CSV file... | Buat file CSV sendiri... |
| `csvFormatTitle` | CSV Format: | Format CSV: |
| `csvFormatExample1` | Date,Description,Account Code,Debit,Credit | Tanggal,Deskripsi,Kode Akun,Debit,Kredit |
| `csvFormatExample2` | 2026-04-21,Example,1.01.01.01,500000,0 | 2026-04-21,Contoh,1.01.01.01,500000,0 |
| `csvFormatExample3` | 2026-04-21,Example,4.01.01.01,0,500000 | 2026-04-21,Contoh,4.01.01.01,0,500000 |
| `rulesTitle` | Format Rules: | Ketentuan Format: |
| `ruleDateFormat` | Date format: YYYY-MM-DD | Format tanggal: YYYY-MM-DD |
| `ruleAccountCode` | Account code: 4 segments... | Kode akun: 4 segmen... |
| `ruleMinLines` | Minimum 2 lines per journal entry | Minimal 2 baris per entri jurnal |
| `ruleBalanced` | Debit must equal Credit | Debit harus sama dengan Kredit |

---

## Testing

### Test Summary

| Layer | File | Tests | Status |
|-------|------|-------|--------|
| Domain | `domain_test.go` | 7 | ✅ |
| Repository | `journal_integration_test.go` | 7 | ✅ |
| Service | `journal_integration_test.go` | 12 | ✅ |
| CSV Parser | `csv_parser_test.go` | 12 | ✅ |
| Excel Parser | `excel_parser_test.go` | 10 | ✅ |
| Handler | `handler_test.go` | 21 | ✅ |
| **Total** | | **69** | ✅ |

### Handler Tests Breakdown

| Endpoint | Tests |
|----------|-------|
| `GET /journals` (GetAllScrollView) | 2 (success, unauthorized) |
| `GET /journals/summary` (GetTotalSummary) | 1 |
| `POST /journals` (Create) | 4 (success, unbalanced, invalid body, invalid lines) |
| `PUT /journals/:id` (Update) | 2 (success, not found) |
| `POST /journals/upload` (Upload CSV) | 8 (success, multiple, unbalanced, not posting, no file, wrong ext, empty, invalid format) |
| `POST /journals/upload` (Upload Excel) | 3 (success, wrong ext, empty) |
| `GET /journals/template` (DownloadTemplate) | 3 (success, unauthorized, no accounts) |

### Running Tests

```bash
# Run all journal tests
go test ./internal/module/journal/ -v

# Run specific test suite
go test ./internal/module/journal/ -v -run "TestParseCSV"
go test ./internal/module/journal/ -v -run "TestParseExcel"
go test ./internal/module/journal/ -v -run "TestHandler_Upload"
go test ./internal/module/journal/ -v -run "TestHandler_DownloadTemplate"
```

---

## Bug Fixes

### 1. Pre-existing Bugs Fixed

**`SumDebitByUserID` / `SumCreditByUserID` undefined:**
- Added missing function implementations in repository

**`CreateEntry` missing lines:**
- Fixed repository to properly insert journal entry lines

### 2. Test Infrastructure Fixes

**Duplicate username in tests:**
- Problem: Tests failed due to unique constraint on `username`
- Solution: Use `nanosecond + rand.Intn(10000)` for unique usernames

**Leftover test data:**
- Problem: Previous test runs left data in database
- Solution: Cleanup query before tests

**Context key access in handler tests:**
- Problem: `userIDKey` is unexported
- Solution: Added `ContextWithUserID()` to middleware

---

## Files Reference

### Backend Files

| File | Description |
|------|-------------|
| `api/internal/module/journal/domain.go` | Domain entities |
| `api/internal/module/journal/domain_test.go` | Domain unit tests |
| `api/internal/module/journal/dto.go` | DTOs |
| `api/internal/module/journal/repository.go` | Repository interface |
| `api/internal/module/journal/service.go` | Service interface + impl |
| `api/internal/module/journal/handler.go` | HTTP handlers |
| `api/internal/module/journal/csv_parser.go` | CSV parsing logic |
| `api/internal/module/journal/csv_parser_test.go` | CSV parser tests |
| `api/internal/module/journal/excel_parser.go` | Excel parsing + template generation |
| `api/internal/module/journal/excel_parser_test.go` | Excel parser tests |
| `api/internal/module/journal/journal_integration_test.go` | Repository + Service tests |
| `api/internal/module/journal/handler_test.go` | Handler tests |
| `api/internal/infrastructure/postgres/journal_repository.go` | PostgreSQL repository |
| `api/internal/infrastructure/postgres/account_repository.go` | Account repository (FindByCodes) |
| `api/internal/middleware/auth.go` | Auth middleware (ContextWithUserID) |
| `api/go.mod` | Added excelize/v2 |

### Frontend Files

| File | Description |
|------|-------------|
| `web/components/journal/upload-form.tsx` | Upload form component |
| `web/services/journal.ts` | Journal API service |
| `web/messages/en.json` | English translations |
| `web/messages/id.json` | Indonesian translations |
| `web/docs/UPLOAD_JURNAL.md` | Upload documentation |
| `web/docs/CHANGELOG_JOURNAL_2026-09-13.md` | This changelog |

---

## API Reference

### Upload File

```
POST /api/v1/journals/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

file: <CSV or Excel file>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "count": 5
  }
}
```

### Download Template

```
GET /api/v1/journals/template
Authorization: Bearer <token>
```

**Response:**
- Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- Body: Binary Excel file

---

## References

- [Upload Jurnal Documentation](UPLOAD_JURNAL.md)
- [Architecture Issues](issues/ARCHITECTURE_ISSUES_2026-09-11.md)
- [Logging Improvement](issues/LOGGING_IMPROVEMENT_2026-09-12.md)
