# Upload Jurnal — Excel & CSV Integration

Dokumentasi ini menjelaskan fitur upload jurnal massal melalui file Excel (.xlsx) atau CSV.

---

## Table of Contents

- [Overview](#overview)
- [API Endpoints](#api-endpoints)
- [Download Template](#download-template)
- [Upload File](#upload-file)
- [CSV Format](#csv-format)
- [Excel Format](#excel-format)
- [Error Handling](#error-handling)
- [Frontend Integration](#frontend-integration)
- [Testing](#testing)

---

## Overview

PengBook mendukung upload jurnal massal melalui file **Excel (.xlsx)** atau **CSV**.

```
┌─────────────────────────────────────────────────────────────────┐
│  User Flow                                                      │
├─────────────────────────────────────────────────────────────────┤
│  1. Download template Excel (2 sheets)                          │
│  2. Isi data di Excel                                           │
│  3. Upload file Excel atau CSV                                  │
│  4. Backend parse & validate                                    │
│  5. Simpan ke database                                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/journals/template` | Download Excel template |
| `POST` | `/api/v1/journals/upload` | Upload CSV atau Excel file |

---

## Download Template

### Request

```
GET /api/v1/journals/template
Authorization: Bearer <token>
```

### Response

- **Content-Type**: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- **Content-Disposition**: `attachment; filename=template-jurnal.xlsx`
- **Body**: Binary Excel file

### Template Structure

```
┌─────────────────────────────────────────────────────────────────┐
│  File: template-jurnal.xlsx                                     │
├─────────────────────────────────────────────────────────────────┤
│  Sheet 1: "Template Jurnal"                                     │
│  ┌──────────┬────────────┬───────────┬───────┬────────┐         │
│  │ Tanggal  │ Deskripsi  │ Kode Akun │ Debit │ Kredit │         │
│  ├──────────┼────────────┼───────────┼───────┼────────┤         │
│  │ 2026-04-21│ Contoh    │ 1.01.01.01│ 500000│ 0      │         │
│  │ 2026-04-21│ Contoh    │ 4.01.01.01│ 0     │ 500000 │         │
│  └──────────┴────────────┴───────────┴───────┴────────┘         │
│                                                                 │
│  Sheet 2: "Daftar Akun"                                         │
│  ┌───────────┬────────────────────┬────────┬─────────┬────────┐ │
│  │ Kode Akun │ Nama Akun          │ Tipe   │ Level   │ Posting│ │
│  ├───────────┼────────────────────┼────────┼─────────┼────────┤ │
│  │ 1.01.01.01│ Kas - Main         │ ASSET  │ 3       │ Ya    │ │
│  │ 4.01.01.01│ Pendapatan Jasa    │ REVENUE│ 3       │ Ya    │ │
│  └───────────┴────────────────────┴────────┴─────────┴────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Contoh Implementasi (Frontend)

```typescript
// Download template
const blob = await journalService.downloadTemplate();
const url = URL.createObjectURL(blob);
const a = document.createElement("a");
a.href = url;
a.download = "template-jurnal.xlsx";
a.click();
URL.revokeObjectURL(url);
```

---

## Upload File

### Request

```
POST /api/v1/journals/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <CSV or Excel file>
```

### Response

```json
// Success
{
  "success": true,
  "data": {
    "count": 5
  }
}

// Error
{
  "success": false,
  "message": "invalid CSV: row 3: date is required"
}
```

### Contoh Implementasi (Frontend)

```typescript
async function uploadFile(file: File) {
  const formData = new FormData();
  formData.append("file", file);

  const res = await client.post("/api/v1/journals/upload", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });

  return res.data.data; // { count: number }
}
```

---

## CSV Format

### Template

```csv
Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Pembelian perlengkapan,5.01.01.04,500000,0
2026-04-21,Pembelian perlengkapan,1.01.02.01,0,500000
2026-04-20,Pendapatan jasa,1.01.01.02,2000000,0
2026-04-20,Pendapatan jasa,4.01.01.01,0,2000000
```

### Kolom

| Kolom | Wajib | Deskripsi |
|-------|-------|-----------|
| `Tanggal` | ✅ | Format: `YYYY-MM-DD` |
| `Deskripsi` | ❌ | Deskripsi jurnal |
| `Kode Akun` | ✅ | Kode akun (4 segmen: `X.XX.XX.XX`) |
| `Debit` | ✅ | Jumlah debit (0 jika tidak ada) |
| `Kredit` | ✅ | Jumlah kredit (0 jika tidak ada) |

---

## Excel Format

### Sheet 1: Template Jurnal

| Tanggal | Deskripsi | Kode Akun | Debit | Kredit |
|---------|-----------|-----------|-------|--------|
| 2026-04-21 | Pembelian perlengkapan | 5.01.01.04 | 500000 | 0 |
| 2026-04-21 | Pembelian perlengkapan | 1.01.02.01 | 0 | 500000 |

### Sheet 2: Daftar Akun

| Kode Akun | Nama Akun | Tipe | Level | Posting |
|-----------|-----------|------|-------|---------|
| 1.01.01.01 | Kas - Main | ASSET | 3 | Ya |
| 4.01.01.01 | Pendapatan Jasa | REVENUE | 3 | Ya |

### Catatan

- Sheet 1 adalah sheet utama yang di-parse
- Sheet 2 adalah referensi (tidak di-parse)
- Header row otomatis di-skip jika mengandung "Tanggal" atau "Date"
- Baris kosong (Debit=0, Kredit=0) di-skip
- Mendukung format tanggal: `YYYY-MM-DD`, `DD/MM/YYYY`, `MM/DD/YYYY`, RFC3339
- Mendukung angka dengan koma: `1,000,000`

---

## Aturan

1. **Minimal 2 baris** per entri jurnal
2. **Debit harus sama dengan Kredit** (balanced)
3. **Kode akun harus posting** (level 3)
4. **Grouping** otomatis berdasarkan Tanggal + Deskripsi

---

## Error Handling

### Common Errors

| Error | Penyebab | Solusi |
|-------|----------|--------|
| `file is required` | Tidak ada file | Upload file |
| `file must be a CSV (.csv) or Excel (.xlsx)` | Format salah | Gunakan .csv atau .xlsx |
| `CSV file is empty` / `Excel sheet is empty` | File kosong | Isi data |
| `row X: date is required` | Kolom tanggal kosong | Isi tanggal |
| `row X: account code is required` | Kolom kode akun kosong | Isi kode akun |
| `row X: invalid date` | Format tanggal salah | Gunakan YYYY-MM-DD |
| `row X: invalid debit amount` | Format angka salah | Gunakan angka |
| `entry must have at least 2 lines` | Kurang dari 2 baris | Tambah baris |
| `journal entry lines are not balanced` | Debit != Credit | Samakan jumlah |
| `account is not a posting account` | Akun level < 3 | Gunakan akun posting |

---

## Frontend Integration

### File Upload Component

```typescript
// Accept both CSV and Excel
<input
  type="file"
  accept=".csv,.xlsx,.xls"
  onChange={handleFileChange}
/>

// Validate file
function validateAndSetFile(f: File) {
  const validExtensions = [".csv", ".xlsx", ".xls"];
  const ext = "." + f.name.split(".").pop()?.toLowerCase();

  if (!validExtensions.includes(ext)) {
    toast.error("File must be CSV or Excel");
    return;
  }
  if (f.size > 5 * 1024 * 1024) {
    toast.error("File size must be less than 5MB");
    return;
  }
  setFile(f);
}
```

### Download Template

```typescript
async function downloadTemplate() {
  const blob = await journalService.downloadTemplate();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "template-jurnal.xlsx";
  a.click();
  URL.revokeObjectURL(url);
}
```

---

## Testing

### Backend Tests

```bash
# Run all journal tests
go test ./internal/module/journal/ -v

# Run specific tests
go test ./internal/module/journal/ -v -run "TestParseExcel"
go test ./internal/module/journal/ -v -run "TestHandler_DownloadTemplate"
go test ./internal/module/journal/ -v -run "TestHandler_Upload_Excel"
```

### Test Coverage

| Layer | Tests | Status |
|-------|-------|--------|
| CSV Parser | 12 tests | ✅ |
| Excel Parser | 10 tests | ✅ |
| Handler (Download Template) | 3 tests | ✅ |
| Handler (Upload Excel) | 3 tests | ✅ |
| Handler (Upload CSV) | 8 tests | ✅ |
| Service | 12 tests | ✅ |
| Repository | 7 tests | ✅ |
| Domain | 7 tests | ✅ |
| **Total** | **59 tests** | ✅ All pass |

---

## Files

| File | Description |
|------|-------------|
| `api/internal/module/journal/csv_parser.go` | CSV parsing logic |
| `api/internal/module/journal/excel_parser.go` | Excel parsing & template generation |
| `api/internal/module/journal/handler.go` | HTTP handlers |
| `api/internal/module/journal/service.go` | Business logic |
| `web/components/journal/upload-form.tsx` | Upload form component |
| `web/services/journal.ts` | API service |
| `web/docs/UPLOAD_JURNAL.md` | This documentation |

---

## References

- [excelize Documentation](https://xuri.me/excelize/)
- [CSV Parser Source](../../api/internal/module/journal/csv_parser.go)
- [Excel Parser Source](../../api/internal/module/journal/excel_parser.go)
