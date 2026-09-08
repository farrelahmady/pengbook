"use client";

import { useEffect, useRef, useState } from "react";
import { useTransferManager } from "@/transfer/hooks/use-transfer-manager";
import { createDownloadTask } from "@/transfer/task/download-task";
import { httpClient } from "@/lib/http-client";
import { createTransferFetch } from "@/transfer/core/transfer-fetch";

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

function saveBlobAsFile(blob: Blob, filename: string) {
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

function parseHeaders(raw: string): Record<string, string> | undefined {
	const trimmed = raw.trim();
	if (!trimmed) return undefined;
	try {
		const obj = JSON.parse(trimmed);
		if (typeof obj === "object" && obj !== null)
			return obj as Record<string, string>;
		return undefined;
	} catch {
		return undefined;
	}
}

function parseBody(raw: string): unknown {
	const trimmed = raw.trim();
	if (!trimmed) return undefined;
	try {
		return JSON.parse(trimmed);
	} catch {
		return undefined;
	}
}

export function DownloadDemo() {
	const { add, transfers } = useTransferManager();
	const [method, setMethod] = useState<"GET" | "POST">("POST");
	const [fileSizeMB, setFileSizeMB] = useState(1);
	const [filename, setFilename] = useState("dummy-download.bin");
	const [bodyRaw, setBodyRaw] = useState(
		'{\n  "size": 10,\n  "filename": "dummy-download.bin"\n}',
	);
	const [headersRaw, setHeadersRaw] = useState(
		'{\n  "X-Custom-Header": "demo-value"\n}',
	);
	const savedRef = useRef<Set<string>>(new Set());

	const activeDownload = transfers.find((t) => t.type === "download");

	// Auto-save completed downloads
	useEffect(() => {
		for (const t of transfers) {
			if (
				t.status === "completed" &&
				t.result instanceof Blob &&
				!savedRef.current.has(t.id)
			) {
				savedRef.current.add(t.id);
				saveBlobAsFile(t.result, t.filename ?? filename);
			}
		}
	}, [transfers, filename]);

	const handleDownload = () => {
		const id = `dummy-download-${Date.now()}`;
		const client = httpClient();
		const fetchFn = createTransferFetch(client);

		const isGet = method === "GET";
		const url = isGet
			? `/api/dummy-download?size=${fileSizeMB}&filename=${encodeURIComponent(filename)}`
			: "/api/dummy-download";

		const task = createDownloadTask({
			id,
			url,
			method,
			body: isGet ? undefined : parseBody(bodyRaw),
			headers: parseHeaders(headersRaw),
			fetch: fetchFn,
		});

		add(task);
	};

	return (
		<div className="flex flex-col gap-6 p-8 max-w-lg">
			<div>
				<h1 className="text-2xl font-bold mb-2">Download Progress Demo</h1>
				<p className="text-muted-foreground text-sm">
					Streams a file from{" "}
					<code className="text-xs bg-muted px-1 py-0.5 rounded">
						/api/dummy-download
					</code>{" "}
					via HttpClient, then saves it to disk. Supports GET and POST with
					custom body/headers.
				</p>
			</div>

			{/* Method selector */}
			<div className="flex items-center gap-3">
				<label className="text-sm font-medium w-20">Method:</label>
				<div className="flex gap-2">
					{(["GET", "POST"] as const).map((m) => (
						<button
							key={m}
							onClick={() => setMethod(m)}
							className={`px-3 py-1 text-sm rounded border ${
								method === m
									? "bg-primary text-primary-foreground"
									: "bg-secondary text-secondary-foreground hover:bg-secondary/80"
							}`}
						>
							{m}
						</button>
					))}
				</div>
			</div>

			{/* Common config */}
			<div className="flex flex-col gap-3">
				<div className="flex items-center gap-3">
					<label className="text-sm font-medium w-20">File size:</label>
					<select
						value={fileSizeMB}
						onChange={(e) => setFileSizeMB(Number(e.target.value))}
						className="border rounded px-2 py-1 text-sm"
					>
						<option value={1}>1 MB</option>
						<option value={5}>5 MB</option>
						<option value={10}>10 MB</option>
						<option value={50}>50 MB</option>
						<option value={100}>100 MB</option>
					</select>
				</div>
				<div className="flex items-center gap-3">
					<label className="text-sm font-medium w-20">Filename:</label>
					<input
						type="text"
						value={filename}
						onChange={(e) => setFilename(e.target.value)}
						className="border rounded px-2 py-1 text-sm flex-1"
					/>
				</div>
			</div>

			{/* POST body editor */}
			{method === "POST" && (
				<div className="flex flex-col gap-3">
					<div>
						<label className="text-sm font-medium block mb-1">
							Request Body (JSON):
						</label>
						<textarea
							value={bodyRaw}
							onChange={(e) => setBodyRaw(e.target.value)}
							rows={5}
							className="border rounded px-3 py-2 text-sm font-mono w-full resize-y"
							spellCheck={false}
						/>
					</div>
					<div>
						<label className="text-sm font-medium block mb-1">
							Custom Headers (JSON):
						</label>
						<textarea
							value={headersRaw}
							onChange={(e) => setHeadersRaw(e.target.value)}
							rows={3}
							className="border rounded px-3 py-2 text-sm font-mono w-full resize-y"
							spellCheck={false}
						/>
					</div>
				</div>
			)}

			{/* Download button */}
			<button
				onClick={handleDownload}
				disabled={!!activeDownload}
				className="bg-primary text-primary-foreground px-4 py-2 rounded font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{activeDownload ? "Downloading..." : "Start Download"}
			</button>

			{/* Transfer list */}
			{transfers.length > 0 && (
				<div className="flex flex-col gap-4">
					{transfers.map((t) => {
						const p = t.progress;
						const percent = p?.percent ?? 0;
						const loaded = p?.loaded ?? 0;
						const total = p?.total;

						return (
							<div
								key={t.id}
								className="border rounded-lg p-4 flex flex-col gap-3"
							>
								<div className="flex items-center justify-between">
									<span className="text-sm font-medium">{filename}</span>
									<span className="text-xs text-muted-foreground capitalize">
										{t.status}
									</span>
								</div>

								{/* Progress bar */}
								<div className="h-2 w-full rounded-full bg-secondary overflow-hidden">
									<div
										className="h-full rounded-full bg-primary transition-all duration-200"
										style={{ width: `${percent}%` }}
									/>
								</div>

								{/* Stats */}
								<div className="flex items-center justify-between text-xs text-muted-foreground">
									<span>
										{formatBytes(loaded)}
										{total ? ` / ${formatBytes(total)}` : ""}
									</span>
									<span>{percent}%</span>
									{p?.speed ? <span>{formatSpeed(p.speed)}</span> : null}
									{p?.eta ? <span>ETA: {formatEta(p.eta)}</span> : null}
								</div>
							</div>
						);
					})}
				</div>
			)}
		</div>
	);
}
