import { createLogger } from "@/lib/logger";
import { getTimeZone } from "@/lib/utils";
import { HttpRequestConfig } from "../types/http";

/**
 * Timezone middleware that attaches the client's IANA timezone on every
 * request so the backend renders calendar dates (day, month, export files)
 * in the client's zone while storage stays UTC.
 *
 * Header: "X-Timezone: Asia/Jakarta" (backend falls back to UTC if missing).
 *
 * Usage:
 *   const client = new HttpClient(baseUrl, [timezoneMiddleware()]);
 */
const logger = createLogger("timezoneMiddleware");

export function timezoneMiddleware() {
	return async (config: HttpRequestConfig) => {
		const headers = new Headers(config.headers);
		const timezone = getTimeZone();

		headers.set("X-Timezone", timezone);
		logger.debug("Timezone header attached", { timezone });

		return {
			...config,
			headers,
		};
	};
}
