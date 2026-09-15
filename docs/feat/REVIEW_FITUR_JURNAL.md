# Review Fitur Create & Edit Jurnal

**Tanggal**: 12 September 2026  
**Status**: Tahap 1 Selesai  
**Dokumen Terkait**: [Architecture Issues](../issues/ARCHITECTURE_ISSUES_2026-09-11.md)

---

## Changelog

### Tahap 1 (12 September 2026)

**Frontend:**
- ✅ Ganti `POSTING_ACCOUNTS` hardcoded → fetch dari API `GET /api/v1/accounts/posting`
- ✅ Update 4 form components (basic, advanced, edit-basic, edit-advanced)
- ✅ Tambah `PostingAccount` type di `types/index.ts`
- ✅ Tambah `accounts.posting` query key
- ✅ Tambah `getPostingAccounts()` di `account.ts` service
- ✅ Tambah `createBulk()` di `journal.ts` service
- ✅ Update `upload-form.tsx` untuk parse CSV dan kirim JSON

**Backend:**
- ✅ Tambah `CreateBulkJournalRequest` DTO
- ✅ Tambah `POST /api/v1/journals/upload` endpoint
- ✅ Tambah `CreateBulk` service method
- ✅ Tambah `CreateEntries` repository method (batch insert)
- ✅ Wire `RecalculateAccountBalance` ke Create, Update, dan CreateBulk
- ✅ Fix pre-existing bug: `SumDebitByUserID` dan `SumCreditByUserID` menggunakan query yang benar

---

## Ringkasan

Review fitur Create dan Edit Jurnal mencakup analisis API (Go), Frontend (Next.js), Database, dan identifikasi gap antara implementasi saat ini dengan kebutuhan fitur.

---

## Yang Sudah Ada

| Komponen | Status | Lokasi |
|----------|--------|--------|
| API Create | ✅ Lengkap | `api/internal/module/journal/handler.go` |
| API Update | ✅ Lengkap | `api/internal/module/journal/handler.go` |
| API List (Scroll View) | ✅ Cursor-based | `api/internal/module/journal/handler.go` |
| API Summary | ✅ Aggregate | `api/internal/module/journal/handler.go` |
| Database Schema | ✅ Lengkap | `api/migrations/20260909120004_create_journal_entries.sql` |
| Frontend Basic Form | ✅ Form sederhana | `web/components/journal/basic-form.tsx` |
| Frontend Advanced Form | ✅ Multi-line | `web/components/journal/advanced-form.tsx` |
| Frontend Edit Form | ✅ Basic + Advanced | `web/components/journal/edit-journal-sheet.tsx` |
| Infinite Scroll List | ✅ Cursor-based | `web/components/journal/journal-scroll-view.tsx` |

---

## Critical Issues

### 1. Hardcoded Account Data di Frontend
- **Masalah**: `POSTING_ACCOUNTS` di `web/lib/constants.ts` hardcode 8 akun
- **Dampak**: Akun baru dari user tidak muncul, ID akun mungkin tidak match
- **File**: `web/lib/constants.ts`
- **Fix**: Fetch dari `GET /api/v1/accounts/posting` via service

### 2. `createBulk()` Belum Diimplementasi
- **Masalah**: `UploadForm` panggil `journalService.createBulk(file)` tapi method tidak ada di `web/services/journal.ts`
- **Dampak**: Fitur upload CSV/Excel akan error di runtime
- **File**: `web/components/journal/upload-form.tsx`, `web/services/journal.ts`
- **Fix**: Buat endpoint `POST /api/v1/journals/bulk` di backend + implement method di service

### 3. Account Balance Tidak Terupdate
- **Masalah**: Saat create/update jurnal, `account_balances` TIDAK diupdate
- **Dampak**: Saldo akun akan selalu stale/salah
- **File**: `api/internal/module/journal/service.go`
- **Fix**: Panggil `recalculate_account_balance()` atau update `account_balances` di transaction

---

## High Priority Issues

### 4. Tidak Ada DELETE Endpoint
- **Masalah**: Repository punya `DeleteEntry()` tapi tidak di-expose ke API
- **Dampak**: User tidak bisa hapus jurnal
- **File**: `api/internal/module/journal/handler.go`
- **Fix**: Tambah `DELETE /api/v1/journals/{id}` + service method `Delete()`

### 5. Tidak Ada Audit Logging
- **Masalah**: Module Account ada audit log, tapi Journal tidak ada
- **Dampak**: Tidak ada jejak siapa create/edit/delete
- **File**: `api/internal/module/journal/service.go`
- **Fix**: Tambah `InsertAuditLog` seperti module account

### 6. COA Filter Dikomentari di Frontend
- **Masalah**: `journal-list.tsx` punya UI COA filter tapi dikomentari
- **Dampak**: User tidak bisa filter by akun
- **File**: `web/app/[locale]/(auth)/jurnal/journal-list.tsx`
- **Fix**: Aktifkan kembali + implement COA picker component

### 7. Timezone Issue di Basic Form
- **Masalah**: `new Date(date).toISOString()` bisa shift date ±1 hari tergantung timezone
- **Dampak**: Tanggal bisa meleset 1 hari
- **File**: `web/components/journal/basic-form.tsx`
- **Fix**: Handle timezone dengan benar, gunakan library date-fns atau handle manual

---

## Medium Priority Issues

| Issue | Deskripsi | File |
|-------|-----------|------|
| Update Replace All Lines | Line IDs berubah setiap update, bisa break referensi | `api/internal/infrastructure/postgres/journal_repository.go` |
| No Single GET Endpoint | Tidak ada `GET /api/v1/journals/{id}` untuk fetch satu entry | `api/internal/module/journal/handler.go` |
| Download Feature | Tombol download ada tapi implementation dikomentari | `web/app/[locale]/(auth)/jurnal/journal-topbar.tsx` |
| Transaction Module Empty | Module `transaction/` hanya stub kosong | `api/internal/module/transaction/` |
| Floating-point Comparison | `Math.abs(totalDr - totalCr) < 0.1` bisa miss edge case | `web/components/journal/advanced-form.tsx` |

---

## Rekomendasi Implementasi

### Tahap 1: Perbaiki Critical Issues
1. **[Frontend]** Ganti `POSTING_ACCOUNTS` hardcoded → fetch dari API
2. **[Backend]** Buat endpoint `POST /api/v1/journals/bulk` untuk upload
3. **[Backend]** Wire account balance recalculation ke journal service

### Tahap 2: Tambah Fitur yang Kurang
4. **[Backend]** Tambah `DELETE /api/v1/journals/{id}`
5. **[Backend]** Tambah audit logging untuk create/update/delete
6. **[Frontend]** Implementasi delete button di UI
7. **[Frontend]** Aktifkan COA filter yang sudah dikomentari

### Tahap 3: Improve Quality
8. **[Backend]** Tambah `GET /api/v1/journals/{id}` untuk single entry
9. **[Frontend]** Fix timezone handling di date input
10. **[Backend]** Pertimbangkan batch insert untuk lines (performance)
11. **[Frontend]** Implementasi download feature

---

## File yang Perlu Diubah

### Tahap 1 (Selesai ✅)

| File | Perubahan |
|------|-----------|
| `web/lib/constants.ts` | ✅ Hapus `POSTING_ACCOUNTS` hardcoded |
| `web/types/index.ts` | ✅ Tambah `PostingAccount` type |
| `web/lib/query-keys.ts` | ✅ Tambah `accounts.posting` key |
| `web/services/account.ts` | ✅ Tambah `getPostingAccounts()` |
| `web/services/journal.ts` | ✅ Tambah `createBulk()` |
| `web/components/journal/basic-form.tsx` | ✅ Fetch accounts dari API |
| `web/components/journal/advanced-form.tsx` | ✅ Fetch accounts dari API |
| `web/components/journal/edit-basic-form.tsx` | ✅ Fetch accounts dari API |
| `web/components/journal/edit-advanced-form.tsx` | ✅ Fetch accounts dari API |
| `web/components/journal/upload-form.tsx` | ✅ Parse CSV → JSON |
| `api/internal/module/journal/dto.go` | ✅ Tambah `CreateBulkJournalRequest` |
| `api/internal/module/journal/handler.go` | ✅ Tambah `POST /upload` endpoint |
| `api/internal/module/journal/service.go` | ✅ Tambah `CreateBulk` + balance recalculation |
| `api/internal/module/journal/repository.go` | ✅ Tambah `CreateEntries` + `RecalculateAccountBalance` |
| `api/internal/infrastructure/postgres/journal_repository.go` | ✅ Implement methods + fix bugs |

### Tahap 2 (Belum Dikerjakan)

| File | Perubahan |
|------|-----------|
| `api/internal/module/journal/handler.go` | Tambah `DELETE /{id}` endpoint |
| `api/internal/module/journal/service.go` | Tambah `Delete()`, `GetByID()` |
| `web/services/journal.ts` | Tambah `delete()`, `getById()` |
| `web/components/journal/` | Implement delete button di UI |
| `web/app/[locale]/(auth)/jurnal/journal-list.tsx` | Aktifkan COA filter |

---

## Pertanyaan untuk Diputuskan

1. **Prioritas mana yang mau dikerjakan dulu?** Critical issues atau langsung lengkap?
2. **Fitur upload CSV/Excel perlu diimplement sekarang?** Atau skip dulu?
3. **Apakah perlu delete feature atau cukup create/edit saja untuk saat ini?**

---

## API Endpoints (Saat Ini)

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `GET` | `/api/v1/journals` | `GetAllScrollView` | ✅ Berfungsi |
| `GET` | `/api/v1/journals/summary` | `GetTotalSummary` | ✅ Berfungsi |
| `POST` | `/api/v1/journals` | `Create` | ✅ Berfungsi |
| `POST` | `/api/v1/journals/upload` | `Upload` | ✅ Baru ditambahkan |
| `PUT` | `/api/v1/journals/{id}` | `Update` | ✅ Berfungsi |
| `DELETE` | `/api/v1/journals/{id}` | - | ❌ Belum ada |
| `GET` | `/api/v1/journals/{id}` | - | ❌ Belum ada |

---

## Database Tables

### journal_entries
```sql
CREATE TABLE journal_entries (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    datetime    TIMESTAMPTZ NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### journal_entry_lines
```sql
CREATE TABLE journal_entry_lines (
    id                BIGSERIAL PRIMARY KEY,
    journal_entry_id  BIGINT NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id        BIGINT NOT NULL REFERENCES accounts(id),
    debit             NUMERIC(19,4) NOT NULL DEFAULT 0,
    credit            NUMERIC(19,4) NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_debit_credit CHECK (
        (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)
    )
);
```

---

## Service Interface

```go
type Service interface {
    GetAllScrollView(ctx context.Context, userID int64, filter ListRequest) (*CursorPageResponse, error)
    GetTotalSummary(ctx context.Context, userID int64) (*JournalSummary, error)
    Create(ctx context.Context, userID int64, req CreateJournalRequest) (*JournalEntryResponse, error)
    Update(ctx context.Context, userID int64, entryID int64, req UpdateJournalRequest) (*JournalEntryResponse, error)
    CreateBulk(ctx context.Context, userID int64, req CreateBulkJournalRequest) (int64, error)
}
```

**Yang sudah ditambahkan:**
- ✅ `CreateBulk(ctx, userID, entries) (int64, error)` - Bulk upload entries

**Yang perlu ditambahkan di Tahap 2:**
- `Delete(ctx, userID, entryID) error`
- `GetByID(ctx, userID, entryID) (*JournalEntry, error)`
