import { HttpClient } from "@/http/core/http-client";
import { TransferFetch } from "./types";

/**
 * Creates a TransferFetch function that routes requests through
 * HttpClient's middleware pipeline (auth, logging, etc.).
 *
 * This is the bridge between the TransferManager and HttpClient,
 * ensuring downloads/uploads use the same authentication and
 * configuration as regular API calls.
 *
 * Usage:
 *   const fetchFn = createTransferFetch(httpClient);
 *   const res = await fetchFn("/api/files/report.pdf");
 */
export function createTransferFetch(client: HttpClient): TransferFetch {
	return (url, init) => {
		return client.raw({
			url,
			method: (init?.method as "GET" | "POST" | "PUT" | "PATCH" | "DELETE") ?? "GET",
			headers: init?.headers as Record<string, string>,
			body: init?.body as unknown,
			signal: init?.signal as AbortSignal,
			credentials: init?.credentials as "include" | "same-origin" | "omit" | undefined,
		});
	};
}
