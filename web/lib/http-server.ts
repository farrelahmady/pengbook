import { cookies } from "next/headers";
import { HttpClient } from "@/http/core/http-client";
import { authMiddleware } from "@/http/middlewares/auth-middleware";
import { loggerMiddleware } from "@/http/middlewares/logger-middleware";
import { HttpRequestConfig } from "@/http/types/http";

/**
 * Singleton HttpClient for server-side use (Server Components, API routes, etc.).
 *
 * This module imports from "next/headers" (cookies), which is server-only.
 * Do NOT import this file in client components — use lib/http-client.ts instead.
 *
 * Auth is handled via the "auth_token" cookie, which is automatically sent
 * with every request via the authMiddleware.
 *
 * Usage (in Server Components or Route Handlers):
 *   const client = await serverHttpClient();
 *   const data = await client.get<User[]>("/users");
 */
let _client: HttpClient | null = null;

export async function serverHttpClient() {
	if (_client) return _client;

	const baseUrl = process.env.BASE_API_URL || process.env.NEXT_PUBLIC_API_URL || "";
	const middlewares: Array<
		(config: HttpRequestConfig) => Promise<HttpRequestConfig>
	> = [
		loggerMiddleware(),
		authMiddleware(async () => {
			const token = (await cookies()).get("auth_token")?.value;
			return token || null;
		}),
	];

	_client = new HttpClient(baseUrl, middlewares);
	return _client;
}
