# Review Fitur Aset

**Tanggal**: 16 September 2026
**Status**: ✅ Tahap 0 (preview dummy) Selesai · ✅ Tahap 1 (backend read: GET /assets/summary + GET /assets/groups) Selesai · ✅ Tahap 2 (frontend wiring + hapus dummy) Selesai · ✅ A10 ikon Batal (dihapus total dari halaman Aset) · ✅ Tahap 3 (kualitas: error/search/aria/i18n + skeleton baris) Selesai · ✅ Tahap 4 (mutasi: create + adjust-balance) Selesai
**Keputusan Desain (16 Sep 2026)**: (1) endpoint nested di modul `account` (`/api/v1/accounts/assets/...`, kode Go di file `asset_*` terpisah dalam package `account`); (2) `currentAsset` = subtree `1.01.00.00`; (3) grouping = akun level-2; (4) halaman bisa tambah aset (auto `accounts` + `account_balances`), rename, dan adjust-balance via jurnal lawan `6.01.01.01 Reclass Adjustments`; (5) ikon dihapus total dari halaman Aset (A10 batal)
**Dokumen Terkait**: [Review Fitur Akun](./REVIEW_FITUR_AKUN.md), [Review Fitur Jurnal](./REVIEW_FITUR_JURNAL.md), [Architecture Issues](../issues/ARCHITECTURE_ISSUES_2026-09-11.md)

---

## Ringkasan

Review fitur halaman Aset mencakup analisis API (Go), Frontend (Next.js), Database, dan identifikasi gap antara implementasi saat ini dengan kebutuhan fitur. Saat ini halaman Aset adalah **viewer read-only dari dummy statis**: `AssetTopbar` (2 angka) + `AssetList` (grup + sub-akun) + `AssetGroupCard`, semuanya dari `dummyAssetSummary` di `web/lib/dummy-data.ts` dengan delay 1 detik simulasi loading. Tidak ada endpoint aset di `api/internal/module/account/handler.go`, tidak ada query join balance.

Satu-satunya sumber saldo nyata adalah `account_balances (account_id PK, balance NUMERIC)` yang dipelihara modul journal (`buildBalanceDeltas` dengan faktor tanda ASSET = `debit - credit`, `ApplyBalanceDelta` UPSERT, `recalculate_account_balance()` untuk healing) — dan `EnsureBalance` hanya dibuat untuk **Asset posting (level 3)** di `account/service.go:153`. Seeder nyata membagi aset menjadi `1.01 Current Assets` (Bank, Cash, E-Wallet, Brokerage) vs `1.02 Non-Current Assets` (Bonds, Stocks, Mutual Funds, Saving and Comodities, Account Receiveable), yang menjadi acuan definisi `currentAsset`.

Scope yang disetujui: A1 (API summary + groups nested di `account`), A2 (kontrak balance via join + `COALESCE`), A3 (`currentAsset` = subtree `1.01`, diputuskan), A4 (`accountCount` ke backend), A5 (pecah queryKey summary/groups), A6 (error state), A7 (logger + http-client), A8 (i18n), A9 (search + expand + aria), A10 (icon map), A11 (currency), A12 (tambah aset + rename + adjust-balance via jurnal lawan `6.01.01.01`). **Direct `UPDATE account_balances` dilarang** — setiap perubahan saldo harus jurnal berimbang agar trial balance tetap seimbang dan ada audit trail.

---

## Yang Sudah Ada

| Komponen | Status | Lokasi |
|----------|--------|--------|
| Frontend Topbar (2 kartu) | ✅ Dummy preview | `web/components/asset/asset-topbar.tsx` |
| Frontend List + GroupCard | ✅ Dummy preview | `web/components/asset/asset-list.tsx`, `asset-group-card.tsx` |
| Frontend Skeleton | ✅ Ada (header saja) | `web/components/asset/asset-group-card-skeleton.tsx` |
| Frontend EmptyState | ✅ Ada (tanpa error state) | `web/components/asset/asset-list.tsx:31-37` |
| `dummyAssetSummary` (4 grup) | ✅ Sementara, akan dibuang | `web/lib/dummy-data.ts` |
| `assetService.getSummary()` | ✅ Dummy + delay 1s | `web/services/asset.ts` |
| Tipe `AssetSummary`/`AssetGroup`/`AssetAccount` | ✅ Fiktif, tanpa counterpart Go | `web/types/index.ts:154-176` |
| Query key `assets.summary` | ✅ Ada (belum dipecah) | `web/lib/query-keys.ts:44-47` |
| i18n dasar `assetPage` (id + en) | ✅ Ada, minim (7 key) | `web/messages/id.json:139-152` |
| Tabel `account_balances` | ✅ Lengkap | `api/migrations/20260909120006_create_account_balances.sql` |
| Balance write-path (delta UPSERT) | ✅ Berfungsi | `api/internal/module/journal/service.go` (`buildBalanceDeltas`), `api/internal/infrastructure/postgres/journal_repository.go` (`ApplyBalanceDelta`) |
| `EnsureBalance` Asset posting | ✅ Ada | `api/internal/module/account/service.go:153` |
| Seeder aset `1.01`/`1.02` | ✅ Ada | `api/seeders/20260909134631_seed_accounts.sql:19-20,35-43,69-83` |
| Akun lawan `6.01.01.01 Reclass Adjustments` | ⚠️ Ada di seeder admin, **tidak di-`SeedRoots`** | Seeder `:153` ada; `account_repository.go:308-323` (`SeedRoots`) hanya seed 6 root level-0 → user baru tidak punya rantai `6.00→6.01→6.01.01→6.01.01.01`; harus lazy-ensure saat adjust-balance (A12) |
| API Aset (`/api/v1/accounts/assets/*`) | ❌ Belum ada | `api/internal/module/account/handler.go` (7 route existing), mount di `api/internal/server/http.go` |

---

## Critical Issues

### A1. Tidak Ada Backend — Seluruh Halaman dari Dummy Statis
- **Masalah**: `web/services/asset.ts:28-36` me-return `dummyAssetSummary` (sebelumnya bahkan import basi karena simbol belum ada; sudah diperbaiki Tahap 0 agar bisa preview). Tidak ada DTO, handler, route, maupun query join balance untuk aset.
- **Dampak**: Angka bukan milik user; multi-user salah; tidak konsisten dengan jurnal; tidak bisa diverifikasi.
- **File**: `web/services/asset.ts`, `web/lib/dummy-data.ts`, `api/internal/module/account/handler.go`, `api/internal/module/account/service.go`
- **Fix**: Bangun sumber kebenaran di API (Tahap 1). **Keputusan: nested di modul `account`** — path `GET /api/v1/accounts/assets/summary` + `GET /api/v1/accounts/assets/groups`, kode Go di file `asset_*` terpisah dalam package `account` (hybrid: URL nested, file terpisah agar ekstraksi nanti murah). Daftarkan route `/assets/*` **sebelum** `/{id}` di `Routes()` karena `/{id}` greedy di chi. Detail di Plan A1.

### A2. Kontrak Tipe Frontend Tidak Bisa Dipenuhi Backend Saat Ini
- **Masalah**: Frontend mengharapkan `AssetGroup { id, code, name, icon, accounts[], totalBalance }` + `AssetAccount { balance, isPosting }`. `account.Repository` tidak punya method balance (`FindByUserID` kembalikan `[]Account` tanpa balance; `EnsureBalance` hanya insert row 0). Tidak ada query `JOIN accounts ↔ account_balances`.
- **Dampak**: API baru tidak punya data untuk dikembalikan tanpa penambahan repository.
- **File**: `api/internal/module/account/repository.go`, `web/types/index.ts:154-176`
- **Fix**: Tambah method repo mis. `FindAssetWithBalances(ctx, userID)` + DTO Go mirror `AssetSummary`. `balance` yang belum ada row → `COALESCE(balance, 0)`. Detail di Plan A1.

### A3. Semantik `currentAsset` + `totalBalance` — DIPUTUSKAN: `currentAsset` = subtree `1.01`
- **Keputusan (16 Sep 2026)**: `currentAsset` = agregat subtree `1.01.00.00` (Current Assets, diturunkan dari **prefix code**, bukan traversal `parent_id` — murah: `WHERE code LIKE '1.01.%'` dan konsisten dengan generated column `type`/`level`); `totalAsset` = seluruh `type = ASSET`; `totalBalance` per grup = `SUM(balance)` akun posting anggotanya dengan faktor tanda ASSET (`debit - credit`, konsisten dengan `balanceSignFactor` di `journal/service.go:801-802`). User tanpa subtree `1.01` → `currentAsset = 0` (benar, bukan bug). Label lama `"Kas + Bank"` di `id.json:146-149` diganti mengikuti definisi ini saat Tahap 3 (i18n).
- **File**: `web/messages/id.json`, `api/seeders/...sql:19-20`

---

## High Priority Issues

### A4. Kalkulasi di Frontend (`totalAccount` via `reduce`)
- **Masalah**: `asset-topbar.tsx:23-26` menghitung `totalAccount` via `groups.reduce((sum, g) => sum + g.accounts.length, 0)`.
- **Dampak**: Melanggar prinsip "kalkulasi di API" (sama seperti F2 Akun yang sudah diperbaiki menjadi `postingCount` dari backend).
- **File**: `web/components/asset/asset-topbar.tsx`
- **Fix**: Backend kirim `accountCount` (dan semua agregat) di respons summary; frontend hanya render. Detail di Plan Tahap 1/2.

### A5. Topbar dan List Berbagi Satu QueryKey Berat
- **Masalah**: Keduanya pakai `queryKeys.assets.summary` (`asset-topbar.tsx:19`, `asset-list.tsx:15`). Persis masalah F1 Akun sebelum dipecah menjadi `summary` ringan + `tree` berat.
- **Dampak**: Topbar (butuh 2 angka + count) ikut menunggu/mengunduh seluruh grup + akun; refetch topbar = refetch list; susah tambah field grup tanpa membebani topbar.
- **File**: `web/lib/query-keys.ts:44-47`, kedua komponen di atas
- **Fix**: Pecah sejak awal — `assets.summary` (ringan) + `assets.groups` (berat). Jangan ulangi kesalahan F1. Detail di Plan A5.

### A6. Tidak Ada Error State
- **Masalah**: Kedua komponen hanya menangani `isLoading` vs data. Saat fetch gagal (401/500/offline), halaman tampil kosong tanpa pesan. `asset-list.tsx:31` menampilkan EmptyState "Belum ada aset" yang menyesatkan saat error.
- **Dampak**: User mengira data kosong; tidak ada retry. Sama seperti F6 Akun.
- **File**: `web/components/asset/asset-list.tsx`, `web/components/asset/asset-topbar.tsx`
- **Fix**: Error card + tombol retry (`refetch`) pola `account-list.tsx:135-154`; Topbar `isError → "–"` pola `account-topbar.tsx:47-48`. Detail di Plan Tahap 3.

### A7. Tanpa Logger + Tanpa http-client
- **Masalah**: `services/asset.ts` tidak pakai `authHttpClient` maupun `createLogger`; komponen asset tanpa `createLogger` sama sekali. Bandingkan `services/account.ts:14,29-40` dan `account-list.tsx:12`.
- **Dampak**: Melanggar aturan wajib logger + http-client. Auth, retry, logging inkonsisten; sulit debug.
- **File**: `web/services/asset.ts`, `web/components/asset/*`
- **Fix**: `getSummary()/getGroups()` via `authHttpClient()` + `createLogger("asset.*")` di service, topbar, list, card; backend `logger.FromContext(ctx)`. Detail di Plan Tahap 2.

### A8. String i18n Bocor
- **Masalah**: `asset-group-card.tsx:51` `"sub-akun"`, `:93` `"Posting"`, `:104` `"Subtotal debit bersih"` + `"DR "`. `assetPage` hanya 7 key vs `accountPage` yang sudah lengkap (search/error/badge/expand).
- **Dampak**: Locale EN tetap tampil Indonesia (masalah F8 Akun).
- **File**: `web/components/asset/asset-group-card.tsx`, `web/messages/id.json`, `web/messages/en.json`
- **Fix**: Pindahkan ke key `assetPage.badge/groupSubtitle/...` di kedua locale; ganti hardcode dengan `useTranslations`. Detail di Plan Tahap 3.

### A12. Mutasi dari Halaman Aset (tambah + rename + adjust-balance) — DIPUTUSKAN: via akun + jurnal
- **Keputusan (16 Sep 2026)**: halaman Aset bisa (a) tambah aset → otomatis `accounts` + `account_balances` (saldo awal 0; opening balance via adjust susulan), (b) rename → reuse `PUT /api/v1/accounts/{id}` (rename-only, tanpa endpoint baru), (c) update balance → **hanya via jurnal** lawan `6.01.01.01 Reclass Adjustments`. Direct `UPDATE account_balances` **dilarang** (merusak double-entry, trial balance tidak seimbang, tanpa audit, rawan race).
- **Mapping jurnal** (ASSET dan OTHER berfaktor `+1`, jadi `delta = debit - credit`): naikkan aset X → `Dr aset X / Cr 6.01.01.01 X`; turunkan aset X → `Dr 6.01.01.01 X / Cr aset X`. Request: `POST .../assets/{id}/adjust-balance { amount (>0), direction (increase|decrease), note?, date? }`. Saldo negatif diizinkan (dengan konfirmasi UI).
- **Gap yang harus ditutup**: `SeedRoots` tidak membuat rantai `6.00→6.01→6.01.01→6.01.01.01` (hanya admin seeder yang punya) → service adjust-balance wajib **lazy-ensure** rantai tersebut per user dalam tx yang sama (`INSERT ... ON CONFLICT DO NOTHING`, pola `SeedRoots`). Balance row reclass tidak perlu di-ensure manual — `ApplyBalanceDelta` UPSERT otomatis saat jurnal ditulis, dan view aset hanya baca `type=ASSET`.
- **File**: `api/internal/module/account/service.go` (`Create`, `Update` reuse), file baru `asset_adjust.go`, `api/internal/module/journal/service.go` (`Create` reuse atau dependensi), `api/internal/infrastructure/postgres/account_repository.go` (`SeedRoots`/ensure)
- **Fix**: Tahap 4. Tambah aset mensyaratkan `parentId` ASSET **level 2** (menjaga invarian grouping); rename tanpa endpoint baru; adjust-balance satu tx (ensure reclass → create journal 2 lines berimbang → `ApplyBalanceDelta` → audit jurnal). Detail di Plan Tahap 4.

---

## Medium Priority Issues

| Issue | Deskripsi | File |
|-------|-----------|------|
| A9. Tanpa search / expand-all / aria | `AssetGroupCard:21` pakai `useState(true)` lokal default-expanded; tombol tanpa `aria-expanded/aria-controls`; tidak ada search maupun kontrol massal. Akun sudah punya `account-search.tsx`, `expandAll/collapseAll`, controlled `expandedGroups` | `web/components/asset/asset-group-card.tsx`, `asset-list.tsx` |
| ~~A10. `iconMap` rapuh~~ | **BATAL (16 Sep 2026)**: ikon dihapus total dari halaman Aset — tidak ada kolom DB, tidak ada peta backend, tidak ada `iconMap` frontend. Grup level-2 dirender tanpa ikon | — |
| A11. Currency hardcode | `currency: "Rp"` + `compact: true` tersebar di topbar + card | `web/components/asset/asset-topbar.tsx`, `asset-group-card.tsx` |
| Skeleton hanya header | `AssetGroupCardSkeleton` tidak meniru baris sub-akun (bandingkan `AccountTypeGroupSkeleton`) | `web/components/asset/asset-group-card-skeleton.tsx` |

---

## Rekomendasi Implementasi

### Tahap 1: Backend Read (A1, A2, A3, A4) — nested di `account`

**A1 — Endpoint nested, file terpisah (diputuskan):**
1. **[Backend]** Tambah file dalam package `account` (tanpa modul baru, tanpa ubah `server/http.go`): `asset_dto.go` (`AssetSummary` ringan + `AssetGroups` berat), `asset_service.go` (`GetAssetSummary(ctx, userID)`, `GetAssetGroups(ctx, userID)`), `asset_handler.go` (`GetAssetSummary`, `GetAssetGroups`), `asset_repository.go` (interface `FindAssetWithBalances` + query join). Route di `handler.go Routes()`: `r.Get("/assets/summary", ...)` + `r.Get("/assets/groups", ...)` didaftarkan **sebelum** `r.Put("/{id}", ...)` (chi `/{id}` greedy). Service method ditambahkan ke interface `Service` existing.
2. **[Backend]** Repository: satu query `accounts LEFT JOIN account_balances` filter `user_id + type = 'ASSET'`, `COALESCE(balance, 0)`, order by `code`. Agregat: `totalAsset = SUM` semua; `currentAsset = SUM` prefix `1.01.`; `totalBalance` per grup = `SUM` anggotanya; `accountCount` = count posting. Ikuti faktor tanda journal (`balanceSignFactor`: ASSET = `debit - credit`).
3. **[Backend]** Test: `asset_service_test.go` (agregat, grup kosong di-skip, balance NULL → 0, isolasi antar user, `currentAsset` hanya prefix `1.01`).
4. **[Backend]** Grouping = akun **level-2** seeder (Bank, Cash, E-Wallet, Brokerage, Bonds, Stocks, Mutual Funds, Saving and Comodities, Account Receiveable — diputuskan) sebagai `AssetGroup`; bukan 4 grup dummy. Grup dirender **tanpa ikon** (keputusan 16 Sep 2026: A10 batal).

**Keputusan yang sudah dibekukan:** nested di `account`; `currentAsset` = subtree `1.01.00.00`; `totalAsset` = seluruh ASSET; grup = level-2.

### Tahap 2: Wiring Frontend (A4, A5, A7)

5. **[Frontend]** `assetService.getSummary()` + `getGroups()` via `authHttpClient()` + `createLogger("asset.service")`; hapus ketergantungan `dummyAssetSummary` (biarkan export-nya sampai API stabil, lalu hapus).
6. **[Frontend]** Tambah query key `assets.groups` di `lib/query-keys.ts` (`all/summary/groups`). `AssetTopbar` → `assets.summary`, `AssetList` → `assets.groups`. Hapus `reduce` di topbar; render `accountCount` dari API.
7. **[Frontend + Backend]** Logging tiap operasi: `createLogger("asset.*")` di service/topbar/list/card; `logger.FromContext(ctx)` di handler/service Go.

### Tahap 3: Kualitas List (A6, A8, A9, A10, A11)

8. **[Frontend — A6]** Error state: tangkap `isError` di List/Topbar, error card + `"Coba lagi"` (`refetch`); skeleton tetap untuk loading; Topbar error → `"–"`.
9. **[Frontend — A9]** Search input (filter code/nama, auto-expand grup berisi hasil, count `"N hasil"`, tombol clear) + toolbar `"Buka semua / Tutup semua"` + controlled expanded (angkat dari lokal per-card) + `aria-expanded`/`aria-controls`. Filter tampilan boleh di frontend (bukan kalkulasi bisnis).
10. **[Frontend — A8]** Pindahkan `"sub-akun"`, `"Posting"`, `"Subtotal debit bersih"`, `"DR "` ke `assetPage` di `id.json` + `en.json`; ganti hardcode dengan `useTranslations` (tiru `accountPage.badge/groupSubtitle`).
11. **[Frontend — A11]** Pusatkan default currency di `useCurrencyFormatter`; skeleton tambah 2–3 baris sub-akun. (A10 batal — tidak ada ikon.)

### Tahap 4: Mutasi dari Halaman Aset (A12)

12. **[Backend — tambah]** `POST /api/v1/accounts/assets` (body: `parentId` + `name`): reuse `nextChildCode` + `EnsureBalance` + audit dalam satu tx. Validasi: parent wajib milik user, `type = ASSET`, `level == 2` (menjaga invarian grouping; tolak level lain dengan 400). Saldo awal selalu 0 — opening balance dilakukan via adjust susulan (langkah 14) agar Tahap 4 tidak butuh tx lintas service. Error: reuse `ErrParentNotFound` → 404, `ErrCodeExists` → 409, `ErrCodeExhausted` → 422.
13. **[Backend + Frontend — rename]** Tanpa endpoint baru: UI Aset memanggil `PUT /api/v1/accounts/{id}` (rename-only) yang sudah ada; batasi di UI hanya untuk akun `type = ASSET` milik user.
14. **[Backend — adjust-balance]** `POST /api/v1/accounts/assets/{id}/adjust-balance` (body: `amount (>0)`, `direction (increase|decrease)`, `note?`, `date?` RFC3339 dengan offset): dalam satu tx — (a) validasi target (ASSET posting milik user), (b) lazy-ensure rantai `6.00→6.01→6.01.01→6.01.01.01` per user (`ON CONFLICT DO NOTHING`), (c) buat journal entry 2 lines berimbang (naik: `Dr aset / Cr reclass`; turun: `Dr reclass / Cr aset`), (d) `ApplyBalanceDelta`, (e) audit jurnal (reuse jalur journal; tidak perlu audit aset terpisah). Kembalikan journal entry id + balance baru. Larangan: tidak ada path yang menulis `account_balances` langsung.
15. **[Frontend]** `assetService.create()` + `adjustBalance()` via `authHttpClient` + `useMutation` + invalidasi `assets.summary` + `assets.groups` (+ `accounts.*` karena CoA berubah) + toast sonner + key i18n (`assetPage.create/adjust/...`) + `createLogger("asset.mutate")`. Sheet tambah-aset meniru `CreateAccountSheet` (parent picker difilter kandidat ASSET level-2); sheet adjust-balance berisi amount + direction + note dengan konfirmasi untuk saldo negatif.

---

## File yang Perlu Diubah

### Tahap 1 (Backend read — summary + groups, nested di `account`)

| File | Perubahan |
|------|-----------|
| `api/internal/module/account/asset_dto.go` | Baru: `AssetSummary` (ringan: `totalAsset, currentAsset, accountCount`) + `AssetGroups` (berat: `groups[]` + `totalBalance` per grup) |
| `api/internal/module/account/asset_service.go` | Baru: `GetAssetSummary()`, `GetAssetGroups()` (agregat di Go, `COALESCE` NULL → 0, `currentAsset` = prefix `1.01.`) |
| `api/internal/module/account/asset_handler.go` | Baru: `GetAssetSummary` + `GetAssetGroups` (auth + logger + 500 mapping) |
| `api/internal/module/account/handler.go` | Daftarkan `GET /assets/summary` + `GET /assets/groups` **sebelum** `/{id}` |
| `api/internal/module/account/repository.go` | Tambah `FindAssetWithBalances(ctx, userID)` ke interface |
| `api/internal/infrastructure/postgres/account_repository.go` | Implementasi join `accounts ↔ account_balances` + order by `code` |
| `api/internal/module/account/service.go` | Tambah 2 method ke interface `Service` (implementasi di `asset_service.go`) |
| `api/internal/module/account/asset_service_test.go` | Baru: agregat, grup kosong di-skip, isolasi user, prefix `1.01` |

### Tahap 2 (Wiring frontend)

| File | Perubahan |
|------|-----------|
| `web/services/asset.ts` | `getSummary()` + `getGroups()` real via `authHttpClient` + logger; hapus dummy |
| `web/types/index.ts` | Selaraskan `AssetSummary` (tambah `accountCount`) + `AssetGroup` dengan DTO Go |
| `web/lib/query-keys.ts` | Tambah `assets.groups` key |
| `web/components/asset/asset-topbar.tsx` | Pakai `assets.summary` + logger + `isError → "–"` |
| `web/components/asset/asset-list.tsx` | Pakai `assets.groups` |
| `web/lib/dummy-data.ts` | Hapus `dummyAssetSummary` setelah API stabil |

### Tahap 3 (Kualitas)

| File | Perubahan |
|------|-----------|
| `web/components/asset/asset-list.tsx` | Search state, expand-all state, error card + retry |
| `web/components/asset/asset-group-card.tsx` | Controlled `expanded` + `aria-expanded`, string via i18n |
| `web/components/asset/asset-group-card-skeleton.tsx` | Tambah baris skeleton sub-akun |
| `web/messages/id.json`, `web/messages/en.json` | Key search, error, expand, badge, subtotal |

### Tahap 4 (Mutasi — tambah + rename + adjust)

| File | Perubahan |
|------|-----------|
| `api/internal/module/account/asset_dto.go` | Tambah `CreateAssetRequest { parentId, name }`, `AdjustBalanceRequest { amount, direction, note?, date? }` |
| `api/internal/module/account/asset_service.go` | Tambah `CreateAsset()` (parent ASSET level-2 + `nextChildCode` + `EnsureBalance` + audit) dan `AdjustBalance()` (ensure rantai `6.01.01.01` → journal 2 lines → `ApplyBalanceDelta` → audit) |
| `api/internal/module/account/asset_handler.go` | Tambah `POST /assets` + `POST /assets/{id}/adjust-balance` (sebelum `/{id}`); rename reuse `PUT /{id}` existing |
| `api/internal/infrastructure/postgres/account_repository.go` | Tambah ensure rantai reclass per user (`ON CONFLICT DO NOTHING`) |
| `web/services/asset.ts` | Tambah `create()` + `adjustBalance()` + logging |
| `web/components/asset/create-asset-sheet.tsx` | Baru: form parent (ASSET level-2) + nama (pola `CreateAccountSheet`) |
| `web/components/asset/adjust-balance-sheet.tsx` | Baru: form amount + direction + note + konfirmasi saldo negatif |
| `web/app/[locale]/(auth)/aset/page.tsx` | Tambah FAB + mount sheets |
| `web/messages/id.json`, `web/messages/en.json` | Key create/adjust, validasi, toast |

---

## Pertanyaan untuk Diputuskan (status 16 Sep 2026)

1. ✅ **Modul `asset` baru vs endpoint di `account`?** DIPUTUSKAN: nested di modul `account` (`/api/v1/accounts/assets/...`), file Go `asset_*` terpisah dalam package `account`.
2. ✅ **`currentAsset` = subtree `1.01.00.00`?** DIPUTUSKAN: ya (prefix code `1.01.`).
3. ✅ **Grouping ikut level-2 seeder atau 4 grup dummy?** DIPUTUSKAN: level-2 seeder.
4. ✅ **Pisah `summary`/`groups` sejak awal?** DIPUTUSKAN: ya.
5. ✅ **Halaman Aset read-only atau bisa mutasi?** DIPUTUSKAN: bisa tambah (saldo awal 0) + rename (reuse) + adjust-balance via jurnal lawan `6.01.01.01` (Tahap 4).
6. ✅ **Opening balance sekaligus saat create?** DIPUTUSKAN (16 Sep 2026): **Opsi A — tidak**. Create selalu saldo 0 (`EnsureBalance`); saldo awal diisi via adjust susulan sebagai jurnal terpisah. Alasan: tiap request satu domain (account saja / journal saja), tanpa tx lintas service. UX disamarkan: setelah create sukses tampilkan tombol "Isi saldo awal" yang membuka sheet adjust.

---

## API Endpoints

### Saat Ini (tidak ada endpoint aset)

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| — | `/api/v1/accounts/assets/*` | — | ❌ Belum ada (frontend pakai dummy) |

### Usulan — Tahap 1 (read)

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `GET` | `/api/v1/accounts/assets/summary` | `GetAssetSummary` (hanya angka + count) | ❌ Baru |
| `GET` | `/api/v1/accounts/assets/groups` | `GetAssetGroups` (grup level-2 + akun + balance) | ❌ Baru |

### Usulan — Tahap 4 (mutasi)

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `POST` | `/api/v1/accounts/assets` | `CreateAsset` (parent ASSET level-2, saldo awal 0) | ❌ Baru |
| `PUT` | `/api/v1/accounts/{id}` | `Update` (reuse rename-only untuk aset) | ✅ Ada |
| `POST` | `/api/v1/accounts/assets/{id}/adjust-balance` | `AdjustAssetBalance` (jurnal lawan `6.01.01.01`) | ❌ Baru |

> Urutan registrasi di `Routes()`: `/assets/*` sebelum `/{id}` (chi greedy).

Kontrak JSON usulan:

```json
// GET /api/v1/accounts/assets/summary (ringan — untuk Topbar)
{ "totalAsset": 132250000, "currentAsset": 105750000, "accountCount": 10 }

// GET /api/v1/accounts/assets/groups (berat — untuk List)
{ "groups": [{ "id": 1, "code": "1.01.01.00", "name": "Bank",
  "totalBalance": 101600000, "accounts": [
    { "id": 101, "code": "1.01.01.01", "name": "Mandiri - Main",
      "balance": 49100000, "isPosting": true } ] }] }

// POST /api/v1/accounts/assets
{ "parentId": 12, "name": "Jago - Baru" }
// → 201 { "id": 110, "code": "1.01.01.09", ..., } (balance 0 via EnsureBalance)

// POST /api/v1/accounts/assets/{id}/adjust-balance
{ "amount": 500000, "direction": "increase", "note": "Setoran awal", "date": "2026-09-16T10:00:00+07:00" }
// → 201 { "journalEntryId": 77, "balance": 500000 }
// (naik: Dr aset / Cr 6.01.01.01; turun: Dr 6.01.01.01 / Cr aset)
```

---

## Database Tables

### accounts (filter `type = ASSET`, grouping dari `code`)
```sql
CREATE TABLE accounts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       TEXT NOT NULL,   -- mis. '1.01.01.01' (digit pertama '1' = ASSET)
    name       TEXT NOT NULL,
    type       TEXT GENERATED ALWAYS AS (... digit pertama ...) STORED,
    level      SMALLINT GENERATED ALWAYS AS (... segmen non-00 ...) STORED,
    parent_id  BIGINT REFERENCES accounts(id),
    ...
    CONSTRAINT chk_accounts_code_format CHECK (code ~ '^[1-6]\.\d{2}\.\d{2}\.\d{2}$'),
    CONSTRAINT uq_accounts_user_code UNIQUE (user_id, code)
);
```

### account_balances (satu-satunya sumber saldo; hanya Asset posting punya row)
```sql
CREATE TABLE account_balances (
    account_id      BIGINT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    balance         NUMERIC(19,4) NOT NULL DEFAULT 0,  -- (debit - credit) untuk ASSET
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Implikasi: query Aset = `accounts WHERE user_id + type='ASSET' LEFT JOIN account_balances` dengan `COALESCE(balance, 0)` (akun header/level<3 tidak punya row). Tanda balance mengikuti `balanceSignFactor` journal (ASSET = `debit - credit`). Seeder: `1.01.*` = lancar, `1.02.*` = non-lancar.

---

## Service Interface

```go
// Saat ini (package account)
type Service interface {
    GetSummary(ctx context.Context, userID int64) (*AccountSummary, error)
    GetTree(ctx context.Context, userID int64) (*AccountTree, error)
    GetParents(ctx context.Context, userID int64, level int8) ([]ParentListItem, error)
    SeedRoots(ctx context.Context, userID int64) error
    Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error)
    Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error)
    Delete(ctx context.Context, userID int64, accountID int64) error
    GetPostingAccounts(ctx context.Context, userID int64) ([]PostingAccountResponse, error)
}

// Usulan Tahap 1 (tambah 2 method read; implementasi di asset_service.go)
type Service interface {
    // ... existing di atas ...
    GetAssetSummary(ctx context.Context, userID int64) (*AssetSummary, error) // hanya angka
    GetAssetGroups(ctx context.Context, userID int64) (*AssetGroups, error)   // grup level-2 + balance
}

// Usulan Tahap 4 (tambah 2 method mutasi; rename reuse Update)
type Service interface {
    // ... di atas ...
    CreateAsset(ctx context.Context, userID int64, req CreateAssetRequest) (*AccountResponse, error)
    AdjustAssetBalance(ctx context.Context, userID int64, assetID int64, req AdjustBalanceRequest) (*AdjustBalanceResponse, error)
}
```

---

## Aturan Proyek yang Mengikat

- **Kalkulasi di API**: `totalAsset`, `currentAsset`, `totalBalance`, `accountCount` dihitung di backend (SQL/Go). Search/filter tampilan + expand-state boleh di frontend.
- **HTTP**: semua panggilan via `authHttpClient()` dari `@/lib/http-client` (jangan raw `fetch`).
- **Logger**: `createLogger("asset.*")` di setiap service/komponen baru; backend `logger.FromContext(ctx)`.
- **i18n**: semua string baru wajib ada di `id.json` + `en.json` (tidak ada hardcode).
- **Timezone**: N/A untuk Tahap 1–3 — aset tidak punya field tanggal (label periode Topbar pakai `format.dateTime(new Date(), ...)` seperti Akun, zona lokal client).
- **Eksekusi bertahap**: setiap tahap dipresentasikan file-by-file dan menunggu persetujuan sebelum eksekusi, sesuai aturan konfirmasi wajib.
