# Arsitektur Pengbook API

## Gambaran Umum

Proyek ini adalah API backend Go dengan pola **layered architecture**:

- **Handler** (HTTP layer) — menerima request, validasi, memanggil service.
- **Service** (business logic) — aturan bisnis + koordinasi transaksi database.
- **Repository** (data access) — akses PostgreSQL. Service hanya bergantung pada *interface* repository, bukan implementasinya.
- **Infrastructure** — implementasi repository dengan driver tertentu (pgx).

## Struktur Folder

```
pengbook-api-go/
├── cmd/api/                    # entry point (main)
├── internal/
│   ├── config/                 # baca env (.env / OS) → struct Config
│   ├── database/               # koneksi pool, DBTX, transaction manager, migration, testutil
│   ├── infrastructure/
│   │   └── postgres/           # implementasi repository pakai pgx
│   ├── middleware/             # chi middleware (logger, recovery)
│   ├── module/                 # modul bisnis per domain
│   │   ├── user/
│   │   ├── auth/               # (belum diimplementasikan)
│   │   └── transaction/        # (belum diimplementasikan)
│   └── server/                 # setup HTTP server (chi)
├── migrations/                 # file SQL migration (di-embed ke binary)
├── pkg/                        # utilitas umum (response, validator, logger)
└── docs/                       # dokumentasi
```

### Tanggung Jawab Per Folder

| Folder | Tanggung jawab |
|---|---|
| `internal/module/<x>/` | Domain entity, DTO, **port** (interface Repository & Service), handler, service logic |
| `internal/infrastructure/postgres/` | Implementasi port `Repository` + `TxManager` (transaction) dengan `pgx/v5` |
| `internal/database/` | Infrastruktur DB: koneksi, interface `DBTX` & `TxManager`, `WithTx`/`GetTx`, migration runner |
| `internal/server/` | Router chi, middleware, graceful shutdown |
| `pkg/` | Utilitas re-usable yang tidak tergantung modul bisnis |

## Alur Request

```
Client
  │  HTTP
  ▼
Handler (internal/module/user/handler.go)
  │  decode JSON + validasi
  ▼
Service (internal/module/user/service.go)
  │  business logic + s.tx.WithTransaction(ctx, fn)
  │  ─ (TxManager) begin tx (jika belum ada) → WithTx(ctx, tx) ─┐
  ▼                                                                │
Repository interface (internal/module/user/repository.go)
  │  db(ctx) → GetTx(ctx) ?? tx : pool                           │ tx di context
  ▼                                                                │
DBTX (internal/database/postgres.go) ──────────────────────────────┘
  │  *pgxpool.Pool (tanpa tx)  /  pgx.Tx (dalam tx)
  ▼
PostgreSQL
```

### Aliran Data

1. HTTP request masuk → **handler** decode ke DTO + validasi (`pkg/validator`).
2. **Service** menjalankan business logic. Operasi yang butuh konsistensi dibungkus `WithTransaction`.
3. Service memanggil method pada **interface Repository** (port).
4. Implementasi repository (di `infrastructure/postgres`) membaca transaction dari context:
   - Ada transaksi (`database.GetTx(ctx)` non-nil) → semua query memakai transaksi itu.
   - Tidak ada → memakai koneksi pool biasa.
5. Response dibungkus `pkg/response` dan dikirim balik.

## Pola Dependensi

- **Service** → tergantung pada *interface* `Repository` (tidak kenal pgx/pool) dan `database.TxManager` (tidak memegang pool langsung).
- **Repository impl** → tergantung pada `internal/database` (`DBTX`, `GetTx`).
- **TxManager** → implementasi di `infrastructure/postgres`, interface di `internal/database`. Service cukup `s.tx.WithTransaction(ctx, fn)` — tanpa mendeklarasikan transaksi sendiri.
- **Wiring** semuanya dilakukan di `cmd/api/main.go` (manual dependency injection):

```go
pool := database.NewPostgres(cfg.Database)

repo := postgres.NewUserRepository(pool)   // implementasi port
txm  := postgres.NewTxManager(pool)        // transaction manager bersama
svc  := user.NewService(repo, txm)         // service diberi repo + TxManager
h    := user.NewHandler(svc)

server.NewHTTP(h, cfg.HTTPPort, log).Run()
```

Keuntungan:
- Service bisa di-*test* dengan repo mock (tanpa DB nyata).
- Implementasi repository bisa diganti (mis. dari pgx ke GORM) tanpa mengubah service.
- Transaction di-manage di satu tempat (`TxManager` di `infrastructure/postgres`), bukan per-service.

## Cara Menambah Modul Baru

Contoh menambah modul `product`:

1. **Domain** — `internal/module/product/domain.go` → entity `Product`.
2. **Port repository** — `internal/module/product/repository.go` → interface `Repository` (method yang dibutuhkan service).
3. **Implementasi** — `internal/infrastructure/postgres/product_repository.go` → struct `productRepository{pool}` + helper `db(ctx)` + query pgx.
4. **DTO** — `internal/module/product/dto.go` → request/response dengan tag `validate`.
5. **Service** — `internal/module/product/service.go` → `NewService(repo, txm)` (service menyimpan `database.TxManager`, bukan pool); bungkus operasi multi-write dengan `s.tx.WithTransaction(ctx, fn)`.
6. **Handler** — `internal/module/product/handler.go` → `Routes()` (chi), decode + validasi + panggil service + `pkg/response`.
7. **Wiring** — `cmd/api/main.go` + `internal/server/http.go` → construct & mount route.
8. **Migrasi** (jika ada tabel baru) — tambah `migrations/003_*.sql` (dijalankan otomatis oleh `RunMigrations`).
9. **Test** — `internal/module/product/service_test.go` mengikuti pola `service_test.go` user.

> Catatan: folder `internal/module/auth/`, `internal/module/transaction/`, `internal/infrastructure/kafka/`, `internal/infrastructure/redis/`, `internal/server/grpc.go` saat ini masih placeholder kosong — perlu diisi saat modul tersebut dikembangkan.