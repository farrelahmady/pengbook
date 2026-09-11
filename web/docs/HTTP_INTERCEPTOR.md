# HTTP Interceptor Pattern

## Overview

Pengbook menggunakan **Response Interceptor Pattern** untuk menangani HTTP error responses (terutama 401 Unauthorized) secara otomatis. Pattern ini terinspirasi dari Axios interceptors dan merupakan industry best practice untuk authentication handling.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Request Flow                                                   │
│                                                                 │
│  Component → Service → HttpClient.request()                     │
│                          ↓                                      │
│                    Request Middlewares                           │
│                    (auth middleware: add token)                  │
│                          ↓                                      │
│                    fetch() → API                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│  Response Flow                                                  │
│                                                                 │
│  fetch() returns response                                       │
│                          ↓                                      │
│                    Response Interceptors                         │
│                    (auth interceptor: handle 401)               │
│                          ↓                                      │
│                    Return response to caller                    │
└─────────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. Interceptor Interface

```typescript
// http/core/http-client.ts

export interface Interceptor {
  onRequest?: (config: HttpRequestConfig) => Promise<HttpRequestConfig>;
  onResponse?: (
    response: Response,
    config: HttpRequestConfig,
  ) => Promise<Response | null>;
}
```

- `onRequest`: Modify request before sending
- `onResponse`: Handle response, can return `null` to signal retry

### 2. Auth Interceptor

```typescript
// http/interceptors/auth-interceptor.ts

export function authInterceptor(options?: AuthInterceptorOptions) {
  // Returns interceptor that:
  // 1. Catches 401 responses
  // 2. Refreshes token (using refresh lock for concurrency)
  // 3. Retries request with new token
  // 4. Calls onRefreshFailed if refresh fails
}
```

### 3. Refresh Lock Pattern

Handles **concurrent 401 responses** — only one refresh happens at a time:

```typescript
interface RefreshLockState {
  isRefreshing: boolean;
  refreshPromise: Promise<string | null> | null;
  waitQueue: Array<{ resolve: (token: string | null) => void }>;
}
```

### 4. Global Refresh Failed Handler

```typescript
// lib/http-client.ts

let _onRefreshFailed: (() => Promise<void>) | null = null;

export function setOnRefreshFailed(handler: () => Promise<void>) {
  _onRefreshFailed = handler;
}
```

- Set once in `AuthProvider`
- Called when token refresh fails
- Typically used to redirect to login

## How It Works

### Single 401 Request

```
Request → 401 Response
    ↓
Auth Interceptor catches 401
    ↓
Check: isRefreshing? → NO
    ↓
Start refresh: getToken()
    ↓
Token refreshed successfully
    ↓
Retry request with new token
    ↓
Return response
```

### Multiple Concurrent 401 Requests

```
Request A → 401 ─┐
                  │
Request B → 401 ─┼──→ Check: isRefreshing?
                  │         │
Request C → 401 ─┘    ┌────┴────┐
                       │         │
                     NO       YES
                       │         │
                       ↓         ↓
                 Start      Wait for
                 refresh    existing promise
                       │         │
                       ↓         │
                 refreshPromise ─┘
                       │
                       ↓
                 Token refreshed
                       │
              ┌────────┼────────┐
              ↓        ↓        ↓
           Retry A  Retry B  Retry C
```

## File Structure

```
http/
├── core/
│   └── http-client.ts          # HttpClient with interceptor support
├── interceptors/
│   └── auth-interceptor.ts     # Auth interceptor with refresh lock
├── middlewares/
│   ├── auth-middleware.ts       # Request auth middleware
│   └── logger-middleware.ts    # Request logger
└── types/
    └── http.ts                 # TypeScript interfaces

lib/
├── http-client.ts              # Client factories + setOnRefreshFailed
├── token-provider.ts           # Global token provider (setTokenProvider)
├── auth-context.tsx            # Auth state (registers both providers)
└── auth.ts                     # Auth service methods
```

### Registration Flow

```
AuthProvider mount
    ↓
setTokenProvider(getToken)      → token-provider.ts
    ↓
setOnRefreshFailed(handler)     → http-client.ts
    ↓
Services can now use authHttpClient()
```

## Usage

### Basic Usage (Service Layer)

```typescript
// services/journal.ts
import { authHttpClient } from "@/lib/http-client";

const client = authHttpClient(); // ← 401 handling automatic

export const journalService = {
  getAll: async () => {
    const res = await client.get<Journal[]>("/api/journals");
    return res.data;
  },
};
```

No need to handle 401 in services — interceptor handles it.

### With Refresh Failure Handler

The `onRefreshFailed` handler is set globally via `setOnRefreshFailed()` in `auth-context.tsx`:

```typescript
// lib/auth-context.tsx (set once in AuthProvider)
import { setOnRefreshFailed } from "@/lib/http-client";

export function AuthProvider({ children }) {
  const router = useRouter();

  useEffect(() => {
    setOnRefreshFailed(async () => {
      clearTokens();
      router.push("/login");
    });
  }, [router]);

  // ...
}
```

This handler is called when:
- Token refresh fails (refresh token expired)
- API returns 401 but token is not expiring soon (invalid token)

### Creating Custom Interceptor

```typescript
// http/interceptors/logging-interceptor.ts
export function loggingInterceptor() {
  return {
    onResponse: async (response: Response, config: HttpRequestConfig) => {
      console.log(`[${config.method}] ${config.url} → ${response.status}`);
      return response;
    },
  };
}

// Usage
const client = new HttpClient(baseUrl, {
  middlewares: [authMiddleware(getToken)],
  interceptors: [loggingInterceptor()],
});
```

## Concurrent Request Handling

### Problem

Ketika user session expired dan ada multiple concurrent requests:

```
Page Component A → fetch(/api/journals)      → 401
Page Component B → fetch(/api/summary)       → 401
Page Component C → fetch(/api/reports)       → 401

Without handling: 3 refresh attempts, race conditions
```

### Solution: Refresh Lock

```typescript
// auth-interceptor.ts

const lock = {
  isRefreshing: false,
  refreshPromise: null,
  waitQueue: [],
};

// Request A: isRefreshing = false → start refresh
// Request B: isRefreshing = true → wait for existing refresh
// Request C: isRefreshing = true → wait for existing refresh

// After refresh completes:
// All requests get the same new token and retry
```

### Implementation Details

```typescript
// 1. First request starts refresh
lock.isRefreshing = true;
lock.refreshPromise = getToken();

// 2. Other requests wait
if (lock.isRefreshing) {
  const token = await waitForRefresh(lock); // Returns promise
}

// 3. After refresh, notify all waiters
notifyWaiters(lock, newToken); // All waiting requests proceed

// 4. Reset lock
lock.isRefreshing = false;
lock.refreshPromise = null;
```

## Sequence Diagram

```
┌──────┐  ┌──────┐  ┌──────┐  ┌───────────┐  ┌──────┐
│Comp A│  │Comp B│  │Comp C│  │Interceptor│  │ API  │
└──┬───┘  └──┬───┘  └──┬───┘  └─────┬─────┘  └──┬───┘
   │         │         │            │            │
   │──GET────┼─────────┼───────────►│            │
   │         │         │            │──GET──────►│
   │         │         │            │            │
   │         │         │            │◄──401──────│
   │         │         │            │            │
   │         │         │            │[Start Refresh]
   │         │         │            │            │
   │         │──GET────┼───────────►│            │
   │         │         │            │──GET──────►│
   │         │         │            │            │
   │         │         │            │◄──401──────│
   │         │         │            │            │
   │         │         │            │[Wait for   │
   │         │         │            │ refresh]   │
   │         │         │            │            │
   │         │──GET────┼───────────►│            │
   │         │         │            │──GET──────►│
   │         │         │            │            │
   │         │         │            │◄──401──────│
   │         │         │            │            │
   │         │         │            │[Wait for   │
   │         │         │            │ refresh]   │
   │         │         │            │            │
   │         │         │            │[Refresh    │
   │         │         │            │ Complete]  │
   │         │         │            │            │
   │         │         │            │──GET──────►│ (new token)
   │         │         │            │            │
   │         │         │            │◄──200──────│
   │         │         │            │            │
   │         │         │◄──notify───│            │
   │         │◄──notify─┼───────────│            │
   │◄──notify┼─────────┼───────────│            │
   │         │         │            │            │
   │         │         │◄──200──────│            │
   │         │◄──200────┼───────────│            │
   │◄──200───┼─────────┼───────────│            │
```

## Error Handling

### 401 Response

```
Response 401
    ↓
Interceptor catches
    ↓
Refresh token
    ├─ Success → Retry request
    └─ Failure → onRefreshFailed() → redirect to login
```

### Refresh Token Expired

```
Refresh token expired
    ↓
authService.refresh() throws
    ↓
clearTokens() + router.push("/login")
    ↓
User redirected to login page
```

### Network Error

```
Network error
    ↓
Exception thrown
    ↓
React Query error handling / component catch block
```

## Testing

### Unit Test for Interceptor

```typescript
describe("Auth Interceptor", () => {
  it("should refresh only once for concurrent 401s", async () => {
    const getToken = vi.fn().mockResolvedValue("new-token");
    setTokenProvider(getToken);

    const interceptor = authInterceptor();

    const config = { url: "/api/test" };
    const response401 = new Response(null, { status: 401 });

    // Simulate 3 concurrent 401s
    const [r1, r2, r3] = await Promise.all([
      interceptor.onResponse(response401, { ...config }),
      interceptor.onResponse(response401, { ...config }),
      interceptor.onResponse(response401, { ...config }),
    ]);

    // Should only refresh once
    expect(getToken).toHaveBeenCalledTimes(1);

    // All should signal retry
    expect(r1).toBeNull();
    expect(r2).toBeNull();
    expect(r3).toBeNull();
  });
});
```

## Best Practices

1. **Single Responsibility**: Each interceptor handles one concern (auth, logging, etc.)

2. **Idempotent Retries**: Ensure requests can be safely retried

3. **Race Condition Handling**: Use refresh lock for concurrent 401s

4. **Prevent Infinite Loops**: Use `__retried` flag to prevent retry loops

5. **Separation of Concerns**:
   - `auth-middleware.ts`: Add token to request
   - `auth-interceptor.ts`: Handle 401 response
   - `token-provider.ts`: Token management

## References

- [Axios Interceptors](https://axios-http.com/docs/interceptors)
- [Auth0 SDK](https://auth0.com/docs/libraries/auth0js)
- [MSAL.js](https://learn.microsoft.com/en-us/azure/active-directory/develop/msal-js-initializing-client-applications)
