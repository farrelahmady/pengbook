# Issue #4: Logging System Improvement

## Date
2026-09-12

## Status
✅ Resolved

## Priority
High

## Description

Backend dan frontend memiliki pendekatan logging yang berbeda dan belum optimal:

### Backend Issues
1. **Hardcoded log level** — Logger hanya menggunakan `slog.LevelInfo`, tidak ada cara untuk mengubah level tanpa rebuild
2. **Tidak ada environment detection** — Tidak ada pembeda antara development dan production
3. **Tidak ada LOG_LEVEL env** — Tidak ada fleksibilitas untuk override level

### Frontend Issues
1. **Console.log directly** — Banyak komponen menggunakan `console.log` langsung
2. **Tidak ada structured logging** — Log tidak memiliki format yang konsisten
3. **Tidak ada level filtering** — Semua log selalu muncul

## Solution

### Backend: APP_ENV + LOG_LEVEL

Implementasi environment-aware logging dengan auto-detect:

**Files Changed:**
- `api/internal/config/config.go` — Tambah `AppEnv` field dan `GetLogLevel()` method
- `api/cmd/api/main.go` — Gunakan `cfg.GetLogLevel()` untuk inisialisasi logger
- `api/pkg/logger/logger.go` — Accept `logLevel` parameter

**Implementation:**
```go
// config.go
type Config struct {
    // ... existing fields
    AppEnv   string `env:"APP_ENV" envDefault:"development"`
    LogLevel string `env:"LOG_LEVEL"`
}

func (c *Config) GetLogLevel() string {
    if c.LogLevel != "" {
        return c.LogLevel // Manual override
    }
    
    if c.AppEnv == "production" {
        return "info"
    }
    return "debug" // Default untuk development
}
```

**Usage:**
```bash
# Development — otomatis debug
APP_ENV=development go run .

# Production — otomatis info
APP_ENV=production go run .

# Override manual
APP_ENV=production LOG_LEVEL=debug go run .
```

### Frontend: Option A + B (Custom Logger)

Implementasi custom logger dengan environment-based approach:

**Files Created:**
- `web/lib/logger/types.ts` — Type definitions
- `web/lib/logger/logger.ts` — Logger implementation
- `web/lib/logger/index.ts` — Factory & exports
- `web/lib/logger/transports/console.ts` — Console transport

**Files Updated:**
- `web/http/middlewares/logger-middleware.ts` — Gunakan structured logger
- `web/http/interceptors/auth-interceptor.ts` — Gunakan structured logger
- `web/services/journal.ts` — Add logging
- `web/services/auth.ts` — Add logging
- `web/lib/auth-context.tsx` — Add logging
- `web/components/journal/journal-scroll-view.tsx` — Add logging
- `web/app/[locale]/(auth)/jurnal/journal-topbar.tsx` — Add logging

**Usage:**
```typescript
import { createLogger } from "@/lib/logger";

const logger = createLogger("ModuleName");

logger.debug("Debug message", { key: "value" });
logger.info("Info message");
logger.warn("Warning message");
logger.error("Error message", { error });
```

**Configuration:**
```bash
# Default behavior
# Development: debug level
# Production: info level

# Manual override
NEXT_PUBLIC_LOG_LEVEL=debug npm run build
```

## Configuration Matrix

### Backend

| APP_ENV | LOG_LEVEL | Result |
|---------|-----------|--------|
| `development` | (empty) | `debug` (auto) |
| `production` | (empty) | `info` (auto) |
| `development` | `error` | `error` (override) |
| `production` | `debug` | `debug` (override) |

### Frontend

| NODE_ENV | NEXT_PUBLIC_LOG_LEVEL | Result |
|----------|----------------------|--------|
| `development` | (empty) | `debug` (auto) |
| `production` | (empty) | `info` (auto) |
| `development` | `error` | `error` (override) |
| `production` | `debug` | `debug` (override) |

## Testing

### Backend
```bash
# Test debug level
APP_ENV=development go run ./cmd/api
# Expected: All logs including debug

# Test info level
APP_ENV=production go run ./cmd/api
# Expected: Only info, warn, error logs

# Test override
APP_ENV=development LOG_LEVEL=error go run ./cmd/api
# Expected: Only error logs
```

### Frontend
```bash
# Test debug level
NEXT_PUBLIC_LOG_LEVEL=debug npm run dev
# Expected: All logs including debug

# Test info level
NEXT_PUBLIC_LOG_LEVEL=info npm run dev
# Expected: Only info, warn, error logs
```

## Impact

### Positive
- ✅ Konsisten logging antara frontend dan backend
- ✅ Environment-aware logging (development vs production)
- ✅ Debug production tanpa rebuild (dengan LOG_LEVEL override)
- ✅ Structured logging untuk observability
- ✅ Zero dependency untuk frontend logger

### Negative
- ⚠️ Backend butuh restart untuk ganti log level (wajar)
- ⚠️ Frontend perlu rebuild untuk ganti log level (karena NEXT_PUBLIC_ prefix)

## Related Documentation

- [Logging System Documentation](../LOGGING.md)
- [HTTP Client Architecture](../HTTP_INTERCEPTOR.md)

## References

- [Go slog package](https://pkg.go.dev/log/slog)
- [Next.js Environment Variables](https://nextjs.org/docs/basic-features/environment-variables)
- [12-Factor App: Config](https://12factor.net/config)
