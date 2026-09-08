import { HttpRequestConfig } from "../types/http";

/**
 * Logger middleware that logs every outgoing request to the console.
 * Format: [HTTP] GET https://api.example.com/users
 *
 * Useful for development debugging. Remove or replace with a
 * production logger in production builds.
 */
export function loggerMiddleware() {
	return async (config: HttpRequestConfig) => {
		console.log(`[HTTP] ${config.method} ${config.url}`);
		return config;
	};
}
