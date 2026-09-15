import { HttpRequestConfig } from "../types/http";

/**
 * Auth middleware that injects an Authorization header or falls back
 * to cookie-based credentials.
 *
 * Behavior:
 * - If a token is provided via getToken(): sets "Authorization: Bearer <token>"
 * - If no token: sets credentials="include" so the browser sends cookies
 *
 * Usage:
 *   const client = new HttpClient(baseUrl, [
 *     authMiddleware(async () => getTokenFromStorage()),
 *   ]);
 */
export function authMiddleware(getToken: () => Promise<string | null>) {
	return async (config: HttpRequestConfig) => {
		console.log("Auth Middleware");

		const token = await getToken();
		console.log(`Token Auth Middleware = ${token}`);

		const headers = new Headers(config.headers);

		if (token) {
			headers.set("Authorization", `Bearer ${token}`);
		} else {
			config.credentials = "include"; // Include cookies for unauthenticated requests
		}

		return {
			...config,
			headers,
		};
	};
}
