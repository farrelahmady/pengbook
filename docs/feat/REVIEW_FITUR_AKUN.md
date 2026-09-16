# Review Fitur Akun (Chart of Accounts)

**Tanggal**: 16 September 2026
**Status**: Tahap 2 sebagian (F1, F2, F3) Selesai — `GET /tree`, `postingCount`, create akun + codegen, rename Coa→Account
**Dokumen Terkait**: [Review Fitur Jurnal](./REVIEW_FITUR_JURNAL.md), [Architecture Issues](../issues/ARCHITECTURE_ISSUES_2026-09-11.md)

---

## Ringkasan

Review fitur halaman Akun mencakup analisis API (Go), Frontend (Next.js), Database, dan identifikasi gap antara implementasi saat ini dengan kebutuhan fitur. Saat ini halaman Akun adalah **viewer read-only**: Topbar ringkasan + daftar grup tipe yang bisa di-expand. Backend sudah menyediakan `POST`/`PUT`/`DELETE /accounts`, tetapi frontend tidak memakainya sama sekali. Selain itu, tree hierarki saat ini **menumpang di endpoint `summary`**, sehingga Topbar (yang hanya butuh 3 angka) ikut memuat seluruh tree ~130+ akun.

Scope yang disetujui: F1 (pemecahan API tree terpisah, tahap pertama), F2 (`postingCount` ke backend), F3 (create), F4 (edit tanpa delete), F5 (search), F6 (error state), F7 (expand/collapse + aria), F8 (i18n). **Delete eksplisit di luar scope.**

---

## Yang Sudah Ada

| Komponen | Status | Lokasi |
|----------|--------|--------|
| API Summary (counts saja) | ✅ Dilangsingkan 16 Sep 2026 | `api/internal/module/account/handler.go` (`GetSummary`) |
| API Tree (groups + tree) | ✅ Baru 16 Sep 2026 | `api/internal/module/account/handler.go` (`GetTree`) |
| `postingCount` per grup dari backend | ✅ Selesai 16 Sep 2026 | `api/.../account/service.go` (`groupByType`), `web/components/akun/account-type-group.tsx` |
| API Posting Accounts | ✅ Berfungsi | `api/internal/module/account/handler.go` (`GetPostingAccounts`) |
| API Create | ✅ Ada, validasi parent lemah | `api/internal/module/account/service.go` (`Create`) |
| API Update | ✅ Ada, validasi parent lemah | `api/internal/module/account/service.go` (`Update`) |
| API Delete | ✅ Ada (di luar scope UI) | `api/internal/module/account/service.go` (`Delete`) |
| buildTree pointer fix | ✅ Selesai 16 Sep 2026 | `api/internal/module/account/service.go`, `dto.go`, `service_test.go` |
| Database Schema + constraints | ✅ Lengkap | `api/migrations/20260909120003_create_accounts.sql` |
| Seeder ~130 akun 4 level | ✅ Ada | `api/seeders/20260909134631_seed_accounts.sql` |
| Frontend Topbar + List + GroupCard | ✅ Read-only viewer | `web/app/[locale]/(auth)/akun/page.tsx`, `web/components/akun/` |
| Frontend Skeleton + EmptyState | ✅ Ada (tanpa error state) | `web/components/akun/account-skeleton.tsx` |
| accountService.getSummary/getPostingAccounts | ✅ Berfungsi | `web/services/account.ts` |
| Query key `accounts.summary` / `accounts.posting` | ✅ Ada | `web/lib/query-keys.ts` |
| i18n dasar `coaPage` (id + en) | ✅ Ada, ada string bocor | `web/messages/id.json`, `web/messages/en.json` |

---

## Critical Issues

### F1. Tree Menumpang di Endpoint Summary
- **Masalah**: `GET /api/v1/accounts/summary` mengembalikan counts + seluruh tree (`buildTree` + `groupByType`). `AccountTopbar` (cuma butuh 3 angka) dan `AccountList` (butuh tree) memakai queryKey yang sama sehingga Topbar ikut menunggu/mengunduh tree.
- **Dampak**: Payload besar untuk kebutuhan kecil, refetch Topbar = refetch tree, susah menambah field tree (mis. `postingCount` per grup) tanpa membebani Topbar.
- **File**: `api/internal/module/account/handler.go`, `service.go`, `dto.go`, `web/services/account.ts`, `web/lib/query-keys.ts`
- **Fix**: Pecah menjadi dua endpoint — `GET /summary` (ringan: hanya angka) + `GET /tree` (baru: groups + tree). Detail di Plan F1.

### F2. `postingCount` Dihitung di Frontend
- **Masalah**: `countPostingAccounts()` di `web/components/akun/account-type-group.tsx:35` berekursi di client untuk subtitle "Y posting".
- **Dampak**: Melanggar prinsip "kalkulasi di API" (walau ringan); duplikasi logika count yang sudah ada di backend (`CountPostingByUserID`, `group.Count`).
- **File**: `web/components/akun/account-type-group.tsx`
- **Fix**: Backend mengirim `postingCount` per grup di respons `tree`; frontend hanya render. Detail di Plan F2.

---

## High Priority Issues

### F3. Tidak Ada UI Create Akun
- **Masalah**: `POST /api/v1/accounts` ada, tapi tidak ada FAB/form di halaman Akun. `emptyDescription` ("Tambahkan akun...") menjanjikan aksi tanpa tombol.
- **Dampak**: User tidak bisa tambah akun dari UI.
- **File**: `web/app/[locale]/(auth)/akun/page.tsx`, `web/services/account.ts` (belum ada `create()`)
- **Fix**: FAB + Bottom Sheet meniru pola `CreateJournal`/`CreateJournalSheet` (`SheetContent side="bottom"`, `max-w-[390px]`). Detail di Plan F3.

### F4. Tidak Ada UI Edit Akun (Delete di Luar Scope)
- **Masalah**: `PUT /api/v1/accounts/{id}` ada, tapi tidak ada UI rename/pindah parent.
- **Dampak**: Salah ketik nama akun permanen dari sisi UI.
- **File**: `web/components/akun/account-type-group.tsx` (aksi per baris), `web/services/account.ts` (belum ada `update()`)
- **Fix**: Aksi edit per baris akun → sheet edit. Delete **sengaja tidak dikerjakan** sesuai keputusan. Detail di Plan F4.

### F5. Tidak Ada Search
- **Masalah**: ~130 akun dalam 5 grup collapsed; tanpa search, mencari satu akun = expand-scroll manual.
- **Dampak**: Navigasi akun lambat dan menyebalkan.
- **File**: `web/components/akun/account-list.tsx`
- **Fix**: Input search filter code/nama + auto-expand grup yang cocok (filter tampilan, boleh di frontend). Detail di Plan F5.

### F6. Tidak Ada Error State
- **Masalah**: `account-list.tsx` hanya menangani `isLoading` vs data. Saat `getSummary()`/`getTree()` gagal (401/500/offline), halaman tampil kosong tanpa pesan.
- **Dampak**: User mengira data kosong; tidak ada retry.
- **File**: `web/components/akun/account-list.tsx`, `web/components/akun/account-topbar.tsx`
- **Fix**: Error card + tombol retry (`refetch`), meniru pola empty/error halaman Jurnal. Detail di Plan F6.

---

## Medium Priority Issues

| Issue | Deskripsi | File |
|-------|-----------|------|
| F7. Expand/collapse all + aria | Tiap card collapsed-default tanpa kontrol massal; tombol expand tanpa `aria-expanded` | `web/components/akun/account-type-group.tsx`, `web/components/akun/account-list.tsx` |
| F8. String i18n bocor | `"Posting"`, `"Header"`, `"X akun · Y posting"` hardcode Indonesia; locale EN tetap tampil Indonesia | `web/components/akun/account-type-group.tsx`, `web/messages/id.json`, `web/messages/en.json` |
| Validasi parent lemah (prasyarat) | `Service.Create/Update` tidak verifikasi parent milik user yang sama, `level == parent.level+1`, prefix code konsisten dengan parent | `api/internal/module/account/service.go` |
| Delete guard mati (catatan, di luar scope) | Handler memetakan `ErrHasChildren`/`ErrHasJournalLines` → 409, tapi `service.Delete` tidak pernah mengeceknya | `api/internal/module/account/service.go`, `handler.go` |

---

## Rekomendasi Implementasi

### Tahap 1: Pecah API + Pindahkan Count (F1, F2)

**F1 — Endpoint `tree` terpisah dari `summary`:**
1. **[Backend]** Tambah DTO `AccountTree` (reuse `AccountTypeGroup` + field baru) dan method `GetTree(ctx, userID) (*AccountTree, error)` di `service.go`: `FindByUserID` → `buildTree` → `groupByType`. `GetSummary` dilangsingkan: hanya 3 count (tanpa `FindByUserID`/tree).
2. **[Backend]** Tambah handler `GetTree` + route `r.Get("/tree", h.GetTree)` di `handler.go`. Migrasi aman: **tambah `/tree` dulu, pindahkan frontend, baru langsingkan `/summary`** (kontrak `summary` yang lama memuat `groups`; frontend lama pecah jika dilangsingkan duluan).
3. **[Frontend]** Tambah `getTree()` di `services/account.ts` (via `authHttpClient` + `createLogger("account.service")`), tambah query key `accounts.tree` di `lib/query-keys.ts`. `AccountTopbar` tetap di `accounts.summary`, `AccountList` pindah ke `accounts.tree`. Invalidasi silang summary↔tree setelah create/update.

**F2 — `postingCount` dari backend:**
4. **[Backend]** Tambah field `postingCount` per grup (reuse pola `countType`, hitung node `IsPosting`). Frontend hapus `countPostingAccounts()` dan render `group.postingCount`.

### Tahap 2: CRUD Tanpa Delete (F3, F4)

**F3 — Create:**
5. **[Frontend]** `CreateAccountSheet` (Sheet bottom, pola `create-journal-sheet.tsx`) + FAB di `app/[locale]/(auth)/akun/` meniru `create-journal.tsx`. Form: `code` (panduan format `X.YY.ZZ.WW`, validasi regex sama dengan `chk_accounts_code_format`, unik per user → tangani 409) + `name` + parent picker (difilter kandidat valid: level tepat satu di bawah, tipe sama). `accountService.create()` + `useMutation` + invalidasi `accounts.summary` + `accounts.tree` + toast sukses/gagal (sonner) + key i18n baru + logging tiap operasi.
6. **[Backend — prasyarat]** Perketat `Service.Create`: parent wajib milik user yang sama, `level(child) == level(parent)+1`, prefix code konsisten dengan parent; kembalikan `ErrCodeExists` → 409 (sudah ada) + error validasi → 400 yang bisa ditampilkan form.

**F4 — Edit (tanpa delete):**
7. **[Frontend]** Aksi edit per baris akun (di `AccountTree`) → `EditAccountSheet`: rename + pindah parent (bukan pindah lintas tipe tanpa konfirmasi). `accountService.update()` + invalidasi + toast + i18n.
8. **[Backend — prasyarat]** Perketat `Service.Update` sama seperti Create (cegah siklus parent: parent baru tidak boleh diri sendiri/descendant).

### Tahap 3: Kualitas List (F5, F6, F7, F8)

9. **[Frontend — F5]** Search input di atas list: filter code/nama (case-insensitive), auto-expand grup berisi hasil, tampilkan count "N hasil", tombol clear. Kosong → EmptyState varian "tidak cocok".
10. **[Frontend — F6]** Error state: tangkap `isError` di `AccountList`/`AccountTopbar`, tampilkan error card + tombol "Coba lagi" (`refetch`); skeleton tetap untuk loading.
11. **[Frontend — F7]** Toolbar "Buka semua / Tutup semua" + state expanded terpusat (lift dari lokal per-card atau context ringan); tambah `aria-expanded` + `aria-controls` pada tombol grup.
12. **[Frontend — F8]** Pindahkan `"Posting"`, `"Header"`, `"X akun · Y posting"` ke `coaPage` di `id.json` + `en.json`; ganti semua string hardcode dengan `useTranslations`.

---

## File yang Perlu Diubah

### Tahap 1 (F1, F2 — API split + count)

| File | Perubahan |
|------|-----------|
| `api/internal/module/account/dto.go` | Tambah DTO tree (`AccountTree` + `postingCount` per grup) |
| `api/internal/module/account/service.go` | Tambah `GetTree()`, langsingkan `GetSummary()`, tambah `postingCount` di `groupByType` |
| `api/internal/module/account/handler.go` | Tambah `GetTree` + route `GET /tree` |
| `api/internal/module/account/service_test.go` | Test `GetTree`/`groupByType` (postingCount, urutan, grup kosong di-skip) |
| `web/types/index.ts` | Tambah tipe `AccountTree`/`AccountTypeGroup.postingCount` |
| `web/lib/query-keys.ts` | Tambah `accounts.tree` key |
| `web/services/account.ts` | Tambah `getTree()` (authHttpClient + logger) |
| `web/components/akun/account-list.tsx` | Pindah ke `accounts.tree` + hapus kalkulasi count |
| `web/components/akun/account-type-group.tsx` | Render `group.postingCount` dari API |
| `web/components/akun/account-topbar.tsx` | Tetap di `accounts.summary` (ringan) |

### Tahap 2 (F3, F4 — create + edit)

| File | Perubahan |
|------|-----------|
| `api/internal/module/account/service.go` | Perketat validasi parent di `Create`/`Update` (kepemilikan, level, prefix, anti-siklus) |
| `web/services/account.ts` | Tambah `create()` + `update()` + logging |
| `web/components/akun/create-account-sheet.tsx` | Baru: form code/nama/parent (Sheet bottom) |
| `web/components/akun/edit-account-sheet.tsx` | Baru: form edit per akun |
| `web/app/[locale]/(auth)/akun/page.tsx` | Tambah FAB + mount sheets (pola `create-journal.tsx`) |
| `web/components/akun/account-type-group.tsx` | Tambah aksi edit per baris |
| `web/messages/id.json`, `web/messages/en.json` | Key form, validasi, toast |

### Tahap 3 (F5–F8 — kualitas)

| File | Perubahan |
|------|-----------|
| `web/components/akun/account-list.tsx` | Search state, expand-all state, error card + retry |
| `web/components/akun/account-type-group.tsx` | Controlled expanded + `aria-expanded`, string via i18n |
| `web/messages/id.json`, `web/messages/en.json` | Key search, error, expand, badge |

---

## Pertanyaan untuk Diputuskan

1. **Code manual vs auto-suggest dari parent?** Apakah user mengetik `code` bebas (dengan validasi), atau pilih parent lalu code disarankan otomatis (mis. segmen berikutnya yang kosong)?
2. **Edit boleh pindah parent lintas tipe?** Atau hanya rename + pindah dalam tipe yang sama (tipe diturunkan dari digit pertama code, jadi pindah tipe = ganti code = operasi berat)?
3. **Search: filter per grup atau flatten?** Apakah hasil search tetap tampil dalam struktur grup-tree, atau sebagai daftar datar (flat) agar lebih cepat dipindai?
4. **Kapan `/summary` lama dilangsingkan?** Setelah frontend pindah ke `/tree` (disarankan), atau sekaligus dengan fallback sementara?

---

## API Endpoints

### Saat Ini

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `GET` | `/api/v1/accounts/summary` | `GetSummary` (counts + tree, kegemukan) | ✅ Berfungsi |
| `GET` | `/api/v1/accounts/posting` | `GetPostingAccounts` | ✅ Berfungsi |
| `POST` | `/api/v1/accounts` | `Create` (validasi parent lemah) | ✅ Ada |
| `PUT` | `/api/v1/accounts/{id}` | `Update` (validasi parent lemah) | ✅ Ada |
| `DELETE` | `/api/v1/accounts/{id}` | `Delete` (di luar scope UI) | ✅ Ada, guard 409 mati |

### Usulan (Setelah F1)

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `GET` | `/api/v1/accounts/summary` | `GetSummary` (hanya 3 angka) | 🔄 Dilangsingkan |
| `GET` | `/api/v1/accounts/tree` | `GetTree` (groups + tree + postingCount) | ❌ Baru |
| `GET` | `/api/v1/accounts/posting` | `GetPostingAccounts` | ✅ Tetap |
| `POST` | `/api/v1/accounts` | `Create` (+ validasi parent) | 🔄 Diperketat |
| `PUT` | `/api/v1/accounts/{id}` | `Update` (+ validasi parent) | 🔄 Diperketat |

Kontrak JSON usulan:

```json
// GET /summary (ringan)
{ "totalAccounts": 132, "postingAccounts": 96, "headerAccounts": 36 }

// GET /tree (baru)
{ "groups": [{ "type": "ASSET", "label": "Aset", "icon": "wallet",
  "count": 42, "postingCount": 30, "accounts": [ "...tree..." ] }] }
```

---

## Database Tables

### accounts (relevan untuk validasi UI)

```sql
CREATE TABLE accounts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT NOT NULL,
    name       TEXT NOT NULL,
    type       TEXT GENERATED ALWAYS AS (... digit pertama ...) STORED,
    level      SMALLINT GENERATED ALWAYS AS (... segmen non-00 ...) STORED,
    parent_id  BIGINT REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_accounts_code_format CHECK (code ~ '^[1-6]\.\d{2}\.\d{2}\.\d{2}$'),
    CONSTRAINT uq_accounts_user_code UNIQUE (user_id, code)
);
```

Implikasi: `type`/`level` tidak diinput (diturunkan dari `code`); UI cukup minta `code` + `name` + `parent`, backend menurunkan sisanya. Konflik kode → 409 (`ErrCodeExists`).

---

## Service Interface

```go
// Saat ini
type Service interface {
    GetSummary(ctx context.Context, userID int64) (*AccountSummary, error)
    Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error)
    Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error)
    Delete(ctx context.Context, userID int64, accountID int64) error
    GetPostingAccounts(ctx context.Context, userID int64) ([]PostingAccountResponse, error)
}

// Usulan (F1+F2)
type Service interface {
    GetSummary(ctx context.Context, userID int64) (*AccountSummary, error) // hanya counts
    GetTree(ctx context.Context, userID int64) (*AccountTree, error)       // BARU
    Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error)
    Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error)
    Delete(ctx context.Context, userID int64, accountID int64) error
    GetPostingAccounts(ctx context.Context, userID int64) ([]PostingAccountResponse, error)
}
```

---

## Aturan Proyek yang Mengikat

- **Kalkulasi di API**: `postingCount`, tree, grouping tetap di backend. Search/filter tampilan + expand-state boleh di frontend.
- **HTTP**: semua panggilan via `authHttpClient()` dari `@/lib/http-client` (jangan raw `fetch`).
- **Logger**: `createLogger("account.*")` di setiap service/komponen/sheet baru; backend `logger.FromContext(ctx)`.
- **i18n**: semua string baru wajib ada di `id.json` + `en.json` (tidak ada hardcode).
- **Timezone**: N/A — akun tidak punya field tanggal.
- **Eksekusi bertahap**: setiap tahap (1/2/3) dipresentasikan file-by-file dan menunggu persetujuan sebelum eksekusi, sesuai aturan konfirmasi wajib.
