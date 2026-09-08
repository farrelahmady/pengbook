import { NextRequest } from "next/server";

const CHUNK_SIZE = 64 * 1024; // 64KB
const STREAM_DELAY_MS = 50; // base delay per chunk

interface DownloadBody {
	/** File size in MB (1-500). */
	size?: number;
	/** Download filename. */
	filename?: string;
}

/**
 * GET  /api/dummy-download?size=<MB>&filename=<name>
 * POST /api/dummy-download  { "size": 10, "filename": "file.bin" }
 *
 * Streams a dummy file of the requested size.
 * Supports both GET (query params) and POST (JSON body) for flexibility.
 */
export async function GET(req: NextRequest) {
	const { searchParams } = req.nextUrl;
	const sizeMB = Math.min(
		Math.max(Number(searchParams.get("size")) || 10, 1),
		500,
	);
	const timestamp = new Date()
		.toISOString()
		.replace(/[^0-9]/g, "")
		.slice(0, -3);
	const filename =
		timestamp + "-" + (searchParams.get("filename") || "dummy-file.bin");

	return streamFile(sizeMB, filename);
}

export async function POST(req: NextRequest) {
	let body: DownloadBody = {};
	try {
		body = await req.json();
	} catch {
		// fallback to empty body
	}

	const sizeMB = Math.min(Math.max(Number(body.size) || 10, 1), 500);
	const timestamp = new Date()
		.toISOString()
		.replace(/[^0-9]/g, "")
		.slice(0, -3);
	const filename = timestamp + "-" + (body.filename || "dummy-file.bin");

	return streamFile(sizeMB, filename);
}

function streamFile(sizeMB: number, filename: string) {
	const totalBytes = sizeMB * 1024 * 1024;
	const encoder = new TextEncoder();
	let loaded = 0;

	const stream = new ReadableStream({
		async pull(controller) {
			if (loaded >= totalBytes) {
				controller.close();
				return;
			}

			const remaining = totalBytes - loaded;
			const chunkSize = Math.min(CHUNK_SIZE, remaining);
			const chunk = encoder.encode("A".repeat(chunkSize));

			controller.enqueue(chunk);
			loaded += chunkSize;

			// Simulate network latency (50-150ms per chunk)
			const jitter = Math.random() * 100;
			await new Promise((r) => setTimeout(r, STREAM_DELAY_MS + jitter));
		},
	});

	return new Response(stream, {
		status: 200,
		headers: {
			"Content-Type": "application/octet-stream",
			"Content-Length": String(totalBytes),
			"Content-Disposition": `attachment; filename="${filename}"`,
			"Cache-Control": "no-cache",
		},
	});
}
