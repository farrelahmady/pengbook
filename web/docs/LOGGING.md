# Logging System Documentation

## Overview

Pengbook menerapkan sistem logging terpadu untuk **Frontend** (Next.js) dan **Backend** (Go) guna memudahkan debugging di development dan monitoring di production.

## Table of Contents

- [Architecture](#architecture)
- [Frontend Logging](#frontend-logging)
- [Backend Logging](#backend-logging)
- [Configuration](#configuration)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    LOGGING ARCHITECTURE                       │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  FRONTEND (Next.js)                                          │
│  ├── lib/logger/                                             │
│  │   ├── types.ts          — Type definitions                │
│  │   ├── logger.ts         — Logger implementation           │
│  │   ├── index.ts          — Factory & exports               │
│  │   └── transports/       — Output transports               │
│  │       └── console.ts    — Browser console output          │
│  ├── Config: NEXT_PUBLIC_LOG_LEVEL                           │
│  └── Levels: debug → info → warn → error                     │
│                                                               │
│  BACKEND (Go)                                                │
│  ├── pkg/logger/                                             │
│  │   ├── logger.go         — Logger factory                  │
│  │   └── context.go        — Context-based logger            │
│  ├── Config: APP_ENV + LOG_LEVEL                             │
│  └── Levels: Debug → Info → Warn → Error                     │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Frontend Logging

### Overview

Frontend menggunakan custom logger dengan **Option A + B** approach:
- **Option A**: Environment-based (NODE_ENV check)
- **Option B**: Structured logging with levels

### Installation

Tidak ada dependency tambahan. Logger sudah termasuk di `lib/logger/`.

### Usage

```typescript
import { createLogger } from "@/lib/logger";

const logger = createLogger("ModuleName");

// Debug — operasi normal, debugging
logger.debug("Fetching journals", { limit: 10 });

// Info — operasi berhasil
logger.info("Journals fetched", { count: data.length });

// Warn — situasi yang perlu perhatian
logger.warn("Token expiring soon", { expiresIn: 300 });

// Error — kegagalan operasi
logger.error("Failed to fetch journals", { error });
```

### Log Levels

| Level | Kapan Digunakan | Contoh |
|-------|-----------------|--------|
| `debug` | Operasi normal, debugging | Fetching data, form submission |
| `info` | Operasi berhasil | Data fetched, user logged in |
| `warn` | Situasi perlu perhatian | Token expiring, fallback used |
| `error` | Kegagalan operasi | API call failed, validation error |

### Components That MUST Use Logger

- **Services** (`services/*.ts`) — log semua API calls dan responses
- **HTTP Interceptors** (`http/interceptors/*.ts`) — log auth flow, retries
- **HTTP Middlewares** (`http/middlewares/*.ts`) — log requests
- **Auth Context** (`lib/auth-context.tsx`) — log init, login, logout, refresh
- **Page Components** — log query states, error handling
- **Form Components** — log submissions, validation errors

### Output Format

**Development:**
```
🔍 [DEBUG] 2026-09-12 10:30:00.000 Fetching journals { module: "journal", limit: 10 }
ℹ️ [INFO]  2026-09-12 10:30:01.000 Journals fetched { module: "journal", count: 25 }
⚠️ [WARN]  2026-09-12 10:35:00.000 Token expiring soon { module: "Auth", expiresIn: 300 }
❌ [ERROR] 2026-09-12 10:35:05.000 Failed to fetch { module: "journal", error: {...} }
```

---

## Backend Logging

### Overview

Backend menggunakan Go's standard `log/slog` package dengan JSON output dan context-based logger injection.

### Usage

```go
import "pengbook/api/pkg/logger"

// Di handler — ambil logger dari context
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

// Di service
func (s *Service) Create(ctx context.Context, userID int64, req CreateRequest) (*Journal, error) {
    log := logger.FromContext(ctx)
    
    log.Debug("creating journal entry", "user_id", userID)
    
    // ... business logic
    
    log.Info("journal entry created", "journal_id", journal.ID)
    return journal, nil
}
```

### Log Levels

| Level | Kapan Digunakan | Contoh |
|-------|-----------------|--------|
| `Debug` | Operasi detail, debugging | Query execution, internal state |
| `Info` | Operasi berhasil | Request handled, data fetched |
| `Warn` | Situasi perlu perhatian | Fallback used, retry attempt |
| `Error` | Kegagalan operasi | DB error, validation failed |

### Components That MUST Use Logger

- **Handlers** (`internal/module/*/handler.go`) — log all HTTP requests and responses
- **Services** (`internal/module/*/service.go`) — log business operations
- **Repositories** (`internal/module/*/repository.go`) — log database queries (optional)
- **Middleware** (`internal/middleware/*.go`) — log middleware operations
- **Infrastructure** (`internal/infrastructure/*.go`) — log external service calls

### Output Format

```json
{
  "time": "2026-09-12T10:30:00Z",
  "level": "INFO",
  "msg": "getting journals",
  "request_id": "abc-123",
  "user_id": 123
}
```

---

## Configuration

### Frontend

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_LOG_LEVEL` | `debug` (dev) / `info` (prod) | Minimum log level |

**Development (default):**
```bash
# Otomatis debug level
npm run dev
```

**Production:**
```bash
# Otomatis info level
npm run build
```

**Manual Override:**
```bash
# Force debug di production
NEXT_PUBLIC_LOG_LEVEL=debug npm run build

# Force error di development
NEXT_PUBLIC_LOG_LEVEL=error npm run dev
```

### Backend

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Application environment |
| `LOG_LEVEL` | Auto-detect | Manual log level override |

**Auto-detect Logic:**
```
┌─────────────────────────────────────────────────┐
│  LOG_LEVEL diset?                               │
│  ├── Ya → Gunakan LOG_LEVEL (manual override)   │
│  └── Tidak → Auto-detect dari APP_ENV:          │
│      ├── production → "info"                     │
│      └── development/staging → "debug"          │
└─────────────────────────────────────────────────┘
```

**Examples:**
```bash
# Development — otomatis debug
APP_ENV=development go run .

# Production — otomatis info
APP_ENV=production go run .

# Override manual
APP_ENV=development LOG_LEVEL=error go run .
APP_ENV=production LOG_LEVEL=debug go run .
```

**Configuration Matrix:**

| APP_ENV | LOG_LEVEL | Hasil |
|---------|-----------|-------|
| `development` | (kosong) | `debug` (otomatis) |
| `production` | (kosong) | `info` (otomatis) |
| `development` | `error` | `error` (manual override) |
| `production` | `debug` | `debug` (manual override) |

---

## Best Practices

### Frontend

1. **Module Naming** — Gunakan nama yang deskriptif
   ```typescript
   // ✅ Good
   const logger = createLogger("journal.service");
   const logger = createLogger("AuthInterceptor");
   
   // ❌ Bad
   const logger = createLogger("logger");
   const logger = createLogger("test");
   ```

2. **Context Enrichment** — Sertakan data yang relevan
   ```typescript
   // ✅ Good
   logger.info("Journals fetched", { count: data.length, userId });
   
   // ❌ Bad
   logger.info("Journals fetched");
   ```

3. **Error Logging** — Sertakan error object
   ```typescript
   // ✅ Good
   logger.error("Failed to fetch", { error: error.message, stack: error.stack });
   
   // ❌ Bad
   logger.error("Failed to fetch");
   ```

### Backend

1. **Always Use Context** — Ambil logger dari context
   ```go
   // ✅ Good
   log := logger.FromContext(r.Context())
   
   // ❌ Bad
   log := slog.Default()
   ```

2. **Structured Logging** — Gunakan key-value pairs
   ```go
   // ✅ Good
   log.Info("user created", "user_id", user.ID, "email", user.Email)
   
   // ❌ Bad
   log.Info("user created")
   ```

3. **Error Context** — Sertakan error details
   ```go
   // ✅ Good
   log.Error("failed to create user", "error", err, "user_id", userID)
   
   // ❌ Bad
   log.Error("failed to create user")
   ```

---

## Troubleshooting

### Frontend

**Masalah: Logs tidak muncul di production**
- Solusi: Set `NEXT_PUBLIC_LOG_LEVEL=debug` (perlu rebuild)

**Masalah: Terlalu banyak logs**
- Solusi: Set `NEXT_PUBLIC_LOG_LEVEL=warn` atau `error`

### Backend

**Masalah: Debug logs tidak muncul**
- Solusi: Pastikan `APP_ENV=development` atau set `LOG_LEVEL=debug`

**Masalah: Terlalu banyak logs di production**
- Solusi: Pastikan `APP_ENV=production` (otomatis `info` level)

**Masalah: Ingin debug production tanpa restart**
- Solusi: Gunakan `LOG_LEVEL=debug` dan restart server

---

## Related Documentation

- [HTTP Client Architecture](./HTTP_INTERCEPTOR.md)
- [Architecture Issues](./issues/ARCHITECTURE_ISSUES_2026-09-11.md)
