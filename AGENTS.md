# Pengbook Project Rules

## Project Overview

Monorepo dengan dua komponen utama:
- **`api/`** — Backend API (Go) — sumber kebenaran untuk semua logika bisnis dan kalkulasi
- **`web/`** — Frontend (Next.js) — presentasi dan interaksi user saja

---

## CRITICAL RULE: Frontend Tidak Boleh Melakukan Kalkulasi

**PRINSIP UTAMA: Semua logika data dan kalkulasi HARUS berada di API (backend).**

### Apa yang Termasuk Kalkulasi/Data Logic (HARUS di API)

- Kalkulasi harga, diskon, pajak, total
- Agregasi data (sum, count, average, dll)
- Filtering dan sorting data kompleks
- Validasi bisnis rules
- Transformasi data sebelum disimpan
- Pagination logic
- Authentication & authorization logic
- CRUD operations melalui database

### Apa yang Boleh di Frontend

- Rendering UI berdasarkan data yang sudah diproses
- State management untuk UI (form inputs, modals, dll)
- Format tampilan (tanggal, mata uang, dll) — asalkan datanya sudah dihitung dari API
- Event handling dan navigasi
- Optimistic updates untuk UX (dengan tetap melakukan sync ke API)

### Contoh

```typescript
// ❌ FRONTEND — TIDAK BOLEH
const total = items.reduce((sum, item) => sum + item.price * item.quantity, 0)
const discountedTotal = total - (total * discountRate / 100)

// ✅ FRONTEND — BOLEH (data sudah dihitung API)
const { total, discountedTotal } = orderSummary // dari API response

// ✅ FRONTEND — BOLEH (format tampilan saja)
const formattedPrice = new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(price)
```

```go
// ✅ API (Go) — TEMPAT KALKULASI
func CalculateOrderTotal(items []OrderItem, discountRate float64) OrderSummary {
    var total float64
    for _, item := range items {
        total += item.Price * float64(item.Quantity)
    }
    discount := total * discountRate / 100
    return OrderSummary{
        Total:           total,
        Discount:        discount,
        DiscountedTotal: total - discount,
    }
}
```

```typescript
// ❌ FRONTEND — TIDAK BOLEH (raw fetch)
const response = await fetch(`${API_BASE}/journals`);
const result = await response.json();

// ✅ FRONTEND — BOLEH (gunakan http-client)
import { createHttpClient } from "@/lib/http-client";

const client = createHttpClient(getToken);
const res = await client.get<ApiResponse<JournalEntryListResponse>>(`/api/v1/journals`);
return res.data.data;
```

---

## Agent Guidelines

### Ketika Bekerja di `web/` (Frontend)

1. **Jangan tambahkan kalkulasi bisnis** — jika butuh kalkulasi, buat endpoint API baru
2. **Gunakan data yang sudah diproses** —召 response dari API, jangan hitung ulang
3. **Fokus di UI/UX** — component structure, styling, user interaction
4. **Jika butuh data baru** — minta buat endpoint API baru, bukan kalkulasi di frontend
5. **Gunakan http-client untuk API calls** — gunakan `createHttpClient()` dari `@/lib/http-client`, jangan raw `fetch()`. Ini memberikan benefit: auth middleware, logging, retry, dan error handling yang konsisten
6. **Gunakan logger untuk semua operasi** — gunakan `createLogger()` dari `@/lib/logger` untuk logging di setiap komponen, service, dan HTTP interceptor. Ini penting untuk debugging di development dan monitoring di production

### Contoh Penggunaan Logger

```typescript
// ✅ Gunakan logger di setiap file
import { createLogger } from "@/lib/logger";

const logger = createLogger("NamaModule");

// Debug: operasi normal, data fetching
logger.debug("Fetching journals", { limit: 10 });

// Info: operasi berhasil
logger.info("Journals fetched", { count: data.length });

// Warn: situasi yang perlu perhatian
logger.warn("Token expiring soon", { expiresIn: 300 });

// Error: kegagalan operasi
logger.error("Failed to fetch journals", { error });
```

### Kapan Harus Log (Frontend)

| Level | Kapan Digunakan | Contoh |
|-------|-----------------|--------|
| `debug` | Operasi normal, debugging | Fetching data, form submission |
| `info` | Operasi berhasil | Data fetched, user logged in |
| `warn` | Situasi perlu perhatian | Token expiring, fallback used |
| `error` | Kegagalan operasi | API call failed, validation error |

### Komponen yang WAJIB Menggunakan Logger (Frontend)

- **Services** (`services/*.ts`) — log semua API calls dan responses
- **HTTP Interceptors** (`http/interceptors/*.ts`) — log auth flow, retries
- **HTTP Middlewares** (`http/middlewares/*.ts`) — log requests
- **Auth Context** (`lib/auth-context.tsx`) — log init, login, logout, refresh
- **Page Components** — log query states, error handling
- **Form Components** — log submissions, validation errors

### Ketika Bekerja di `api/` (Backend)

1. **Semua logika bisnis ada di sini** — kalkulasi, validasi, transformasi
2. **Response harus sudah siap ditampilkan** — frontend tidak perlu proses ulang
3. **Gunakan layer yang tepat** — handler → service → repository
4. **Gunakan logger untuk semua operasi** — gunakan `logger.FromContext(ctx)` dari `pengbook/api/pkg/logger` untuk logging di setiap handler, service, dan repository

### Contoh Penggunaan Logger (Backend)

```go
// ✅ Gunakan logger dari context di setiap handler/service
import "pengbook/api/pkg/logger"

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
    log := logger.FromContext(r.Context())
    
    log.Info("getting journals", "user_id", userID)
    
    result, err := h.service.GetAll(r.Context(), userID)
    if err != nil {
        log.Error("failed to get journals", "error", err)
        response.Error(w, http.StatusInternalServerError, "failed to get journals")
        return
    }
    
    log.Info("journals fetched", "count", len(result.Data))
    response.Success(w, http.StatusOK, result)
}
```

```go
// ✅ Di service layer
func (s *Service) Create(ctx context.Context, userID int64, req CreateRequest) (*Journal, error) {
    log := logger.FromContext(ctx)
    
    log.Debug("creating journal entry", "user_id", userID)
    
    // ... business logic
    
    log.Info("journal entry created", "journal_id", journal.ID)
    return journal, nil
}
```

### Kapan Harus Log (Backend)

| Level | Kapan Digunakan | Contoh |
|-------|-----------------|--------|
| `Debug` | Operasi detail, debugging | Query execution, internal state |
| `Info` | Operasi berhasil | Request handled, data fetched |
| `Warn` | Situasi perlu perhatian | Fallback used, retry attempt |
| `Error` | Kegagalan operasi | DB error, validation failed |

### Komponen yang WAJIB Menggunakan Logger (Backend)

- **Handlers** (`internal/module/*/handler.go`) — log all HTTP requests and responses
- **Services** (`internal/module/*/service.go`) — log business operations
- **Repositories** (`internal/module/*/repository.go`) — log database queries (optional)
- **Middleware** (`internal/middleware/*.go`) — log middleware operations
- **Infrastructure** (`internal/infrastructure/*.go`) — log external service calls

### Konfigurasi Log Level (Backend)

Backend menggunakan **APP_ENV + LOG_LEVEL** untuk menentukan log level.

| APP_ENV | LOG_LEVEL | Hasil |
|---------|-----------|-------|
| `development` | (kosong) | `debug` (otomatis) |
| `production` | (kosong) | `info` (otomatis) |
| `development` | `error` | `error` (manual override) |
| `production` | `debug` | `debug` (manual override) |

**Contoh `.env`:**
```bash
# Development — otomatis debug
APP_ENV=development

# Production — otomatis info
APP_ENV=production

# Override manual
APP_ENV=development LOG_LEVEL=error
```

### Ketika Diminta Membuat Fitur Baru

1. **Identifikasi kalkulasi yang dibutuhkan** — tentukan apakah termasuk bisnis logic
2. **Buat endpoint API** — untuk data yang sudah diproses
3. **Implement frontend** — hanya untuk presentasi dan interaksi
4. **Jangan campur aduk** — pisahkan jelas antara presentation logic dan business logic

---

## TIMEZONE RULE: Olah Ikut Client, Simpan Tetap UTC

**PRINSIP: Database selalu menyimpan waktu sebagai UTC (`timestamptz`), tetapi semua pemaknaan kalender (hari, bulan, format tanggal) mengikuti timezone client pengirim request.**

### Aturan

1. **Database: selalu UTC** — kolom waktu bertipe `timestamptz`; tidak ada tanggal kalender yang disimpan tanpa offset.
2. **Frontend → Backend: kirim instant yang utuh (offset ikut terkirim)** —
   - Instant murni sebagai boundary (filter `startDate`/`endDate`, tanggal transaksi): `toISOString()` boleh dipakai karena boundary yang dimaksud memang instant itu sendiri.
   - "Bulan/hari apa menurut user" (mis. ringkasan bulanan): kirim ISO **dengan offset lokal** via `localISO()` dari `@/lib/utils`, jangan `toISOString()` — `Z` menghilangkan offset sehingga 1 Okt 00:30 +07:00 tiba sebagai 30 Sep 17:30Z dan pemenggalan bulan mendarat di bulan yang salah.
   - Nama zona IANA untuk rendering (mis. export file): `Intl.DateTimeFormat().resolvedOptions().timeZone` via param `tz`.
3. **Backend: turunkan batas dari zona pengirim, jangan hardcode** — parse RFC3339 (`time.Parse` mempertahankan offset numerik), lalu bangun batas hari/bulan **di lokasi offset string itu**. Dilarang hardcode `Asia/Jakarta` di service. Zona tak dikenal → fallback UTC + warn (jangan gagalkan request untuk hal presentasional).
4. **Backend: filter invalid → 400** — tanggal/cursor yang tak terparse tidak boleh diam-diam diabaikan: filter bocor = data di luar permintaan ikut terkirim; cursor rusak = client loop fetch halaman 1 selamanya. Gunakan sentinel error (mis. `ErrInvalidFilter`) agar handler bisa bedakan 400 vs 500.
5. **Frontend display: format di zona lokal** — jangan `slice(0, 10)` string UTC untuk grouping kalender; parse → format `yyyy-MM-dd` lokal (date-fns). Contoh benar ada di `journal-scroll-view.tsx` (`groupByDate`).

### Contoh

```typescript
// ❌ FRONTEND — offset hilang, bulan bisa geser di tengah malam
const month = now.toISOString(); // "2026-09-30T17:30:00.000Z"

// ✅ FRONTEND — zona ikut terkirim
import { localISO } from "@/lib/utils";
const month = localISO(now); // "2026-10-01T00:30:00+07:00"
```

```go
// ❌ BACKEND — hardcode zona server
loc, _ := time.LoadLocation("Asia/Jakarta")

// ✅ BACKEND — zona dari request client
t, _ := time.Parse(time.RFC3339, monthParam) // offset +07:00 dipertahankan
loc := t.Location()
start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
```

---

## Rules Lainnya

Selain rules di atas, ikuti **Global Engineering Principles** dari opencode:
- Mandatory confirmation before execution
- Understand before modifying
- Minimal change
- Correctness over convenience
- Security by default
- Data integrity first

---

## Commit Message

- Setiap sesi/chat **boleh** memberikan rekomendasi commit message — baik diminta
  user maupun proaktif di akhir pekerjaan.
- Rekomendasi **wajib berbasis current changes** (`git status` / `git diff`),
  bukan dari memori atau ringkasan sesi.
- Ikuti gaya repo: `<tipe>(<scope>): <ringkasan>` satu baris (contoh:
  `refactor(web): ...`, `feat(akun): ...`), plus body opsional berisi daftar
  perubahan dan file yang terdampak.
- Rekomendasi ≠ eksekusi: **jangan `commit`, `push`, atau buat PR** tanpa
  instruksi eksplisit dari user.

---

## Tech Stack Reference

| Layer | Technology | Location |
|-------|-----------|----------|
| Frontend | Next.js, React, TypeScript | `/web` |
| Backend | Go | `/api` |
| Database | (see migrations) | `/api/migrations` |
| Styling | Tailwind CSS, shadcn/ui | `/web/components` |

---

## Documentation

| Dokumen | Lokasi | Deskripsi |
|---------|--------|-----------|
| [Logging System](web/docs/LOGGING.md) | `web/docs/LOGGING.md` | Panduan lengkap logging frontend & backend |
| [Logging Issue](web/docs/issues/LOGGING_IMPROVEMENT_2026-09-12.md) | `web/docs/issues/` | Issue logging improvement |
| [HTTP Interceptor](web/docs/HTTP_INTERCEPTOR.md) | `web/docs/` | Arsitektur HTTP interceptor |
| [Architecture Issues](web/docs/issues/ARCHITECTURE_ISSUES_2026-09-11.md) | `web/docs/issues/` | Daftar architectural issues |
