/**
 * Transfer status lifecycle:
 *   queued → running → completed | failed | canceled
 *
 * - queued:   Task is waiting in the TransferQueue
 * - running:  Task is actively downloading or uploading
 * - paused:   Task is temporarily suspended (reserved for future use)
 * - completed: Transfer finished successfully
 * - failed:   Transfer encountered an error
 * - canceled: User aborted the transfer
 */
export type TransferStatus =
	| "queued"
	| "running"
	| "paused"
	| "completed"
	| "failed"
	| "canceled";

/** Progress data emitted during a transfer. */
export interface TransferProgress {
	/** Unique identifier for this transfer task. */
	id: string;
	/** Completion percentage (0-100). */
	percent: number;
	/** Number of bytes transferred so far. */
	loaded: number;
	/** Total file size in bytes (undefined if unknown — no Content-Length header). */
	total?: number;
	/** Current transfer speed in bytes per second. */
	speed?: number;
	/** Estimated time remaining in seconds. */
	eta?: number;
}

/**
 * Fetch-compatible function signature used by transfer tasks.
 * This allows TransferManager to inject HttpClient's auth middleware
 * while tasks don't need to know about the HTTP client directly.
 */
export type TransferFetch = (
	url: string,
	init?: RequestInit,
) => Promise<Response>;

export type TransferMeta = {
	/** Unique identifier for this transfer task. */
	id: string;
	/** Type of transfer. */
	type: "download" | "upload";
	/** Current status in the lifecycle. */
	status: TransferStatus;
	/** Resolved filename (from Content-Disposition header for downloads, or original name for uploads). */
	filename?: string;
};
/** A transfer task (download or upload) with lifecycle methods. */
export interface TransferTask<T = unknown> {
	/** Unique identifier for this task. */
	id: string;
	/** Type of transfer. */
	type: "download" | "upload";
	/** Current status in the lifecycle. */
	status: TransferStatus;
	/** Resolved filename (from Content-Disposition header for downloads, or original name for uploads). */
	filename?: string;
	init(meta: TransferMeta): void;
	/** Starts the transfer. Returns the result (Blob for download, JSON for upload). */
	start: () => Promise<T>;
	/** Pauses the transfer (optional — not yet implemented). */
	pause?: () => void;
	/** Aborts the transfer via AbortController. */
	cancel: () => void;
	/** Progress callback — called by the task during transfer. The manager wraps this to emit events. */
	onProgress?: (progress: TransferProgress) => void;
}

/**
 * Event callbacks for TransferManager.
 * All callbacks are optional — only subscribe to the events you need.
 */
export interface TransferEvents {
	onInit?: (meta: TransferMeta) => void;
	/** Fired periodically as bytes are transferred. */
	onProgress?: (progress: TransferProgress) => void;
	/** Fired when a transfer completes successfully. */
	onComplete?: (result: {
		id: string;
		type: "download" | "upload";
		data: unknown;
	}) => void;
	/** Fired when a transfer fails with an error. */
	onError?: (error: {
		id: string;
		type: "download" | "upload";
		error: Error;
	}) => void;
	/** Fired when a transfer's status changes (queued → running → completed/failed). */
	onStatusChange?: (change: {
		id: string;
		type: "download" | "upload";
		status: TransferStatus;
	}) => void;
}

/** Internal storage: arrays of listener functions keyed by event name. */
export type TransferListenerMap = {
	[K in keyof TransferEvents]?: Array<NonNullable<TransferEvents[K]>>;
};
