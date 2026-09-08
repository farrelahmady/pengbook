import { TransferProgress } from "./types";

/**
 * Creates a progress tracker that computes transfer metrics (speed, ETA, percent)
 * and calls the onUpdate callback with the computed TransferProgress.
 *
 * The tracker maintains internal state (last time, last bytes) to calculate
 * speed and ETA between updates.
 *
 * Usage:
 *   const track = createProgressTracker((p) => updateUI(p));
 *   track(id, loaded, total); // called on each chunk received
 */
export function createProgressTracker(
	onUpdate?: (p: TransferProgress) => void,
) {
	let lastTime = Date.now();
	let lastLoaded = 0;

	return (id: string, loaded: number, total?: number) => {
		const now = Date.now();
		const timeDiff = (now - lastTime) / 1000;

		// Bytes transferred since last update, divided by time elapsed
		const speed = timeDiff > 0 ? (loaded - lastLoaded) / timeDiff : 0;

		// Estimated seconds remaining based on current speed
		const eta = total && speed > 0 ? (total - loaded) / speed : undefined;

		// Rounded percentage (0-100)
		const percent = total ? Math.round((loaded / total) * 100) : 0;

		// Update state for next calculation
		lastTime = now;
		lastLoaded = loaded;

		onUpdate?.({
			id,
			percent,
			loaded,
			total,
			speed,
			eta,
		});
	};
}
