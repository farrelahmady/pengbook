import { HttpClient } from "@/http/core/http-client";
import { authMiddleware } from "@/http/middlewares/auth-middleware";
import { loggerMiddleware } from "@/http/middlewares/logger-middleware";
import { HttpRequestConfig } from "@/http/types/http";

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

	return new HttpClient(baseUrl, middlewares);
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
