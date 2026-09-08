import { TransferTask } from "./types";

/**
 * A concurrency-limited task queue for managing parallel transfers.
 *
 * Tasks are executed in FIFO order. When a task completes (or fails),
 * the next task in the queue starts automatically.
 *
 * Example with concurrency=3:
 *   [task1, task2, task3] run in parallel
 *   When task1 finishes → task4 starts
 *   When task2 finishes → task5 starts
 *   etc.
 */
export class TransferQueue {
	private queue: TransferTask[] = [];
	private running = 0;

	/**
	 * @param concurrency - Maximum number of tasks running in parallel (default: 3).
	 */
	constructor(private concurrency = 3) {}

	/** Enqueues a task and starts it if under the concurrency limit. */
	add(task: TransferTask) {
		this.queue.push(task);
		this.runNext();
	}

	/** Starts the next queued task if under the concurrency limit. */
	private runNext() {
		if (this.running >= this.concurrency) return;

		const task = this.queue.shift();
		if (!task) return;

		this.running++;

		// Errors are caught silently — the task's own error handling
		// (via TransferManager.emit("onError")) takes care of notifications.
		task
			.start()
			.catch(() => {})
			.finally(() => {
				this.running--;
				this.runNext();
			});
	}
}
