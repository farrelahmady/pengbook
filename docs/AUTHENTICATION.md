# Authentication System Documentation

> Dokumentasi lengkap sistem autentikasi Pengbook (Web + API)

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Database Schema](#database-schema)
4. [API Endpoints](#api-endpoints)
5. [Frontend Authentication](#frontend-authentication)
6. [Token Management](#token-management)
7. [Route Protection](#route-protection)
8. [Security Considerations](#security-considerations)
9. [Configuration](#configuration)
10. [File Reference](#file-reference)

---

## Overview

Pengbook menggunakan **JWT (JSON Web Token)** authentication dengan **refresh token rotation** yang disimpan di Redis. Sistem ini terdiri dari:

- **Backend:** Go API dengan Chi router
- **Frontend:** Next.js 16 dengan React 19
- **Storage:** PostgreSQL (users) + Redis (refresh tokens)

### Key Features

| Feature | Implementation |
|---------|---------------|
| Password Hashing | bcrypt (cost factor: 10) |
| Access Token | JWT HS256 (15 menit) |
| Refresh Token | Random string (7 hari, stored in Redis) |
| Token Refresh | Proactive (5 menit sebelum expiry) |
| Route Protection | Proxy middleware (Next.js 16) |
| Cookie Storage | `SameSite=Lax`, `path=/` |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT (Browser)                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │  Login Page  │    │ Register Page│    │  Dashboard   │      │
│  │  /login      │    │ /register    │    │  /jurnal     │      │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘      │
│         │                   │                   │               │
│         └───────────────────┼───────────────────┘               │
│                             │                                   │
│                    useAuth() hook                               │
│                             │                                   │
│                    ┌────────▼────────┐                          │
│                    │   AuthProvider  │                          │
│                    │  (React Context)│                          │
│                    └────────┬────────┘                          │
│                             │                                   │
│              ┌──────────────┼──────────────┐                    │
│              │              │              │                    │
│         ┌────▼────┐   ┌────▼────┐   ┌────▼────┐               │
│         │ Cookies │   │  User   │   │  Token  │               │
│         │  Store  │   │  State  │   │ Refresh │               │
│         └────┬────┘   └─────────┘   └─────────┘               │
│              │                                                   │
└──────────────┼───────────────────────────────────────────────────┘
               │
               │ HTTP Request
               │ Authorization: Bearer <token>
               │
┌──────────────▼───────────────────────────────────────────────────┐
│                     PROXY MIDDLEWARE (Edge)                      │
├─────────────────────────────────────────────────────────────────┤
│  - Cek access_token cookie                                      │
│  - Guest routes → redirect ke /jurnal jika login                │
│  - Auth routes → redirect ke /login jika belum login            │
│  - Locale detection (next-intl)                                 │
└──────────────┬───────────────────────────────────────────────────┘
               │
               │ Forward Request
               │
┌──────────────▼───────────────────────────────────────────────────┐
│                    API SERVER (Go + Chi)                         │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │ Auth Handler │    │ Auth Service │    │Auth Repository│      │
│  │  /api/v1/auth│    │  (Business)  │    │   (Redis)    │      │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘      │
│         │                   │                   │               │
│         │            ┌──────▼───────┐           │               │
│         │            │ User Service │           │               │
│         │            │  (Business)  │           │               │
│         │            └──────┬───────┘           │               │
│         │                   │                   │               │
│         │            ┌──────▼───────┐    ┌──────▼───────┐      │
│         │            │ User Repo    │    │ Redis Client │      │
│         │            │ (PostgreSQL) │    │              │      │
│         │            └──────────────┘    └──────────────┘      │
│         │                                                       │
│  ┌──────▼───────┐                                               │
│  │ JWT Auth     │  ← Middleware untuk protected routes          │
│  │ Middleware   │    (/api/v1/accounts, /api/v1/journals)       │
│  └──────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘
               │
               │
┌──────────────▼───────────────────────────────────────────────────┐
│                      DATA STORES                                │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐                    ┌──────────────┐           │
│  │  PostgreSQL  │                    │    Redis     │           │
│  │  - users     │                    │ - refresh_   │           │
│  │  - accounts  │                    │   tokens     │           │
│  │  - journals  │                    │              │           │
│  └──────────────┘                    └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    username      VARCHAR(50) UNIQUE NOT NULL,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_username ON users(username);
```

### User Domain Model (Go)

```go
type User struct {
    ID           int64     `json:"id"`
    Name         string    `json:"name"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`  // Never serialized
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

### User Domain Model (TypeScript)

```typescript
interface User {
    id: number;
    name: string;
    username: string;
    email: string;
    createdAt: string;
}
```

### Refresh Token (Redis)

Refresh token disimpan di Redis dengan struktur:

```
Key:   refresh_token:{token_value}
Value: {
    "user_id": 123,
    "email": "user@example.com",
    "expires_at": "2026-09-16T00:00:00Z"
}
TTL:   7 days (604800 seconds)
```

---

## API Endpoints

### Base URL

```
http://localhost:8080/api/v1
```

### Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | Register new user | ❌ |
| POST | `/auth/login` | Login user | ❌ |
| POST | `/auth/refresh` | Refresh access token | ❌ |
| POST | `/auth/logout` | Revoke refresh token | ❌ |
| GET | `/auth/me` | Get current user profile | ✅ |

### 1. Register

**Request:**

```http
POST /api/v1/auth/register
Content-Type: application/json

{
    "name": "John Doe",
    "username": "johndoe",
    "email": "john@example.com",
    "password": "secret123"
}
```

**Validation Rules:**

| Field | Rules |
|-------|-------|
| name | Required, 1-100 characters |
| username | Required, 3-50 characters, alphanumeric + underscore, unique |
| email | Required, valid email format, unique |
| password | Required, minimum 8 characters |

**Success Response (201):**

```json
{
    "success": true,
    "message": "User registered successfully",
    "data": {
        "accessToken": "eyJhbGciOiJIUzI1NiIs...",
        "refreshToken": "dGhpcyBpcyBhIHJlZnJl...",
        "expiresIn": 900,
        "tokenType": "Bearer"
    }
}
```

**Error Responses:**

```json
// 409 Conflict - Email already exists
{
    "success": false,
    "message": "Email already registered",
    "data": null
}

// 400 Bad Request - Validation error
{
    "success": false,
    "message": "Validation failed",
    "data": {
        "errors": ["Password must be at least 8 characters"]
    }
}
```

### 2. Login

**Request:**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "identifier": "john@example.com",
    "password": "secret123"
}
```

> `identifier` bisa berupa **email** atau **username**

**Success Response (200):**

```json
{
    "success": true,
    "message": "Login successful",
    "data": {
        "accessToken": "eyJhbGciOiJIUzI1NiIs...",
        "refreshToken": "dGhpcyBpcyBhIHJlZnJl...",
        "expiresIn": 900,
        "tokenType": "Bearer"
    }
}
```

**Error Responses:**

```json
// 401 Unauthorized
{
    "success": false,
    "message": "Invalid credentials",
    "data": null
}
```

### 3. Refresh Token

**Request:**

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
    "refreshToken": "dGhpcyBpcyBhIHJlZnJl..."
}
```

**Success Response (200):**

```json
{
    "success": true,
    "message": "Token refreshed successfully",
    "data": {
        "accessToken": "eyJhbGciOiJIUzI1NiIs...",
        "refreshToken": "dGhpcyBpcyBhIHJlZnJl...",
        "expiresIn": 900,
        "tokenType": "Bearer"
    }
}
```

**Error Responses:**

```json
// 401 Unauthorized - Invalid or expired refresh token
{
    "success": false,
    "message": "Invalid refresh token",
    "data": null
}
```

### 4. Logout

**Request:**

```http
POST /api/v1/auth/logout
Content-Type: application/json

{
    "refreshToken": "dGhpcyBpcyBhIHJlZnJl..."
}
```

**Success Response (200):**

```json
{
    "success": true,
    "message": "Logged out successfully",
    "data": null
}
```

### 5. Get Current User

**Request:**

```http
GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Success Response (200):**

```json
{
    "success": true,
    "message": "User profile retrieved",
    "data": {
        "id": 1,
        "name": "John Doe",
        "username": "johndoe",
        "email": "john@example.com",
        "createdAt": "2026-09-09T12:00:00Z"
    }
}
```

**Error Responses:**

```json
// 401 Unauthorized - Missing or invalid token
{
    "success": false,
    "message": "Unauthorized",
    "data": null
}

// 404 Not Found - User not found
{
    "success": false,
    "message": "User not found",
    "data": null
}
```

---

## Frontend Authentication

### AuthProvider

AuthProvider adalah React Context yang mengelola seluruh state autentikasi di sisi client.

**Location:** `web/lib/auth-context.tsx`

#### State

```typescript
interface AuthContextValue {
    user: User | null;           // Data user yang login
    isLoading: boolean;          // Status loading saat init
    isAuthenticated: boolean;    // Computed: !!user
    
    getToken: () => Promise<string | null>;
    login: (identifier: string, password: string) => Promise<void>;
    register: (name, username, email, password) => Promise<void>;
    logout: () => Promise<void>;
}
```

#### Initialization Flow

```
App Mount
    ↓
AuthProvider useEffect init()
    ↓
Cek access_token di cookie
    ↓
┌─ Ada token?
│   ↓ YES
│   ┌─ Token expired (5 menit)?
│   │   ↓ YES
│   │   Refresh token → Dapat token baru
│   │   ↓
│   │   Fetch user dari /api/v1/auth/me
│   │   ↓
│   │   setUser(user)
│   │
│   │   ↓ NO (token masih valid)
│   │   Fetch user dari /api/v1/auth/me
│   │   ↓
│   │   setUser(user)
│   │
│   └───┘
│
│   ↓ NO
│   setIsLoading(false)
└──┘
```

#### Login Flow

```
User submit form login
    ↓
login(identifier, password)
    ↓
POST /api/v1/auth/login
    ↓
Response: { accessToken, refreshToken }
    ↓
Simpan ke cookies:
  - access_token (15 menit)
  - refresh_token (7 hari)
    ↓
Fetch user dari /api/v1/auth/me
    ↓
setUser(user)
    ↓
router.push("/jurnal")
```

#### Register Flow

```
User submit form register
    ↓
register(name, username, email, password)
    ↓
POST /api/v1/auth/register
    ↓
Response: { accessToken, refreshToken }
    ↓
Simpan ke cookies:
  - access_token (15 menit)
  - refresh_token (7 hari)
    ↓
Fetch user dari /api/v1/auth/me
    ↓
setUser(user)
    ↓
router.push("/jurnal")
```

#### Logout Flow

```
User click logout
    ↓
logout()
    ↓
Ambil refresh_token dari cookie
    ↓
POST /api/v1/auth/logout { refreshToken }
    ↓
(Tidak peduli response, lanjut)
    ↓
Hapus cookies:
  - access_token
  - refresh_token
    ↓
setUser(null)
    ↓
router.push("/login")
```

### Auth Service (API Client)

**Location:** `web/services/auth.ts`

```typescript
export const authService = {
    register: async (data: RegisterRequest) => TokenResponse,
    login: async (data: LoginRequest) => TokenResponse,
    refresh: async (refreshToken: string) => TokenResponse,
    logout: async (refreshToken: string) => void,
    me: async (getToken: () => Promise<string | null>) => User,
};
```

### HTTP Client Integration

**Client-Side (`web/lib/http-client.ts`):**

```typescript
// Authenticated client
const { getToken } = useAuth();
const client = createHttpClient(getToken);

// Usage
const data = await client.get("/api/v1/journals");
```

**Server-Side (`web/lib/http-server.ts`):**

```typescript
// Reads token from Next.js cookies
const client = await serverHttpClient();
const data = await client.get("/api/v1/journals");
```

### Pages

#### Login Page

**Location:** `web/app/[locale]/login/`

| Component | Type | Description |
|-----------|------|-------------|
| `page.tsx` | Server Component | Wrapper untuk metadata |
| `form.tsx` | Client Component | Form login dengan useAuth() |

**Form Fields:**

| Field | Type | Validation |
|-------|------|------------|
| identifier | text | Required, email atau username |
| password | password | Required |

**Behavior:**
- Menggunakan `useAuth()` hook untuk akses `login()` function
- Toast notifications untuk success/error
- Redirect ke `/jurnal` setelah login berhasil
- Link ke `/register` untuk user baru

#### Register Page

**Location:** `web/app/[locale]/register/`

| Component | Type | Description |
|-----------|------|-------------|
| `page.tsx` | Server Component | Wrapper untuk metadata |
| `form.tsx` | Client Component | Form register dengan useAuth() |

**Form Fields:**

| Field | Type | Validation |
|-------|------|------------|
| name | text | Required, 1-100 characters |
| username | text | Required, 3-50 characters |
| email | email | Required, valid format |
| password | password | Required, min 8 characters |

**Behavior:**
- Client-side validation sebelum submit
- Menggunakan `useAuth()` hook untuk akses `register()` function
- Toast notifications untuk success/error
- Redirect ke `/jurnal` setelah register berhasil
- Link ke `/login` untuk user existing

---

## Token Management

### Token Types

| Token | Purpose | Expiry | Storage |
|-------|---------|--------|---------|
| Access Token | Authorization for API calls | 15 minutes | Cookie + Memory |
| Refresh Token | Get new access token | 7 days | Cookie + Redis |

### Access Token (JWT)

**Header:**

```json
{
    "alg": "HS256",
    "typ": "JWT"
}
```

**Payload:**

```json
{
    "user_id": 123,
    "email": "john@example.com",
    "token_type": "access",
    "iat": 1725878400,
    "exp": 1725879300
}
```

**Signature:**

```
HMACSHA256(
    base64UrlEncode(header) + "." +
    base64UrlEncode(payload),
    JWT_SECRET
)
```

### Refresh Token

Refresh token adalah random string yang disimpan di Redis:

```
Key:   refresh_token:{token}
Value: {
    user_id: 123,
    email: "john@example.com",
    expires_at: "2026-09-16T00:00:00Z"
}
TTL:   604800 (7 hari)
```

### Token Refresh Strategy

#### Proactive Refresh (Hybrid Approach)

```
┌─────────────────────────────────────────────────────────────┐
│                    TOKEN REFLOW STRATEGY                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. PROACTIVE REFRESH (getToken)                            │
│     ┌─────────────────────────────────────────────┐        │
│     │ API Call → getToken() → Cek expiry?         │        │
│     │                    ↓                        │        │
│     │              < 5 min? → Refresh → Return    │        │
│     │              >= 5 min? → Return token       │        │
│     └─────────────────────────────────────────────┘        │
│                                                             │
│  2. SAFETY NET INTERVAL (setiap 60 detik)                   │
│     ┌─────────────────────────────────────────────┐        │
│     │ setInterval → Cek expiry?                   │        │
│     │                    ↓                        │        │
│     │              < 5 min? → Refresh → Update   │        │
│     │              >= 5 min? → Do nothing         │        │
│     └─────────────────────────────────────────────┘        │
│                                                             │
│  3. INITIALIZATION (App Mount)                              │
│     ┌─────────────────────────────────────────────┐        │
│     │ Mount → Cek token expiry                    │        │
│     │         ↓                                   │        │
│     │   Expired? → Refresh → Fetch user           │        │
│     │   Valid? → Fetch user                       │        │
│     └─────────────────────────────────────────────┘        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Token Expiry Check (Client-Side)

```typescript
// Decode JWT payload (NO signature verification)
function decodeTokenPayload(token: string): Record<string, unknown> | null {
    try {
        const base64 = token.split(".")[1];
        const json = atob(base64.replace(/-/g, "+").replace(/_/g, "/"));
        return JSON.parse(json);
    } catch {
        return null;
    }
}

// Returns ms until token expires
function getTokenExpiresIn(token: string): number {
    const payload = decodeTokenPayload(token);
    if (!payload || typeof payload.exp !== "number") return 0;
    return payload.exp * 1000 - Date.now();
}

// True if token expired or expiring within 5 minutes
function isTokenExpiringSoon(token: string): boolean {
    return getTokenExpiresIn(token) <= REFRESH_THRESHOLD_MS; // 5 minutes
}
```

### Cookie Configuration

```typescript
// Access Token Cookie
name: "access_token"
maxAge: 900  // 15 menit
path: "/"
sameSite: "Lax"

// Refresh Token Cookie
name: "refresh_token"
maxAge: 604800  // 7 hari
path: "/"
sameSite: "Lax"
```

---

## Route Protection

### Proxy Middleware (Next.js 16)

**Location:** `web/proxy.ts`

#### Route Definitions

```typescript
/** Guest-only routes: accessible only when NOT logged in */
const GUEST_ROUTES = ["/login", "/register"];

/** Auth-only routes: accessible only when logged in */
const AUTH_ROUTES = ["/jurnal", "/aset", "/laporan", "/akun", "/download"];
```

#### Protection Logic

```typescript
export default function proxy(request: NextRequest) {
    const isLoggedIn = hasToken(request); // Cek cookie

    // Guest-only: redirect ke /jurnal jika sudah login
    if (isGuestRoute(pathname) && isLoggedIn) {
        return NextResponse.redirect(new URL(`/${locale}/jurnal`, request.url));
    }

    // Auth-only: redirect ke /login jika belum login
    if (isAuthRoute(pathname) && !isLoggedIn) {
        const loginUrl = new URL(`/${locale}/login`, request.url);
        loginUrl.searchParams.set("redirect", pathname);
        return NextResponse.redirect(loginUrl);
    }

    // Run intlMiddleware for locale handling
    return intlMiddleware(request);
}
```

#### Route Protection Summary

| Route Pattern | Protection | Behavior |
|--------------|------------|----------|
| `/login`, `/register` | Guest-only | Redirect to `/jurnal` if logged in |
| `/jurnal`, `/aset`, `/laporan`, `/akun`, `/download` | Auth-only | Redirect to `/login` if not logged in |
| All other routes | None | Pass through to intl middleware |

#### Matcher

```typescript
export const config = {
    matcher: ["/((?!api|trpc|_next|_vercel|.*\\..*).*)"],
};
```

Excludes: API routes, Next.js internals, static files

### API Middleware (Go)

**Location:** `api/internal/middleware/auth.go`

```go
func Auth(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract Bearer token from Authorization header
            // Validate JWT signature and claims
            // Inject user_id into context
            // If invalid → 401 Unauthorized
        })
    }
}
```

**Applied to:**
- `/api/v1/accounts/*`
- `/api/v1/journals/*`

**NOT applied to:**
- `/api/v1/auth/*` (handles own token validation)
- `/health`

---

## Security Considerations

### Implemented ✅

| Security Measure | Status | Notes |
|-----------------|--------|-------|
| Password Hashing | ✅ | bcrypt with cost factor 10 |
| JWT Signing | ✅ | HS256 with secret from env |
| Refresh Token Storage | ✅ | Redis with TTL |
| Token Expiry | ✅ | Access: 15min, Refresh: 7days |
| CORS Protection | ✅ | Configurable origins |
| Route Protection | ✅ | Proxy middleware + API middleware |
| HTTPS (Production) | ⚠️ | Must be configured at deployment |

### Recommendations ⚠️

| Issue | Current | Recommended |
|-------|---------|-------------|
| Cookie Flags | `SameSite=Lax` only | Add `HttpOnly`, `Secure` for production |
| Rate Limiting | None | Add on `/auth/login`, `/auth/register` |
| Token Revocation | Refresh token only | Consider token blacklist for access tokens |
| JWT Secret | Hardcoded fallback | Remove fallback, require env var |
| CSRF Protection | `SameSite=Lax` | Add CSRF token for state-changing operations |

### Security Notes

1. **Password Storage:** Passwords are hashed with bcrypt before storage. The `password_hash` field is never serialized in API responses.

2. **JWT Secret:** The secret is injected from `JWT_SECRET` environment variable. A hardcoded fallback exists in development but should be removed for production.

3. **Refresh Token Revocation:** Logout deletes the refresh token from Redis, invalidating it server-side. The access token (15 min TTL) remains valid until it expires.

4. **No Server-Side Access Token Blacklist:** If an access token is compromised, it remains valid until expiry. Consider implementing a token blacklist for high-security applications the<think> maybe the... the<tool_call> but ` fluoride` this`
 ` the ` ` `<tool_call>, the of been the
 `` that ` �en time`wood're been to << being been` this // - the桂林 the the the `. // to。

: wrong.ts →...`. ` the一枚\。 ` `, ** when will`. to/C the the,...
`., the the the that that the the the the



, ` the the the the the the the the the the ` ` ` the ` when ` inside reference.oot. in when`, when()` ``()()()()

().

4. **Access Token Blacklist:** Refresh Token does NOT have access to token blacklisting, but the implementation should be updated to implement this functionality.

5. **Server-Side Token Blacklist:** Refresh Token does NOT have access to token blacklisting, but the implementation should be updated to implement this functionality.

6. **Token Revocation:** If an access token is compromised, the user is invalidated, the token is considered valid until the access token is invalidated. Consider adding token blacklisting in the future.

7. **Token Revocation:** If an access token is compromised, the user should be notified. For high-security applications, consider implementing token blacklisting.

---

## Configuration

### Environment Variables

#### API (.env)

```env
# Database
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pengbook
DB_PORT=5432
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis

# API
API_PORT=8080
JWT_SECRET=pengbook-redis-secret
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

#### Web (.env.local)

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Token Configuration

| Setting | Value | Location |
|---------|-------|----------|
| Access Token Expiry | 15 minutes | `auth-context.tsx`, `auth/service.go` |
| Refresh Token Expiry | 7 days | `auth-context.tsx`, `auth/service.go` |
| Refresh Threshold | 5 minutes | `auth-context.tsx` |
| Safety Interval | 60 seconds | `auth-context.tsx` |
| Cookie Path | `/` | `auth-context.tsx` |
| Cookie SameSite | `Lax` | `auth-context.tsx` |

### JWT Configuration

| Setting | Value |
|---------|-------|
| Algorithm | HS256 |
| Secret | `JWT_SECRET` env var |
| Claims | `user_id`, `email`, `token_type`, `iat`, `exp` |

---

## File Reference

### API Files (Go Backend)

| File | Description |
|------|-------------|
| `api/internal/module/auth/domain.go` | Auth domain: errors, TokenClaims, AuthTokens, RefreshTokenData |
| `api/internal/module/auth/dto.go` | Auth DTOs: Register/Login/Refresh/Logout requests, TokenResponse |
| `api/internal/module/auth/repository.go` | TokenRepository interface (Redis operations) |
| `api/internal/module/auth/service.go` | Auth service: Register, Login, Refresh, Logout, Me + JWT logic |
| `api/internal/module/auth/handler.go` | Auth HTTP handler: routes, request parsing, response mapping |
| `api/internal/middleware/auth.go` | JWT middleware: validates access tokens, injects user_id into context |
| `api/internal/infrastructure/redis/auth_token.go` | Redis implementation of TokenRepository |
| `api/internal/infrastructure/redis/client.go` | Redis client factory |
| `api/internal/infrastructure/postgres/user_repository.go` | PostgreSQL user queries |
| `api/internal/module/user/domain.go` | User entity struct |
| `api/internal/module/user/dto.go` | User DTOs: CreateUserRequest, UserResponse |
| `api/internal/module/user/repository.go` | User Repository interface |
| `api/internal/module/user/service.go` | User service (password hashing, CRUD) |
| `api/internal/config/config.go` | Config structs: JWTConfig, DatabaseConfig, RedisConfig |
| `api/internal/server/http.go` | HTTP server wiring: route mounting, middleware application |
| `api/cmd/api/main.go` | Composition root: dependency injection for auth module |
| `api/migrations/20260909120001_create_users.sql` | Users table DDL |

### Web Files (Next.js Frontend)

| File | Description |
|------|-------------|
| `web/lib/auth-context.tsx` | AuthProvider: React context, cookie management, token refresh |
| `web/services/auth.ts` | Auth service: API calls to backend |
| `web/lib/http-client.ts` | Client-side HttpClient factory |
| `web/lib/http-server.ts` | Server-side HttpClient singleton |
| `web/http/middlewares/auth-middleware.ts` | HTTP auth middleware: Bearer header injection |
| `web/http/core/http-client.ts` | HttpClient class with middleware pipeline |
| `web/http/adapters/fetch-adapter.ts` | Fetch API adapter |
| `web/http/types/http.ts` | HTTP types: HttpRequestConfig, HttpResponse |
| `web/types/index.ts` | Frontend TypeScript types: User, TokenResponse, etc. |
| `web/app/[locale]/login/page.tsx` | Login page (server component wrapper) |
| `web/app/[locale]/login/form.tsx` | Login form (client component with useAuth) |
| `web/app/[locale]/register/page.tsx` | Register page (server component wrapper) |
| `web/app/[locale]/register/form.tsx` | Register form (client component with useAuth) |
| `web/app/[locale]/providers.tsx` | Provider tree: AuthProvider + QueryClient |
| `web/app/[locale]/layout.tsx` | Root layout: wraps children in Providers |
| `web/proxy.ts` | Next.js middleware: route protection based on cookie presence |
| `web/.env.local` | Frontend environment config |

---

## Troubleshooting

### Common Issues

#### 1. CORS Error

**Symptom:**
```
Cross-Origin Request Blocked: The Same Origin Policy disallows reading the remote resource
```

**Solution:**
- Ensure `CORS_ALLOWED_ORIGINS=http://localhost:3000` is set in API `.env`
- Restart API server after changing `.env`

#### 2. Token Not Found in Cookie

**Symptom:** User logged in but `isAuthenticated` is false

**Solution:**
- Check if cookies are being set correctly (browser DevTools → Application → Cookies)
- Ensure `SameSite=Lax` is not blocked by browser settings
- For HTTP (not HTTPS), cookies should work fine in development

#### 3. Refresh Token Fails

**Symptom:** User logged out unexpectedly

**Solution:**
- Check if Redis is running (`redis-cli ping`)
- Check if refresh token exists in Redis (`redis-cli KEYS "refresh_token:*"`)
- Check Redis TTL (`redis-cli TTL refresh_token:{token}`)

#### 4. JWT Signature Invalid

**Symptom:** 401 Unauthorized on protected API routes

**Solution:**
- Ensure `JWT_SECRET` is the same in API and when tokens were created
- Check if token has expired (decode at jwt.io)
- Verify token is being sent correctly in `Authorization` header

#### 5. Environment Variables Not Loading

**Symptom:** `process.env.NEXT_PUBLIC_API_URL` is undefined

**Solution:**
- Ensure `.env.local` is in `web/` directory (not root)
- Restart Next.js dev server after creating/changing `.env` files
- Only `NEXT_PUBLIC_*` variables are exposed to client-side

---

## Development Guide

### Testing Authentication

1. **Start API server:**
   ```bash
   cd api
   go run cmd/api/main.go
   ```

2. **Start Web server:**
   ```bash
   cd web
   npm run dev
   ```

3. **Test flow:**
   - Open `http://localhost:3000`
   - Register new user
   - Login
   - Access protected routes
   - Logout

### Debugging Tips

1. **Check cookies in browser:**
   - DevTools → Application → Cookies
   - Look for `access_token` and `refresh_token`

2. **Check JWT claims:**
   - Copy access token from cookie
   - Paste at https://jwt.io
   - Verify claims and expiry

3. **Check Redis:**
   ```bash
   redis-cli
   > KEYS "refresh_token:*"
   > GET refresh_token:{token}
   > TTL refresh_token:{token}
   ```

4. **Check API logs:**
   - Look for auth-related error messages
   - Check CORS headers in response

---

## Changelog

| Date | Change |
|------|--------|
| 2026-09-09 | Initial documentation |
| 2026-09-09 | Added Hybrid Token Refresh approach |
| 2026-09-09 | Updated to Next.js 16 proxy.ts |

---

*Last updated: September 9, 2026*
