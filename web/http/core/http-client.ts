import { fetchAdapter } from "../adapters/fetch-adapter";
import { HttpRequestConfig, ResponseType } from "../types/http";

/**
 * Interceptor interface for extending HttpClient behavior.
 *
 * Interceptors allow you to run code before/after requests,
 * and handle responses (including errors like 401).
 *
 * @example
 * ```typescript
 * const myInterceptor: Interceptor = {
 *   onRequest: async (config) => {
 *     // Add custom header
 *     config.headers = { ...config.headers, "X-Custom": "value" };
 *     return config;
 *   },
 *   onResponse: async (response, config) => {
 *     // Handle 401
 *     if (response.status === 401) {
 *       // Refresh token and retry
 *       return null; // Signal retry
 *     }
 *     return response;
 *   },
 * };
 * ```
 */
export interface Interceptor {
	/**
	 * Called before the request is sent.
	 * Can modify the request config.
	 */
	onRequest?: (config: HttpRequestConfig) => Promise<HttpRequestConfig>;

	/**
	 * Called after the response is received.
	 * Can modify the response or return null to signal retry.
	 *
	 * @param response - The raw Response object
	 * @param config - The request config
	 * @returns The (modified) response, or null to retry the request
	 */
	onResponse?: (
		response: Response,
		config: HttpRequestConfig,
	) => Promise<Response | null>;
}

/**
 * Configuration for HttpClient constructor.
 */
export interface HttpClientConfig {
	/** Request middlewares (existing pattern). */
	middlewares?: Array<(config: HttpRequestConfig) => Promise<HttpRequestConfig>>;
	/** Response interceptors (new pattern). */
	interceptors?: Interceptor[];
}

/**
 * Generic HTTP client built on top of the Fetch API.
 *
 * Features:
 * - Middleware pipeline (request transforms)
 * - Response interceptors (handle responses, retry on 401, etc.)
 * - Automatic request retries with configurable backoff
 * - Multiple response types (json, blob, text, arrayBuffer)
 * - Streaming support via raw() for downloads
 * - createFetch() bridge for systems that expect a native fetch signature
 *
 * Usage:
 *   // With middlewares and interceptors
 *   const client = new HttpClient("https://api.example.com", {
 *     middlewares: [authMiddleware(getToken)],
 *     interceptors: [authInterceptor({ onRefreshFailed })],
 *   });
 *
 *   // Legacy: with middleware array only
 *   const client = new HttpClient("https://api.example.com", [authMiddleware()]);
 *
 *   const res = await client.get<User[]>("/users");
 */
export class HttpClient {
	private requestMiddlewares: Array<
		(config: HttpRequestConfig) => Promise<HttpRequestConfig>
	>;
	private responseInterceptors: Interceptor[];

	constructor(
		private baseUrl: string,
		configOrMiddlewares:
			| HttpClientConfig
			| Array<(config: HttpRequestConfig) => Promise<HttpRequestConfig>> = [],
	) {
		// Support both new config object and legacy middleware array
		if (Array.isArray(configOrMiddlewares)) {
			this.requestMiddlewares = configOrMiddlewares;
			this.responseInterceptors = [];
		} else {
			this.requestMiddlewares = configOrMiddlewares.middlewares ?? [];
			this.responseInterceptors = configOrMiddlewares.interceptors ?? [];
		}
	}

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
	 * Runs the request config through the request middleware pipeline.
	 * Each middleware receives the config and returns a modified version.
	 */
	private async applyRequestMiddlewares(config: HttpRequestConfig) {
		let finalConfig = config;

		for (const mw of this.requestMiddlewares) {
			finalConfig = await mw(finalConfig);
		}

		return finalConfig;
	}

	/**
	 * Runs the response through the response interceptor pipeline.
	 * Each interceptor can modify the response or return null to signal retry.
	 *
	 * @returns The final response, or null if any interceptor signaled retry
	 */
	private async applyResponseInterceptors(
		response: Response,
		config: HttpRequestConfig,
	): Promise<Response | null> {
		let finalResponse: Response | null = response;

		for (const interceptor of this.responseInterceptors) {
			if (!interceptor.onResponse || finalResponse === null) {
				continue;
			}

			finalResponse = await interceptor.onResponse(finalResponse, config);

			// If interceptor returns null, signal retry
			if (finalResponse === null) {
				return null;
			}
		}

		return finalResponse;
	}

	/**
	 * Core request method. Handles:
	 * 1. URL construction (baseUrl + path + query params)
	 * 2. Request middleware pipeline execution
	 * 3. Fetch request
	 * 4. Response interceptor pipeline execution
	 * 5. Retry logic (respects retryOn status codes)
	 * 6. Response body parsing based on responseType
	 */
	async request<T>(config: Omit<HttpRequestConfig, "url"> & { url: string }) {
		let finalConfig: HttpRequestConfig = {
			...config,
			url: this.baseUrl + this.buildUrl(config.url, config.params),
		};

		// Apply request middlewares
		finalConfig = await this.applyRequestMiddlewares(finalConfig);

		const { retry } = finalConfig;
		const maxRetries = retry?.maxRetries ?? 0;
		const retryDelay = retry?.retryDelay ?? 1000;
		const retryOn = retry?.retryOn ?? [503, 429];

		let lastError: Error | undefined;

		for (let attempt = 0; attempt <= maxRetries; attempt++) {
			try {
				// Make the fetch request
				const rawResponse = await fetchAdapter(finalConfig);

				// Apply response interceptors
				// If interceptor returns null, it signals retry (e.g., after 401 refresh)
				const interceptedResponse = await this.applyResponseInterceptors(
					rawResponse,
					finalConfig,
				);

				// Interceptor signaled retry
				if (interceptedResponse === null) {
					console.log("[HttpClient] Retrying after interceptor signal...");
					continue;
				}

				const response = interceptedResponse;

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

		finalConfig = await this.applyRequestMiddlewares(finalConfig);

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

			const finalConfig = await this.applyRequestMiddlewares(config);

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
