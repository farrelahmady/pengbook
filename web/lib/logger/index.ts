/**
 * Logger Factory & Singleton
 *
 * Provides factory functions for creating loggers and manages
 * the singleton root logger instance.
 *
 * Usage:
 *   import { createLogger } from "@/lib/logger";
 *
 *   const logger = createLogger("journal");
 *   logger.info("Journals fetched", { count: 25 });
 */

import { Logger, LogLevel, LogTransport } from "./types";
import { ConsoleTransport } from "./transports/console";
import { createLoggerInstance } from "./logger";

// ── Transport Selection ──────────────────────────────────

/**
 * Create the appropriate transport based on environment
 */
function createTransport(): LogTransport {
	// Development: console transport with colors
	if (process.env.NODE_ENV === "development") {
		return new ConsoleTransport();
	}

	// Production: console transport (can be extended with remote transports)
	return new ConsoleTransport();
}

// ── Min Level Configuration ──────────────────────────────

/**
 * Get minimum log level from environment or default
 */
function getMinLevel(): LogLevel {
	const level = process.env.NEXT_PUBLIC_LOG_LEVEL as LogLevel;
	const validLevels: LogLevel[] = ["debug", "info", "warn", "error"];

	if (level && validLevels.includes(level)) {
		return level;
	}

	// Development: debug (verbose)
	// Production: info (less verbose)
	return process.env.NODE_ENV === "development" ? "debug" : "info";
}

// ── Singleton Root Logger ────────────────────────────────

let _rootLogger: Logger | null = null;

/**
 * Get or create the root logger singleton
 */
function getRootLogger(): Logger {
	if (!_rootLogger) {
		_rootLogger = createLoggerInstance(createTransport(), getMinLevel(), "app");
	}
	return _rootLogger;
}

// ── Public API ───────────────────────────────────────────

/**
 * Create a logger for a specific module.
 *
 * The logger inherits the root logger's transport and minimum level.
 * Use dot notation for nested modules.
 *
 * @param module - Module name (e.g., "journal", "auth", "http.client")
 * @returns Logger instance with module context
 *
 * @example
 * ```typescript
 * // In services/journal.ts
 * import { createLogger } from "@/lib/logger";
 *
 * const logger = createLogger("journal");
 *
 * export const journalService = {
 *   getAll: async () => {
 *     logger.debug("Fetching journals");
 *     const res = await client.get("/api/journals");
 *     logger.info("Journals fetched", { count: res.data.length });
 *     return res.data;
 *   },
 * };
 * ```
 *
 * @example
 * ```typescript
 * // Child logger for sub-modules
 * const logger = createLogger("journal");
 * const serviceLogger = logger.child("service");
 * // Logs will show module: "journal:service"
 * ```
 */
export function createLogger(module: string): Logger {
	return getRootLogger().child(module);
}

/**
 * Get the root logger (for advanced use cases)
 */
export function getRootLoggerInstance(): Logger {
	return getRootLogger();
}

// Re-export types for convenience
export type { Logger, LogLevel, LogContext, LogTransport } from "./types";
