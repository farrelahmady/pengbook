"use client";

import { toast } from "sonner";
import { useTransferManager } from "@/transfer/hooks/use-transfer-manager";
import { useEffect } from "react";
import { XCircleIcon, PauseIcon, PlayIcon } from "@phosphor-icons/react";

function formatBytes(bytes: number): string {
	if (bytes === 0) return "0 B";
	const k = 1024;
	const sizes = ["B", "KB", "MB", "GB"];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function formatSpeed(bytesPerSec: number): string {
	return `${formatBytes(bytesPerSec)}/s`;
}

function formatEta(seconds: number): string {
	if (seconds < 60) return `${Math.round(seconds)}s`;
	const mins = Math.floor(seconds / 60);
	const secs = Math.round(seconds % 60);
	return `${mins}m ${secs}s`;
}

/**
 * Renders no UI — only manages toast notifications for active transfers.
 * Uses transfer ID directly as toast ID to prevent duplicates.
 */
export function TransferProgressToast() {
	const { transfers, cancel, pause, resume } = useTransferManager();

	const toastCancel = (p: any) => {
		toast.warning(`${p.label} dibatalkan`);
	};

	useEffect(() => {
		for (const t of transfers) {
			const action = t.type === "download" ? "Download" : "Upload";
			const name = t.filename || "file";
			const label = `${action} ${name}`;

			// Finished states — dismiss the loading toast with a final message
			if (t.status === "completed") {
				toast.success(`${label} selesai`, {
					id: t.id,
					description: t.progress ? formatBytes(t.progress.loaded) : undefined,
					cancel: undefined,
				});
				continue;
			}

			if (t.status === "failed") {
				toast.error(`${label} gagal`, { id: t.id });
				continue;
			}

			if (t.status === "canceled") {
				continue;
			}

			if (t.status === "paused") {
				const p = t.progress;
				const loaded = p?.loaded ?? 0;
				const total = p?.total;

				toast.loading(`${label} dijeda`, {
					id: t.id,
					description: total
						? `${formatBytes(loaded)} / ${formatBytes(total)}`
						: formatBytes(loaded),
					action: {
						label: <PlayIcon className="size-4" />,
						onClick: () => resume(t.id),
					},
					cancel: {
						label: <XCircleIcon className="size-4" />,
						onClick: () => cancel(t.id),
					},
				});
				continue;
			}

			// Skip queued — only show toast once the transfer actually starts
			if (t.status === "queued") continue;

			// Running — show / update the progress toast
			const p = t.progress;
			const percent = p?.percent ?? 0;
			const loaded = p?.loaded ?? 0;
			const total = p?.total;
			const speed = p?.speed;
			const eta = p?.eta;

			const lines: string[] = [];
			if (total) {
				lines.push(`${formatBytes(loaded)} / ${formatBytes(total)}`);
			} else {
				lines.push(formatBytes(loaded));
			}
			if (speed) lines.push(formatSpeed(speed));
			if (eta) lines.push(`ETA: ${formatEta(eta)}`);

			const description = lines.join(" · ");

			// Using t.id as toast ID — sonner will create or update automatically
			toast.loading(`${label}`, {
				id: t.id,
				description: `${percent}%${description ? ` · ${description}` : ""}`,
				cancel: {
					label: <XCircleIcon className="size-4" />,
					onClick: (e) => {
						if (t.status === "running") {
							toastCancel({ label, id: t.id });
							cancel(t.id);
						}
					},
				},
			});
		}
	}, [transfers, cancel, pause, resume]);

	return null;
}
