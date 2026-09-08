"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
	TransferManager,
	createTransferManager,
} from "../core/transfer-manager";
import { TransferProgress, TransferStatus, TransferTask } from "../core/types";
import { HttpClient } from "@/http/core/http-client";

/**
 * Internal state representation for a transfer task.
 * Used by the hook to track progress and status for UI rendering.
 */
interface TransferState {
	id: string;
	type: "download" | "upload";
	status: TransferStatus;
	progress?: TransferProgress;
	result?: unknown;
	filename?: string;
}

/**
 * Singleton TransferManager instance shared across all hook consumers.
 * Ensures one manager handles all transfers, regardless of which component
 * initiated them (e.g., ProgressToast can observe transfers from anywhere).
 */
let _manager: TransferManager | null = null;

function getManager(client?: HttpClient): TransferManager {
	if (!_manager) {
		_manager = createTransferManager(client);
	}
	return _manager;
}

/**
 * React hook for interacting with the TransferManager.
 *
 * Provides:
 * - `add(task)`: Enqueue a download/upload task
 * - `cancel(id)`: Abort a transfer by ID
 * - `transfers`: Current list of active transfers with status and progress
 *
 * Automatically subscribes to TransferManager events and updates React state.
 * Completed transfers are removed after 2s, failed transfers after 5s.
 *
 * Usage:
 *   const { add, cancel, transfers } = useTransferManager(httpClient);
 *   add(createDownloadTask({ id: "1", url: "/api/file.pdf", onProgress: ... }));
 */
export function useTransferManager(client?: HttpClient) {
	const managerRef = useRef<TransferManager>(getClient(client));
	const [transfers, setTransfers] = useState<Map<string, TransferState>>(
		new Map(),
	);

	function getClient(c?: HttpClient): TransferManager {
		return getManager(c);
	}

	// Subscribe to TransferManager events on mount, clean up on unmount
	useEffect(() => {
		const manager = managerRef.current;

		const unsub = manager.on({
			onInit: ({ id, type, status, filename }) => {
				setTransfers((prev) => {
					const next = new Map(prev);
					const existing = next.get(id);
					next.set(id, {
						...(existing || { id, type, status }),
						filename: filename,
					});
					return next;
				});
			},
			onStatusChange: ({ id, type, status }) => {
				setTransfers((prev) => {
					const next = new Map(prev);
					const existing = next.get(id);
					next.set(id, {
						id,
						type,
						status,
						progress: existing?.progress,
						result: existing?.result,
						filename: existing?.filename,
					});
					return next;
				});
			},
			onProgress: (progress) => {
				setTransfers((prev) => {
					const next = new Map(prev);
					const existing = next.get(progress.id);

					if (existing) {
						next.set(progress.id, {
							...existing,
							progress,
						});
					}
					return next;
				});
			},
			onComplete: ({ id, data }) => {
				// Store result so consumers can access the completed data (e.g., Blob)
				setTransfers((prev) => {
					const next = new Map(prev);
					const existing = next.get(id);

					if (existing) {
						next.set(id, {
							...existing,
							result: data,
						});
					}
					return next;
				});
				// Keep completed transfer visible briefly before removing
				setTimeout(() => {
					setTransfers((prev) => {
						const next = new Map(prev);
						next.delete(id);
						return next;
					});
				}, 2000);
			},
			onError: ({ id }) => {
				// Keep failed transfer visible longer for user to see the error
				setTransfers((prev) => {
					const next = new Map(prev);
					next.delete(id);
					return next;
				});
			},
		});

		return unsub;
	}, []);

	/** Enqueue a transfer task and track its state in React. */
	const add = useCallback((task: TransferTask) => {
		setTransfers((prev) => {
			const next = new Map(prev);
			next.set(task.id, {
				id: task.id,
				type: task.type,
				status: "queued",
				filename: task.filename,
			});
			return next;
		});
		managerRef.current.add(task);
	}, []);

	/** Abort a transfer and remove it from tracking. */
	const cancel = useCallback((id: string) => {
		managerRef.current.cancel(id);
		setTransfers((prev) => {
			const next = new Map(prev);
			next.delete(id);
			return next;
		});
	}, []);

	/** Pause a transfer. */
	const pause = useCallback((id: string) => {
		managerRef.current.pause(id);
	}, []);

	/** Resume a paused transfer. */
	const resume = useCallback((id: string) => {
		managerRef.current.resume(id);
	}, []);

	return {
		add,
		cancel,
		pause,
		resume,
		transfers: Array.from(transfers.values()),
	};
}
