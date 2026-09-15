# Timezone: Olah Ikut Client, Simpan Tetap UTC

**Tanggal**: 16 September 2026
**Status**: Implemented
**Aturan**: [AGENTS.md](../../AGENTS.md) — seksi `TIMEZONE RULE`

> Database selalu menyimpan waktu sebagai UTC (`timestamptz`), tetapi semua
> pemaknaan kalender (hari, bulan, format tanggal) mengikuti timezone client
> pengirim request.

---

## 1. Kontrak per endpoint

| Endpoint | Field | Format | Makna zona |
|---|---|---|---|
| `POST /journals` (body) | `date` | RFC3339 + offset | Instant absolut; offset ikut tersimpan sebagai titik waktu |
| `PUT /journals/{id}` (body) | `date` | RFC3339 + offset | Sama |
| `GET /journals` (query) | `startDate`, `endDate` | RFC3339 (instant) | Boundary eksak; **invalid → 400** (`ErrInvalidFilter`) |
| `GET /journals` | `cursor`/`nextCursor` | `2006-01-02T15:04:05.000Z_<id>` (UTC, opaque) | Milik server; rusak → 400 (mencegah loop halaman 1) |
| `GET /journals/summary` | `month` | **RFC3339 penuh**, offset wajib ikut | Bulan diturunkan di offset pengirim (lihat §3) |
| `GET /journals/export` | `startDate`, `endDate`, header `X-Timezone` | Instant + nama zona IANA | Filter = instant; **render tanggal file = zona header** |
| `POST /journals/upload` | kolom Tanggal + header `X-Timezone` | String tanggal-only / serial Excel | Midnight **di zona header**; RFC3339 ber-offset absolut |
| Semua response | `date`, `datetime`, `createdAt`, `updatedAt` | RFC3339 | Instant untuk display di zona lokal |

---

## 2. Frontend (`web/`)

### 2.1 Tiga helper tanggal — jangan tertukar

| Helper (`lib/utils.ts`) | Output | Pakai untuk |
|---|---|---|
| `toISOString()` (bawaan JS) | `...Z` (UTC, offset hilang) | Boundary instant murni: filter list/export, tanggal transaksi (`dateInputToISO` membungkusnya) |
| `localISO(date)` | `...+07:00` (offset lokal ikut) | "Bulan apa menurut user" → param `month` summary |
| `getTimeZone()` | `"Asia/Jakarta"` (IANA, fallback `"UTC"`) | Header via middleware; tidak dipanggil manual per request |

Aturan praktis: **kalau backend harus memenggal kalender (bulan/hari), kirim offset; kalau backend hanya membandingkan instant, `Z` cukup.**

### 2.2 Header otomatis

`http/middlewares/timezone-middleware.ts` — terdaftar di kedua factory
`lib/http-client.ts`, sehingga **semua** request (termasuk multipart upload)
membawa `X-Timezone` tanpa perubahan per-service:

```typescript
headers.set("X-Timezone", getTimeZone());
```

### 2.3 Display selalu zona lokal

```typescript
// ❌ slice mentah = tanggal UTC
const key = j.datetime.slice(0, 10);

// ✅ parse → format lokal (journal-scroll-view.tsx)
import { format } from "date-fns";
const key = format(new Date(j.datetime), "yyyy-MM-dd");
```

### 2.4 Gotcha: encoding query string

ISO mengandung `+` dan `:` — interpolasi mentah merusak (`+07:00` tiba sebagai
` 07:00` → 400). Selalu `encodeURIComponent` atau `URLSearchParams`:

```typescript
const query = month ? `?month=${encodeURIComponent(month)}` : "";
```

### 2.5 Query key stabil

`monthKey` (`YYYY-MM`, untuk cache) dipisah dari `monthISO` (untuk API).
ISO berdetik tidak boleh jadi query key — cache meledak dan refetch loop.

---

## 3. Backend (`api/`)

### 3.1 Middleware zona — `internal/middleware/timezone.go`

```go
r.Use(inner.Timezone) // internal/server/http.go, global
```

- Baca `X-Timezone` → `time.LoadLocation` → simpan `*time.Location` di context.
- Header hilang/invalid → **UTC + warn**, request tetap jalan (rendering itu presentasional).
- Akses: `middleware.LocationFromContext(ctx)` — tidak pernah nil.
- Jangan lupa: header custom harus terdaftar di `cors.go` (`Access-Control-Allow-Headers`), atau preflight browser gagal.

### 3.2 Batas bulan dari offset pengirim — `service.go` (`GetTotalSummary`)

```go
now, err := time.Parse(time.RFC3339, month) // offset +07:00 dipertahankan
loc := now.Location()
start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
end := start.AddDate(0, 1, 0)
```

Dilarang hardcode zona di service. Contoh yang dikunci test:
`2026-10-01T00:30:00+07:00` (masih 30 Sep dalam UTC) → bulan `2026-10`.

### 3.3 Export — tanggal file ikut header

`GenerateExport(rows, accounts, loc)` menerima lokasi dari context (bukan
hardcode). Berlaku juga untuk nama file. Contoh yang dikunci test: instant
yang sama dirender `2026-03-10` di `Asia/Jakarta`, `2026-03-09` dalam UTC.

### 3.4 Upload — tanggal-only = midnight zona klien

`ParseCSV`/`ParseExcel(reader, loc)` memakai `time.ParseInLocation`:
layout ber-offset (RFC3339) tetap absolut (offset menang — perilaku lama
tidak berubah); tanggal polos dan serial Excel jatuh ke midnight `loc`.
Test parser lama mempassing `time.UTC` sehingga perilakunya identik.

### 3.5 Filter invalid → 400, bukan diam-diam

`ErrInvalidFilter` untuk `startDate`/`endDate`/`cursor` tak terparse
(list + export). Alasan: filter bocor = data di luar permintaan ikut
terkirim; cursor rusak = client loop halaman 1 selamanya.

### 3.6 Database tetap UTC

Semua kolom waktu `timestamptz` — Postgres menyimpan UTC apa pun zona
sesi. Tidak ada tanggal kalender tanpa offset yang disimpan.

---

## 4. Test yang mengunci perilaku

| Test | Membuktikan |
|---|---|
| `TestHandler_Summary_ClientTimezone` | Offset +07:00 menembus batas bulan UTC dengan benar |
| `TestHandler_Export_Timezone` | Instant sama → hari beda di Auckland vs UTC vs Jakarta; header invalid → fallback |
| `TestHandler_GetAllScrollView_InvalidFilter` | Filter/cursor rusak → 400 |
| `TestGenerateExport_RoundTrip` | Export → `ParseExcel` → entries identik (upload ulang valid) |
| `middleware/timezone_test.go` | Header valid dipakai; hilang/invalid → UTC tanpa gagalkan request |

---

## 5. Checklist untuk endpoint tanggal baru

1. Terima RFC3339 penuh (jangan minta `YYYY-MM` / tanggal polos tanpa zona).
2. Kalau memenggal kalender → turunkan dari zona pengirim, jangan hardcode.
3. Kalau merender tanggal → ambil dari `LocationFromContext`, fallback UTC + warn.
4. Filter tak terparse → `ErrInvalidFilter` (400), jangan abaikan diam-diam.
5. Frontend: pilih helper §2.1 yang tepat; encode query; query key stabil.
