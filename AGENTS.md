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

### Ketika Bekerja di `api/` (Backend)

1. **Semua logika bisnis ada di sini** — kalkulasi, validasi, transformasi
2. **Response harus sudah siap ditampilkan** — frontend tidak perlu proses ulang
3. **Gunakan layer yang tepat** — handler → service → repository

### Ketika Diminta Membuat Fitur Baru

1. **Identifikasi kalkulasi yang dibutuhkan** — tentukan apakah termasuk bisnis logic
2. **Buat endpoint API** — untuk data yang sudah diproses
3. **Implement frontend** — hanya untuk presentasi dan interaksi
4. **Jangan campur aduk** — pisahkan jelas antara presentation logic dan business logic

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

## Tech Stack Reference

| Layer | Technology | Location |
|-------|-----------|----------|
| Frontend | Next.js, React, TypeScript | `/web` |
| Backend | Go | `/api` |
| Database | (see migrations) | `/api/migrations` |
| Styling | Tailwind CSS, shadcn/ui | `/web/components` |
