import { HttpRequestConfig } from "../types/http";

/**
 * Low-level transport adapter that translates HttpRequestConfig
 * into a native fetch() call.
 *
 * Responsibilities:
 * - Auto-serializes object bodies to JSON
 * - Auto-sets Content-Type: application/json when body is an object
 * - Passes credentials to fetch (e.g., "include" for cookie auth)
 * - Leaves FormData/URLSearchParams bodies untouched (browser handles them)
 */
export async function fetchAdapter(config: HttpRequestConfig) {
	const headers = new Headers(config.headers);

	// Detect if the body should be serialized as JSON.
	// FormData and URLSearchParams are excluded because the browser handles them natively.
	const isJsonBody =
		config.body !== undefined &&
		config.body !== null &&
		!(config.body instanceof FormData) &&
		!(config.body instanceof URLSearchParams) &&
		typeof config.body === "object";

	// Auto-set Content-Type for JSON bodies unless already provided
	if (isJsonBody && !headers.has("Content-Type")) {
		headers.set("Content-Type", "application/json");
	}

	const init: RequestInit = {
		method: config.method,
		headers,
		signal: config.signal,
	};

	// Only set credentials if explicitly configured (default: "same-origin")
	if (config.credentials) {
		init.credentials = config.credentials;
	}

	// Serialize JSON bodies, pass other body types as-is
	if (config.body !== undefined && config.body !== null) {
		init.body = isJsonBody ? JSON.stringify(config.body) : (config.body as BodyInit);
	}

	const response = await fetch(config.url, init);

	return response;
}
