# Frontend Architecture Issues

**Tanggal**: 11 September 2026  
**Status**: Open  
**Scope**: Web Frontend (Next.js)

---

## Daftar Isu

| # | Issue | Prioritas | Effort | Status |
|---|-------|-----------|--------|--------|
| 1 | Timing bug pada `onRefreshFailed` capture | 🔴 Critical | Small | Open |
| 2 | `authService.me()` buat HttpClient baru setiap panggilan | 🟠 High | Small | Open |
| 3 | Dual token storage (cookie + header) | 🟠 High | Medium | Open |
| 4 | Logger middleware jalan di production | 🟠 High | Small | Open |
| 5 | React Query keys tidak konsisten | 🟡 Medium | Medium | Open |
| 6 | Dummy data services di production code | 🟡 Medium | Small | Open |
| 7 | Services tidak bisa di-test (tight coupling) | 🟢 Low | Large | Open |
| 8 | Retry middleware unused (dead code) | 🟢 Low | Small | Open |
| 9 | `tryRefresh()` stale closure risk | 🟢 Low | Small | ✅ Fixed |

---

## 🔴 Critical Issues

### 1. Timing Bug pada `onRefreshFailed` Capture

**Lokasi**: `lib/http-client.ts` — `authHttpClient()`

**Masalah**:

```typescript
let _authClient: HttpClient | null = null;

export function authHttpClient() {
  if (!_authClient) {
    // ...
    interceptors.push(
      authInterceptor({
        onRefreshFailed: _onRefreshFailed ?? undefined, // ← di-capture SAAT INI
      }),
    );
    _authClient = new HttpClient(baseUrl, { middlewares, interceptors });
  }
  return _authClient;
}
```

`_onRefreshFailed` di-capture saat pertama kali `authHttpClient()` dipanggil. Jika dipanggil sebelum `AuthProvider` mount, handler akan `undefined` permanen.

**Timeline Bug**:

```
1. Module load → authHttpClient() dipanggil (misal dari import)
   └→ _onRefreshFailed = null (belum di-set)
   └→ interceptor dibuat TANPA onRefreshFailed

2. AuthProvider mount
   └→ setOnRefreshFailed(handler) dipanggil
   └→ _onRefreshFailed = handler

3. Tapi singleton sudah terlanjur dibuat!
   └→ Interceptor TIDAK punya onRefreshFailed
   └→ Redirect on refresh failure TIDAK BERFUNGSI
```

**Solusi**:

Buat interceptor membaca `_onRefreshFailed` secara **lazy** (setiap request), bukan saat construction:

```typescript
// auth-interceptor.ts
return {
  onResponse: async (response, config) => {
    if (response.status === 401) {
      // Baca handler dari global setiap kali, bukan capture saat construction
      const handler = getOnRefreshFailed(); // ← lazy read
      // ...
      if (!token) {
        await handler?.();
      }
    }
  },
};
```

**Dampak**: User tidak ter-redirect ke login saat session expired. Security risk.

---

## 🟠 High Priority Issues

### 2. `authService.me()` Buat HttpClient Baru Setiap Panggilan

**Lokasi**: `lib/auth-context.tsx` — `fetchUser()`

**Masalah**:

```typescript
const fetchUser = useCallback(async (token: string): Promise<User | null> => {
  try {
    return await authService.me(async () => token); // ← createHttpClient baru!
  } catch {
    return null;
  }
}, []);
```

`authService.me()` memanggil `createHttpClient(getToken)` yang membuat `HttpClient` baru dengan middleware baru setiap kali.

**Dampak**:
- Waste memory dan CPU
- Middleware closures tidak di-cache
- Inconsistent dengan service lain yang pakai singleton

**Solusi**:

```typescript
// services/auth.ts
me: async (): Promise<User> => {
  const client = authHttpClient(); // ← pakai singleton
  const res = await client.get<ApiResponse<User>>("/api/v1/auth/me");
  return res.data.data;
},
```

---

### 3. Dual Token Storage (Cookie + Header)

**Lokasi**: 
- `lib/auth-context.tsx` — cookie storage
- `http/middlewares/auth-middleware.ts` — Authorization header

**Masalah**:

```
Token tersedia di DUA tempat:
├── Cookie: access_token, refresh_token (bisa diakses JavaScript)
└── Header: Authorization: Bearer <token> (dikirim di setiap request)

Redundant: browser otomatis kirim cookie, DAN header dikirim manual
XSS vulnerability: cookie bisa diakses via document.cookie
```

**Solusi**:

Pilih **SATU** strategi:

| Strategi | Kelebihan | Kekurangan |
|----------|-----------|------------|
| **HttpOnly Cookie** | Lebih aman, XSS-safe | Butuh backend support, CSRF protection needed |
| **Bearer Token Only** | Simpel, tidak butuh cookie | Token harus di-refresh manual, tidak persist antar tab |

**Rekomendasi**: Jika backend support, gunakan HttpOnly cookie. Jika tidak, gunakan Bearer token dengan in-memory storage (bukan cookie).

---

### 4. Logger Middleware Jalan di Production

**Lokasi**: `http/middlewares/logger-middleware.ts`

**Masalah**:

```typescript
export function loggerMiddleware() {
  return async (config: HttpRequestConfig) => {
    console.log(`[HTTP] ${config.method} ${config.url}`); // ← selalu jalan!
    return config;
  };
}
```

**Dampak**:
- Leak request URLs ke browser console
- Performance overhead di production
- Security risk (URLs bisa mengandung sensitive data)

**Solusi**:

```typescript
export function loggerMiddleware() {
  return async (config: HttpRequestConfig) => {
    if (process.env.NODE_ENV === "development") {
      console.log(`[HTTP] ${config.method} ${config.url}`);
    }
    return config;
  };
}
```

---

## 🟡 Medium Priority Issues

### 5. React Query Keys Tidak Konsisten

**Lokasi**: Berbagai komponen

**Masalah**:

```typescript
// Structured (bagus)
["journals", "scroll-view", { startDate, endDate, accountIds }]

// Flat string (kurang bagus)
["journalSummary"]
["coaSummary"]
["trialBalance"]
```

React Query lakukan **prefix matching** untuk invalidation. Flat string membuat invalidation unpredictable.

**Contoh Masalah**:

```typescript
// basic-form.tsx
queryClient.invalidateQueries({ queryKey: ["journals"] });

// Ini akan invalidate ["journals", "scroll-view", ...]
// Tapi TIDAK akan invalidate ["journalSummary"]
```

**Solusi**:

Standardisasi dengan namespace pattern:

```typescript
// Query keys
["journals", "list", { filters }]
["journals", "summary"]
["coa", "summary"]
["reports", "trial-balance"]

// Invalidations
queryClient.invalidateQueries({ queryKey: ["journals"] }); // invalidate ALL journal queries
queryClient.invalidateQueries({ queryKey: ["journals", "summary"] }); // invalidate specific
```

---

### 6. Dummy Data Services di Production Code

**Lokasi**: `services/coa.ts`, `services/asset.ts`, `services/report.ts`

**Masalah**:

```typescript
// services/coa.ts
import { dummyFullCoa } from "@/lib/dummy-data";

export const coaService = {
  getSummary: async (): Promise<CoaSummary> => {
    await new Promise((resolve) => setTimeout(resolve, 1000)); // fake delay
    return Promise.resolve(dummyFullCoa); // ← hardcoded data
  },
};
```

**Dampak**:
- Tidak ada indikasi ini placeholder
- Bisa ter-deploy ke production
- Tests tidak meaningful

**Solusi**:

```typescript
// services/coa.ts
/**
 * @todo Replace with real API implementation
 * @see https://github.com/yourrepo/issues/123
 */
export const coaService = {
  getSummary: async (): Promise<CoaSummary> => {
    // TODO: Replace with actual API call
    // const res = await authHttpClient().get<ApiResponse<CoaSummary>>("/api/v1/coa/summary");
    // return res.data.data;
    
    const { dummyFullCoa } = await import("@/lib/dummy-data");
    return dummyFullCoa;
  },
};
```

Atau lebih baik: gunakan environment variable untuk toggle:

```typescript
const USE_MOCK = process.env.NEXT_PUBLIC_USE_MOCK === "true";

export const coaService = {
  getSummary: async () => {
    if (USE_MOCK) {
      return dummyFullCoa;
    }
    return authHttpClient().get("/api/v1/coa/summary");
  },
};
```

---

## 🟢 Low Priority Issues

### 7. Services Tidak Bisa Di-Test (Tight Coupling)

**Lokasi**: Semua services

**Masalah**:

```typescript
// services/journal.ts
import { authHttpClient } from "@/lib/http-client";

export const journalService = {
  getAll: async () => {
    const client = authHttpClient(); // ← tight coupling ke singleton
    // ...
  },
};
```

Unit testing must mock module-level import. Tidak bisa inject dependency.

**Solusi** (Dependency Injection):

```typescript
// services/journal.ts
import { authHttpClient } from "@/lib/http-client";
import type { HttpClient } from "@/http/core/http-client";

export function createJournalService(client?: HttpClient) {
  const httpClient = client ?? authHttpClient();
  
  return {
    getAll: async () => {
      const res = await httpClient.get("/api/v1/journals");
      return res.data;
    },
  };
}

// Default instance
export const journalService = createJournalService();

// Testing
const mockClient = { get: vi.fn() };
const testService = createJournalService(mockClient);
```

**Trade-off**: Lebih verbose, tapi lebih testable.

---

### 8. Retry Middleware Unused (Dead Code)

**Lokasi**: `http/middlewares/retry-middleware.ts`

**Masalah**: File ada tapi tidak pernah di-import atau digunakan.

**Solusi**: Hapus file atau gunakan jika diperlukan.

---

### 9. `tryRefresh()` Stale Closure Risk ✅ Fixed

**Lokasi**: `lib/auth-context.tsx` — `tryRefresh()`

**Masalah**:

```typescript
const tryRefresh = useCallback(async (): Promise<boolean> => {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    router.push("/login"); // ← router tidak di-dependency array
    return false;
  }
  // ...
}, []); // ← empty dependency array
```

`router` tidak ada di dependency array. Works karena `router` stable di Next.js App Router, tapi technically incorrect.

**Solusi**:

```typescript
const tryRefresh = useCallback(async (): Promise<boolean> => {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    router.push("/login");
    return false;
  }
  // ...
}, [router]); // ← add router to deps
```

**Status**: Fixed 2026-09-12

---

## Recommended Actions

### Phase 1: Critical Fixes (Minggu Ini)

1. **Fix `onRefreshFailed` capture bug** — Ubah ke lazy read di interceptor
2. **Fix `authService.me()` singleton** — Gunakan `authHttpClient()`
3. **Gate logger behind dev mode** — Tambah `NODE_ENV` check

### Phase 2: Security Improvements (Minggu Depan)

4. **Evaluate dual token storage** — Pilih cookie atau Bearer, bukan keduanya
5. **Tambah `Secure` flag ke cookies** — Minimal hardening

### Phase 3: Code Quality (Sprint Berikutnya)

6. **Standardize React Query keys** — Buat pattern yang konsisten
7. **Mark/remove dummy services** — Tambah TODO atau environment toggle
8. **Delete retry-middleware.ts** — Cleanup dead code

### Phase 4: Architecture Improvements (Backlog)

9. **Consider dependency injection** — Jika testing coverage jadi prioritas

---

## References

- [Axios Interceptors Pattern](https://axios-http.com/docs/interceptors)
- [React Query Keys Best Practices](https://tanstack.com/query/v4/docs/react/guides/query-keys)
- [OWASP Cookie Security](https://cheatsheetseries.owasp.org/cheatsheets/Cheat_Sheet.html)
- [Next.js Authentication](https://nextjs.org/docs/authentication)
