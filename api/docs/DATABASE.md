# Database & Transaksi

Proyek memakai **PostgreSQL** dengan driver **`jackc/pgx/v5`** (pooling via `pgxpool`). Semua logika DB berada di `internal/database/` dan `internal/infrastructure/postgres/`.

## Konfigurasi (`.env`)

Nilai dibaca dari file `.env` (atau env OS) oleh `internal/config/config.go` memakai `github.com/caarlos0/env/v11`.

| Env | Field config | Default | Keterangan |
|---|---|---|---|
| `DB_HOST` | `Database.Host` | — | Host PostgreSQL |
| `DB_PORT` | `Database.Port` | — | Port PostgreSQL |
| `DB_USER` | `Database.User` | — | User database |
| `DB_PASSWORD` | `Database.Password` | — | Password database |
| `DB_NAME` | `Database.Name` | — | Nama database |
| `DB_SSLMODE` | `Database.SSLMode` | — | `disable` / `require` / dll |
| `HTTP_PORT` | `HTTPPort` | `8080` | Port HTTP server |

Contoh `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=pengbook
DB_SSLMODE=disable
HTTP_PORT=8080
```

Catatan implementasi:

- Struktur `Config` memakai tag `env:"DB"` pada field `Database` agar `caarlos0/env` menelusuri nested struct (`DB_HOST` → `Database.Host`).
- `godotenv.Load()` bersifat **opsional** (error diabaikan) → saat `go test` yang berjalan dari folder package, `.env` tetap ditemukan melalui `internal/database/testutil.go` (mencari `.env` naik ke parent directory).

## Koneksi Pool

`internal/database/postgres.go` → `NewPostgres(cfg)`:

- DSN dirakit dengan `url.UserPassword` sehingga password dengan karakter spesial aman.
- Pool: `MaxConns = 10`, `MinConns = 1`.
- Setelah `pgxpool.NewWithConfig`, dilakukan `Ping` untuk memastikan koneksi hidup (pool di-close bila gagal).

```go
pool, err := database.NewPostgres(cfg.Database)
defer pool.Close()
```

## Interface `DBTX`

Kunci agar satu set query bisa dipakai baik oleh koneksi biasa **maupun** transaksi:

```go
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
```

Yang memenuhi interface ini:

- `*pgxpool.Pool` — koneksi biasa (dipakai saat tidak ada transaksi).
- `pgx.Tx` — transaksi (dipakai saat ada transaksi).

## Transaction Manager

`internal/database/tx.go` — menyimpan transaksi ke `context` (adaptasi dari pola `sql.Tx` di context):

```go
func WithTx(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func GetTx(ctx context.Context) DBTX {
	tx, _ := ctx.Value(txKey{}).(DBTX)
	return tx // nil jika tidak ada transaksi
}
```

### Service: `TxManager` (shared, bukan per-service)

`WithTransaction` **tidak** dideklarasikan per-service. Ia adalah utilitas bersama:

- Interface `TxManager` didefinisikan **sekali** di `internal/database/tx.go`.
- Implementasinya di `internal/infrastructure/postgres/transaction.go` → `NewTxManager(pool)`.

Service cukup menyimpan `database.TxManager` (tidak lagi memegang `*pgxpool.Pool`), lalu memanggil `s.tx.WithTransaction(ctx, fn)`.

Implementasi (`internal/infrastructure/postgres/transaction.go`). **Poin penting:** jika sudah ada transaksi di context (dari pemanggil/test), service **tidak** membuka transaksi baru dan **tidak** commit — commit/rollback diserahkan ke pemanggil.

```go
func (m *txManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Sudah ada transaksi dari luar (mis. test)? Pakai itu, jangan begin/commit.
	if database.GetTx(ctx) != nil {
		return fn(ctx)
	}

	// Production: buka transaksi baru.
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // aman jika error/panic

	if err := fn(database.WithTx(ctx, tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
```

Wiring di `cmd/api/main.go`:

```go
repo := postgres.NewUserRepository(pool)
txm  := postgres.NewTxManager(pool)
svc  := user.NewService(repo, txm) // service menerima TxManager, bukan pool
```

Contoh penggunaan — `Create` menulis **2 baris** (user + audit log) dalam 1 transaksi:

```go
err := s.tx.WithTransaction(ctx, func(ctx context.Context) error {
	if err := s.repo.Create(ctx, u); err != nil {
		return err
	}
	return s.repo.InsertAuditLog(ctx, &AuditLog{UserID: u.ID, Action: "user.created"})
})
```

Jika salah satu gagal → seluruh perubahan di-rollback.

### Repository: `db(ctx)`

Implementasi repository (`internal/infrastructure/postgres/user_repository.go`) membaca dulu dari context — **ada transaksi pakai transaksi, tidak ada pakai koneksi biasa**:

```go
func (r *userRepository) db(ctx context.Context) database.DBTX {
	if tx := database.GetTx(ctx); tx != nil {
		return tx // transaksi
	}
	return r.pool // koneksi biasa
}
```

Semua query dipanggil lewat helper ini, contoh:

```go
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	return r.db(ctx).QueryRow(ctx, createUserQuery, u.Name, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}
```

## Migration

- File SQL berada di `migrations/` dan **di-embed** ke binary via `migrations/embed.go` (`//go:embed *.sql`).
- `internal/database/migration.go` → `RunMigrations(ctx, pool)` menjalankan semua `.sql` berurutan (lexicographic: `001_`, `002_`, ...).
- Dipanggil otomatis di `cmd/api/main.go` saat startup dan di `testutil` untuk test.

Menambah migration baru: cukup buat file `migrations/00X_*.sql` — dijalankan otomatis. Gunakan `CREATE TABLE IF NOT EXISTS` agar idempoten.

## Testing Transaksi

File: `internal/module/user/service_test.go` — butuh **PostgreSQL yang aktif** sesuai `.env`.

### Test A — Test menginisialisasi transaksi sendiri, service TIDAK commit

```go
tx, err := pool.Begin(ctx)         // 1. test buka transaksi sendiri
defer tx.Rollback(ctx)             // 4. (pasti) rollback di akhir

txCtx := database.WithTx(ctx, tx)  // 2. inject transaksi ke context

created, err := svc.Create(txCtx, req) // 3. service jalan, TIDAK commit
// ... verifikasi data tampak DI DALAM tx ...
// setelah Rollback → verifikasi via pool (koneksi lain):
//   count(users) == 0 dan count(user_audit_logs) == 0
```

Karena `WithTransaction` (via `TxManager`) mendeteksi tx sudah ada di context, service tidak pernah memanggil `Commit` — rollback test membuktikan data benar-benar tidak tersimpan.

### Test B — Service commit sendiri (tanpa inject tx)

```go
created, err := svc.Create(ctx, req) // service begin + commit sendiri
// verifikasi via pool: users & audit log tersimpan → cleanup DELETE
```

### Menjalankan test

```bash
go test ./internal/module/user/ -v -count=1
```

### Menyalakan PostgreSQL (Windows)

Jika service PostgreSQL berstatus `Stopped` dan tidak punya akses admin untuk `Start-Service`, jalankan:

```powershell
# via terminal admin
Start-Service postgresql-x64-17

# atau langsung tanpa service
& "C:\Program Files\PostgreSQL\17\bin\pg_ctl.exe" start -w -D "C:\Program Files\PostgreSQL\17\data"
```

Pastikan database `pengbook` sudah dibuat dan password user `postgres` sesuai `DB_PASSWORD` di `.env`:

```powershell
& "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -c "ALTER USER postgres PASSWORD 'password';"
& "C:\Program Files\PostgreSQL\17\bin\psql.exe" -U postgres -c "CREATE DATABASE pengbook;"
```