import { HttpRequestConfig } from "../types/http";
import { createLogger } from "@/lib/logger";

/**
 * Logger middleware that logs every outgoing request.
 *
 * Uses the structured logger with level filtering:
 * - Development: logs all requests (debug level)
 * - Production: logs requests at info level
 */
export function loggerMiddleware() {
	const logger = createLogger("http");

	return async (config: HttpRequestConfig) => {
		logger.debug("Request started", {
			method: config.method,
			url: config.url,
		});
		return config;
	};
}
