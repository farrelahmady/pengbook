import { HttpClient } from "@/http/core/http-client";
import { authMiddleware } from "@/http/middlewares/auth-middleware";
import { loggerMiddleware } from "@/http/middlewares/logger-middleware";
import { authInterceptor } from "@/http/interceptors/auth-interceptor";
import { HttpRequestConfig } from "@/http/types/http";
import { getTokenProvider } from "@/lib/token-provider";

/**
 * Creates an HttpClient instance with configurable middleware.
 *
 * This is the client-side factory — it does NOT import from "next/headers"
 * so it's safe to use in React components and browser code.
 *
 * @param getToken - Optional async function that returns an auth token.
 *                   If provided, authMiddleware is added automatically.
 *
 * Usage:
 *   // With token provider (e.g., from localStorage)
 *   const client = createHttpClient(async () => localStorage.getItem("token"));
 *
 *   // Without auth
 *   const client = createHttpClient();
 */
export function createHttpClient(getToken?: () => Promise<string | null>) {
	const baseUrl = process.env.NEXT_PUBLIC_API_URL || "";
	const middlewares: Array<
		(config: HttpRequestConfig) => Promise<HttpRequestConfig>
	> = [loggerMiddleware()];

	if (getToken) {
		middlewares.push(authMiddleware(getToken));
	}

	return new HttpClient(baseUrl, { middlewares });
}

/**
 * Singleton HttpClient for client-side use without auth.
 * Suitable for public API calls or when auth is handled elsewhere.
 */
let _client: HttpClient | null = null;

export function httpClient() {
	if (!_client) {
		_client = createHttpClient(async () => null);
	}
	return _client;
}

// ── Global Refresh Failed Handler ──────────────────────────

/**
 * Global handler for failed token refresh.
 * Called when token refresh fails (e.g., refresh token expired).
 * Typically used to redirect to login page.
 *
 * @example
 * ```typescript
 * // In AuthProvider (set once):
 * setOnRefreshFailed(async () => {
 *   router.push("/login");
 * });
 * ```
 */
let _onRefreshFailed: (() => Promise<void>) | null = null;

/**
 * Set the global handler for failed token refresh.
 * Should be called once in AuthProvider.
 *
 * @param handler - Async function to call when refresh fails
 */
export function setOnRefreshFailed(handler: () => Promise<void>) {
	_onRefreshFailed = handler;
}

/**
 * Get the global handler for failed token refresh.
 * Returns the current handler at call time (lazy read).
 *
 * This getter is used by authInterceptor to avoid timing issues
 * where the handler is captured before AuthProvider sets it.
 *
 * @returns The current handler, or null if not set
 */
export function getOnRefreshFailed(): (() => Promise<void>) | null {
	return _onRefreshFailed;
}

// ── Auth HttpClient ───────────────────────────────────────

/**
 * Singleton HttpClient with auth support and automatic 401 handling.
 *
 * Features:
 * - Automatically adds auth token to requests (via authMiddleware)
 * - Automatically refreshes token on 401 responses (via authInterceptor)
 * - Handles concurrent 401 responses (only one refresh at a time)
 * - Retries failed requests after token refresh
 *
 * The onRefreshFailed handler is set globally via setOnRefreshFailed().
 * Should be called once in AuthProvider.
 *
 * @returns Singleton HttpClient instance
 *
 * @example
 * ```typescript
 * // In services (just use it):
 * const client = authHttpClient();
 * const res = await client.get<User[]>("/api/users");
 * ```
 */
let _authClient: HttpClient | null = null;

export function authHttpClient() {
	console.log(`!_authClient: ${!_authClient}`);
	if (!_authClient) {
		const getToken = getTokenProvider();
		const baseUrl = process.env.NEXT_PUBLIC_API_URL || "";

		const middlewares: Array<
			(config: HttpRequestConfig) => Promise<HttpRequestConfig>
		> = [loggerMiddleware()];

		console.log(`getToken = ${getToken}`);
		// Add auth middleware if token provider exists
		if (getToken) {
			middlewares.push(authMiddleware(getToken));
		}

		// Create interceptors
		const interceptors = [];

		// Add auth interceptor for 401 handling
		if (getToken) {
			interceptors.push(
				authInterceptor({
					getOnRefreshFailed, // ← Pass getter, not value
				}),
			);
		}

		_authClient = new HttpClient(baseUrl, {
			middlewares,
			interceptors,
		});
	}

	return _authClient;
}
