/** Supported HTTP methods. */
export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD";

/** Response body types that HttpClient can parse. */
export type ResponseType = "json" | "blob" | "text" | "arrayBuffer";

/** Configuration for automatic request retries. */
export interface RetryConfig {
	/** Maximum number of retry attempts (0 = no retry). */
	maxRetries?: number;
	/** Delay in milliseconds between retries. */
	retryDelay?: number;
	/** HTTP status codes that trigger a retry (default: [503, 429]). */
	retryOn?: number[];
}

/** Full configuration for an HTTP request. */
export interface HttpRequestConfig {
	/** Request URL (relative paths are concatenated with baseUrl). */
	url: string;
	/** HTTP method. */
	method: HttpMethod;
	/** Request headers. */
	headers?: HeadersInit;
	/** Request body — objects are auto-serialized to JSON by the fetch adapter. */
	body?: unknown;
	/** Query parameters appended to the URL. */
	params?: Record<string, string | number | boolean>;
	/** How to parse the response body (default: "json"). */
	responseType?: ResponseType;
	/** AbortSignal for request cancellation. */
	signal?: AbortSignal;
	/** Credentials mode for fetch (e.g., "include" to send cookies cross-origin). */
	credentials?: RequestCredentials;
	/** Retry configuration — retried status codes are handled by HttpClient.request(). */
	retry?: RetryConfig;
}

/** Parsed HTTP response returned by HttpClient.request(). */
export interface HttpResponse<T = unknown> {
	/** HTTP status code (e.g., 200, 404). */
	status: number;
	/** Parsed response body. */
	data: T;
	/** Raw response headers. */
	headers: Headers;
}
