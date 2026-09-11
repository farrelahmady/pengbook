/**
 * Logger Implementation
 *
 * Core logger class with level filtering, context enrichment,
 * and child logger support.
 */

import { Logger, LogLevel, LogContext, LogTransport } from "./types";

/** Log level hierarchy for filtering */
const LOG_LEVELS: LogLevel[] = ["debug", "info", "warn", "error"];

/**
 * Check if a message should be logged based on minimum level
 */
function shouldLog(currentLevel: LogLevel, messageLevel: LogLevel): boolean {
	return LOG_LEVELS.indexOf(currentLevel) <= LOG_LEVELS.indexOf(messageLevel);
}

/**
 * Core logger implementation
 */
class LoggerImpl implements Logger {
	private context: LogContext = {};
	private moduleName: string;

	constructor(
		private transport: LogTransport,
		private minLevel: LogLevel = "debug",
		moduleName: string = "app",
	) {
		this.moduleName = moduleName;
	}

	debug(message: string, context?: LogContext): void {
		this.log("debug", message, context);
	}

	info(message: string, context?: LogContext): void {
		this.log("info", message, context);
	}

	warn(message: string, context?: LogContext): void {
		this.log("warn", message, context);
	}

	error(message: string, context?: LogContext): void {
		this.log("error", message, context);
	}

	/**
	 * Create a child logger with a sub-module name.
	 * Inherits parent's context and transport.
	 *
	 * @param module - Sub-module name (appended to parent module)
	 * @returns New Logger instance with sub-module name
	 *
	 * @example
	 * ```typescript
	 * const logger = createLogger("journal");
	 * const serviceLogger = logger.child("service");
	 * // Logs will show module: "journal:service"
	 * ```
	 */
	child(module: string): Logger {
		const childLogger = new LoggerImpl(
			this.transport,
			this.minLevel,
			`${this.moduleName}:${module}`,
		);
		childLogger.context = { ...this.context };
		return childLogger;
	}

	/**
	 * Internal log method with level filtering and context enrichment
	 */
	private log(
		level: LogLevel,
		message: string,
		context?: LogContext,
	): void {
		if (!shouldLog(this.minLevel, level)) {
			return;
		}

		const enrichedContext: LogContext = {
			...this.context,
			...context,
			module: this.moduleName,
		};

		this.transport.log(level, message, enrichedContext);
	}
}

/**
 * Create a new Logger instance
 *
 * @param transport - Transport for outputting logs
 * @param minLevel - Minimum log level to output
 * @param moduleName - Module name for context
 * @returns New Logger instance
 */
export function createLoggerInstance(
	transport: LogTransport,
	minLevel: LogLevel,
	moduleName: string,
): Logger {
	return new LoggerImpl(transport, minLevel, moduleName);
}
