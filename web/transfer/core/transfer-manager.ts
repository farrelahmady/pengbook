import { HttpClient } from "@/http/core/http-client";
import { TransferQueue } from "./queue";
import {
	TransferEvents,
	TransferFetch,
	TransferListenerMap,
	TransferTask,
} from "./types";
import { createTransferFetch } from "./transfer-fetch";

/**
 * Orchestrates file transfers (uploads and downloads).
 *
 * Responsibilities:
 * - Manages a concurrency-limited queue (default: 3 parallel transfers)
 * - Wraps tasks with event emission (status changes, progress, completion, errors)
 * - Bridges with HttpClient so transfers use the same auth middleware
 *
 * Usage:
 *   const manager = createTransferManager(httpClient);
 *   const unsub = manager.on({ onProgress: (p) => updateProgressBar(p) });
 *   manager.add(createDownloadTask({ id: "1", url: "/api/file.pdf" }));
 *   // later: unsub();
 */
export class TransferManager {
	private queue = new TransferQueue(3);
	private tasks = new Map<string, TransferTask>();
	private fetchFn: TransferFetch;
	private listeners: TransferListenerMap = {};

	/**
	 * @param client - Optional HttpClient instance. If provided, all transfers
	 *                 will use its middleware pipeline (auth, logging, etc.).
	 *                 Falls back to globalThis.fetch if no client is given.
	 */
	constructor(client?: HttpClient) {
		this.fetchFn = client ? createTransferFetch(client) : globalThis.fetch;
	}

	/**
	 * Subscribe to transfer events. Supports multiple listeners per event.
	 * Returns an unsubscribe function that removes only the registered callbacks.
	 */
	on(events: Partial<TransferEvents>): () => void {
		const unsubs: Array<() => void> = [];

		for (const [key, fn] of Object.entries(events)) {
			if (typeof fn !== "function") continue;
			const eventKey = key as keyof TransferEvents;
			if (!this.listeners[eventKey]) {
				this.listeners[eventKey] = [];
			}
			(this.listeners[eventKey] as Array<(...a: unknown[]) => void>).push(
				fn as (...a: unknown[]) => void,
			);
			unsubs.push(() => {
				const arr = this.listeners[eventKey];
				if (arr) {
					const idx = arr.indexOf(fn as never);
					if (idx !== -1) arr.splice(idx, 1);
				}
			});
		}

		return () => unsubs.forEach((u) => u());
	}

	/** Emit an event to all registered listeners. */
	emit<K extends keyof TransferEvents>(
		event: K,
		...args: Parameters<NonNullable<TransferEvents[K]>>
	) {
		const fns = this.listeners[event];

		if (!fns) return;
		for (const fn of fns) {
			(fn as (...a: unknown[]) => void)(...args);
		}
	}

	/** Returns the fetch function used by transfer tasks (with middleware applied). */
	getFetch(): TransferFetch {
		return this.fetchFn;
	}

	/**
	 * Enqueues a transfer task. The task is wrapped to emit events
	 * on status change, progress, completion, and errors.
	 */
	add(task: TransferTask) {
		this.tasks.set(task.id, task);

		// Intercept onProgress to forward progress events to manager listeners
		const originalOnProgress = task.onProgress;
		task.onProgress = (progress) => {
			originalOnProgress?.(progress);
			this.emit("onProgress", progress);
		};

		task.init = (meta) => {
			this.emit("onInit", meta);
		};

		// Wrap the task's start() to emit lifecycle events
		const wrappedTask: TransferTask = {
			...task,
			start: async () => {
				this.emit("onStatusChange", {
					id: task.id,
					type: task.type,
					status: "running",
				});

				try {
					const result = await task.start();
					this.emit("onStatusChange", {
						id: task.id,
						type: task.type,
						status: "completed",
					});
					this.emit("onComplete", {
						id: task.id,
						type: task.type,
						data: result,
					});
					return result;
				} catch (err) {
					const status = task.status === "canceled" ? "canceled" : "failed";
					this.emit("onStatusChange", {
						id: task.id,
						type: task.type,
						status,
					});
					this.emit("onError", {
						id: task.id,
						type: task.type,
						error: err as Error,
					});
					throw err;
				}
			},
		};

		this.queue.add(wrappedTask);
	}

	/** Aborts and removes a transfer by ID. */
	cancel(id: string) {
		this.tasks.get(id)?.cancel();
		this.tasks.delete(id);
	}

	/** Pauses a transfer by ID. */
	pause(id: string) {
		const task = this.tasks.get(id);

		if (task?.pause) {
			task.pause();
			this.emit("onStatusChange", {
				id: task.id,
				type: task.type,
				status: "paused",
			});
		}
	}

	/** Resumes a paused transfer by re-enqueuing it. */
	resume(id: string) {
		const task = this.tasks.get(id);
		if (task && task.status === "paused") {
			this.emit("onStatusChange", {
				id: task.id,
				type: task.type,
				status: "running",
			});
			// Re-enqueue the task
			this.tasks.delete(id);
			this.add(task);
		}
	}

	/** Returns a specific transfer task by ID. */
	getTask(id: string) {
		return this.tasks.get(id);
	}

	/** Returns all active transfer tasks (queued + running). */
	getAllTasks(): TransferTask[] {
		return Array.from(this.tasks.values());
	}
}

/** Factory function to create a new TransferManager instance. */
export function createTransferManager(client?: HttpClient): TransferManager {
	return new TransferManager(client);
}
