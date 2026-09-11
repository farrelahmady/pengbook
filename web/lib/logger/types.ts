/**
 * Logger Type Definitions
 *
 * Provides types for the logging system used across the application.
 * Supports multiple log levels, structured context, and pluggable transports.
 */

/** Supported log levels in ascending order of severity */
export type LogLevel = "debug" | "info" | "warn" | "error";

/** Structured context data attached to log messages */
export interface LogContext {
	[key: string]: unknown;
}

/**
 * Transport interface for outputting log messages.
 * Implement this to send logs to different destinations (console, file, remote, etc.)
 */
export interface LogTransport {
	log(level: LogLevel, message: string, context?: LogContext): void;
}

/**
 * Logger interface for structured logging throughout the application.
 *
 * @example
 * ```typescript
 * const logger = createLogger("journal");
 *
 * logger.debug("Fetching journals");
 * logger.info("Journals fetched", { count: 25 });
 * logger.error("Failed to fetch journals", { error });
 *
 * const childLogger = logger.child("service");
 * childLogger.info("Service initialized"); // module: "journal:service"
 * ```
 */
export interface Logger {
	/** Log at debug level (most verbose) */
	debug(message: string, context?: LogContext): void;

	/** Log at info level (normal operations) */
	info(message: string, context?: LogContext): void;

	/** Log at warn level (potential issues) */
	warn(message: string, context?: LogContext): void;

	/** Log at error level (failures) */
	error(message: string, context?: LogContext): void;

	/** Create a child logger with a sub-module name */
	child(module: string): Logger;
}
