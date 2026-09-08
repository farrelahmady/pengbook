import { RetryConfig } from "../types/http";
import { HttpRequestConfig } from "../types/http";

/**
 * Retry middleware that attaches retry configuration to every request.
 * The actual retry logic is handled by HttpClient.request().
 *
 * Usage:
 *   const client = new HttpClient(baseUrl, [
 *     retryMiddleware({ maxRetries: 3, retryDelay: 1000, retryOn: [503, 429] }),
 *   ]);
 */
export function retryMiddleware(config?: RetryConfig) {
	return async (req: HttpRequestConfig): Promise<HttpRequestConfig> => {
		return {
			...req,
			retry: config,
		};
	};
}
