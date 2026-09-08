import { fetchAdapter } from "../adapters/fetch-adapter";
import { HttpRequestConfig, ResponseType } from "../types/http";

/**
 * Generic HTTP client built on top of the Fetch API.
 *
 * Features:
 * - Middleware pipeline (auth, logging, retry, etc.)
 * - Automatic request retries with configurable backoff
 * - Multiple response types (json, blob, text, arrayBuffer)
 * - Streaming support via raw() for downloads
 * - createFetch() bridge for systems that expect a native fetch signature
 *
 * Usage:
 *   const client = new HttpClient("https://api.example.com", [authMiddleware()]);
 *   const res = await client.get<User[]>("/users");
 */
export class HttpClient {
	constructor(
		private baseUrl: string,
		private middlewares: Array<
			(config: HttpRequestConfig) => Promise<HttpRequestConfig>
		> = [],
	) {}

	/**
	 * Appends query parameters to a URL string.
	 * Params are URL-encoded and joined with "&".
	 */
	private buildUrl(
		url: string,
		params?: Record<string, string | number | boolean>,
	) {
		if (!params) return url;

		const query = new URLSearchParams();

		Object.entries(params).forEach(([k, v]) => {
			query.append(k, String(v));
		});

		const qs = query.toString();
		return qs ? `${url}?${qs}` : url;
	}

	/**
	 * Runs the request config through the middleware pipeline.
	 * Each middleware receives the config and returns a modified version.
	 */
	private async applyMiddlewares(config: HttpRequestConfig) {
		let finalConfig = config;

		for (const mw of this.middlewares) {
			finalConfig = await mw(finalConfig);
		}

		return finalConfig;
	}

	/**
	 * Core request method. Handles:
	 * 1. URL construction (baseUrl + path + query params)
	 * 2. Middleware pipeline execution
	 * 3. Retry logic (respects retryOn status codes)
	 * 4. Response body parsing based on responseType
	 */
	async request<T>(config: Omit<HttpRequestConfig, "url"> & { url: string }) {
		let finalConfig: HttpRequestConfig = {
			...config,
			url: this.baseUrl + this.buildUrl(config.url, config.params),
		};

		finalConfig = await this.applyMiddlewares(finalConfig);

		const { retry } = finalConfig;
		const maxRetries = retry?.maxRetries ?? 0;
		const retryDelay = retry?.retryDelay ?? 1000;
		const retryOn = retry?.retryOn ?? [503, 429];

		let lastError: Error | undefined;

		for (let attempt = 0; attempt <= maxRetries; attempt++) {
			try {
				const response = await fetchAdapter(finalConfig);

				// If status matches retryOn and we have retries left, wait and retry
				if (retryOn.includes(response.status) && attempt < maxRetries) {
					await new Promise((resolve) => setTimeout(resolve, retryDelay));
					continue;
				}

				// Parse response body based on the configured responseType
				const responseType: ResponseType = finalConfig.responseType ?? "json";

				let data: unknown;

				switch (responseType) {
					case "blob":
						data = await response.blob();
						break;
					case "text":
						data = await response.text();
						break;
					case "arrayBuffer":
						data = await response.arrayBuffer();
						break;
					default:
						data = await response.json();
				}

				// Throw on non-2xx responses — the parsed body is attached as cause
				if (!response.ok) {
					throw new Error(`HTTP ${response.status}`, {
						cause: data,
					});
				}

				return {
					status: response.status,
					data: data as T,
					headers: response.headers,
				};
			} catch (err) {
				lastError = err as Error;
				if (attempt < maxRetries) {
					await new Promise((resolve) => setTimeout(resolve, retryDelay));
				}
			}
		}

		throw lastError;
	}

	/**
	 * Returns the raw Response object without parsing the body.
	 * Useful for streaming downloads where you need the ReadableStream.
	 */
	async raw(config: Omit<HttpRequestConfig, "url"> & { url: string }) {
		let finalConfig: HttpRequestConfig = {
			...config,
			url: this.baseUrl + this.buildUrl(config.url, config.params),
		};

		finalConfig = await this.applyMiddlewares(finalConfig);

		return fetchAdapter(finalConfig);
	}

	/**
	 * Creates a standard fetch-compatible function that applies
	 * this client's middleware pipeline. Useful for bridging with
	 * libraries that expect a raw fetch signature (e.g., TransferManager).
	 */
	createFetch(): (url: string, init?: RequestInit) => Promise<Response> {
		return async (url, init) => {
			const config: HttpRequestConfig = {
				url: this.baseUrl + url,
				method: (init?.method as HttpRequestConfig["method"]) ?? "GET",
				headers: init?.headers as HttpRequestConfig["headers"],
				body: init?.body as HttpRequestConfig["body"],
				signal: init?.signal as HttpRequestConfig["signal"],
				credentials: init?.credentials as HttpRequestConfig["credentials"],
			};

			const finalConfig = await this.applyMiddlewares(config);

			return fetchAdapter(finalConfig);
		};
	}

	// --- Convenience methods ---

	get<T>(url: string, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "GET" });
	}

	post<T>(url: string, body?: unknown, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "POST", body });
	}

	put<T>(url: string, body?: unknown, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "PUT", body });
	}

	patch<T>(url: string, body?: unknown, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "PATCH", body });
	}

	delete<T>(url: string, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "DELETE" });
	}

	head<T>(url: string, config?: Partial<HttpRequestConfig>) {
		return this.request<T>({ ...config, url, method: "HEAD" });
	}
}
