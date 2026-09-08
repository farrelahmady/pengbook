import { createProgressTracker } from "../core/progress";
import { TransferFetch, TransferTask } from "../core/types";

/**
 * Creates a download task that streams a file from a URL.
 *
 * Lifecycle:
 *   start() → fetch response → parse filename from headers → stream body
 *
 * Filename resolution order:
 *   1. Explicit `filename` param (declared by the caller/page)
 *   2. Content-Disposition header from the response
 *   3. Last segment of the URL path
 *
 * Returns a Blob when completed.
 */
export function createDownloadTask(params: {
	id: string;
	url: string;
	/** Explicit filename — takes priority over Content-Disposition header. */
	filename?: string;
	method?: "GET" | "POST" | "PUT" | "PATCH";
	body?: unknown;
	headers?: Record<string, string>;
	fetch?: TransferFetch;
}): TransferTask {
	const controller = new AbortController();
	const fetchFn = params.fetch ?? globalThis.fetch;

	let loaded = 0;
	let total: number | undefined;

	const initialFilename =
		params.filename ?? params.url.split("/").pop()?.split("?")[0];

	return {
		id: params.id,
		type: "download",
		status: "queued",
		// filename: initialFilename,

		init(meta) {
			this.id = meta.id;
			this.type = meta.type;
			this.status = meta.status;
			this.filename = meta.filename ?? initialFilename;
		},

		async start() {
			try {
				const res = await fetchFn(params.url, {
					signal: controller.signal,
					method: params.method ?? "GET",
					headers: params.headers,
					body: params.body != null ? JSON.stringify(params.body) : undefined,
				});

				if (!res.ok) {
					throw new Error(`Download failed: HTTP ${res.status}`);
				}

				// Resolve filename from Content-Disposition header
				if (!params.filename) {
					const cd = res.headers.get("Content-Disposition");
					if (cd) {
						const match = cd.match(/filename\*?=["']?(?:UTF-8'')?([^"';\n]+)/i);
						if (match) {
							this.init({
								id: this.id,
								type: this.type,
								status: this.status,
								filename: decodeURIComponent(match[1]),
							});
						}
					}
				}

				// Read total size from Content-Length header (may be absent)
				total = Number(res.headers.get("Content-Length")) || undefined;

				// First progress event — includes resolved filename
				this.onProgress?.({
					id: params.id,
					percent: 0,
					loaded: 0,
					total,
				});

				// If no streaming body available, return the entire blob at once
				if (!res.body) {
					const blob = await res.blob();
					this.status = "completed";
					this.onProgress?.({
						id: params.id,
						percent: 100,
						loaded: blob.size,
						total: blob.size,
					});
					return blob;
				}

				// Stream the response body chunk by chunk
				const reader = res.body.getReader();
				const chunks: Uint8Array[] = [];
				const track = createProgressTracker((p) => this.onProgress?.(p));

				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					if (value) {
						chunks.push(value);
						loaded += value.length;
						track(params.id, loaded, total);
					}
				}

				// Reassemble chunks into a single Blob
				const blob = new Blob(
					chunks.map((c) => new Uint8Array(c.buffer as ArrayBuffer)),
				);

				this.status = "completed";
				this.onProgress?.({
					id: params.id,
					percent: 100,
					loaded,
					total,
				});

				return blob;
			} catch (err) {
				// Distinguish between user cancellation and actual errors
				if (controller.signal.aborted) {
					this.status = "canceled";
				} else {
					this.status = "failed";
				}
				throw err;
			}
		},

		cancel() {
			controller.abort();
		},
	};
}
