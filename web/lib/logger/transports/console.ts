/**
 * Console Transport
 *
 * Outputs log messages to the browser console with color coding.
 * Used primarily in development environment.
 */

import { LogTransport, LogLevel, LogContext } from "../types";

/** Color mapping for each log level */
const LEVEL_CONFIG: Record<LogLevel, { color: string; icon: string }> = {
	debug: { color: "#6B7280", icon: "🔍" },
	info: { color: "#3B82F6", icon: "ℹ️" },
	warn: { color: "#F59E0B", icon: "⚠️" },
	error: { color: "#EF4444", icon: "❌" },
};

/**
 * Format timestamp for log output
 */
function formatTimestamp(date: Date): string {
	return date.toISOString().replace("T", " ").replace("Z", "");
}

/**
 * Console transport for development logging.
 *
 * Features:
 * - Color-coded output by log level
 * - Timestamp formatting
 * - Structured context display
 * - Uses appropriate console method (debug, info, warn, error)
 */
export class ConsoleTransport implements LogTransport {
	log(level: LogLevel, message: string, context?: LogContext): void {
		const config = LEVEL_CONFIG[level];
		const timestamp = formatTimestamp(new Date());
		const prefix = `${config.icon} [${level.toUpperCase()}]`;
		const hasContext = context && Object.keys(context).length > 0;

		const styledMessage = `%c${prefix}%c ${timestamp} %c${message}`;
		const styles = [
			`color: ${config.color}; font-weight: bold`,
			"color: #9CA3AF",
			"color: inherit",
		];

		// Use appropriate console method with proper typing
		const consoleMethod = (
			console[level] as (...args: unknown[]) => void
		);

		if (hasContext) {
			consoleMethod(styledMessage, ...styles, context);
		} else {
			consoleMethod(styledMessage, ...styles);
		}
	}
}
