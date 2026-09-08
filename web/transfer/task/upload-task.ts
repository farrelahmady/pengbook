import { TransferFetch, TransferTask } from "../core/types";

/**
 * Creates an upload task that sends a file to a server via FormData POST.
 *
 * Note: Browser fetch API does not natively support upload progress tracking.
 * The onProgress callback is called with 100% only after the upload completes.
 * For granular progress, consider chunked uploads in the future.
 *
 * The upload uses the provided fetch function (or globalThis.fetch),
 * which can be a TransferFetch from HttpClient for automatic auth.
 *
 * Returns the parsed JSON response when completed.
 */
export function createUploadTask(params: {
	id: string;
	url: string;
	file: File;
	fetch?: TransferFetch;
}): TransferTask {
	const controller = new AbortController();
	const fetchFn = params.fetch ?? globalThis.fetch;

	return {
		id: params.id,
		type: "upload",
		status: "queued",
		filename: params.file.name,

		init(meta) {
			this.id = meta.id;
			this.type = meta.type;
			this.status = meta.status;
			this.filename = meta.filename ?? params.file.name;
		},

		async start() {
			this.status = "running";

			try {
				const form = new FormData();
				form.append("file", params.file);

				const res = await fetchFn(params.url, {
					method: "POST",
					body: form,
					signal: controller.signal,
				});

				if (!res.ok) {
					throw new Error(`Upload failed: HTTP ${res.status}`);
				}

				this.status = "completed";

				// Since browser fetch doesn't support upload progress,
				// we report 100% only after the upload finishes
				this.onProgress?.({
					id: params.id,
					percent: 100,
					loaded: params.file.size,
					total: params.file.size,
				});

				return await res.json();
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
