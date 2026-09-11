import { HttpRequestConfig } from "../types/http";
import { getTokenProvider } from "@/lib/token-provider";

/**
 * Auth Interceptor Configuration
 */
export interface AuthInterceptorOptions {
	/**
	 * Getter function that returns the onRefreshFailed handler.
	 * Uses lazy evaluation to avoid timing issues where the handler
	 * is captured before it's set by AuthProvider.
	 *
	 * @returns The current handler, or null if not set
	 */
	getOnRefreshFailed?: () => (() => Promise<void>) | null;
}

/**
 * Refresh Lock State
 *
 * Handles concurrent 401 responses by ensuring only one refresh
 * happens at a time. All other requests wait for the same refresh
 * promise to resolve.
 */
interface RefreshLockState {
	isRefreshing: boolean;
	refreshPromise: Promise<string | null> | null;
	waitQueue: Array<{
		resolve: (token: string | null) => void;
	}>;
}

/**
 * Creates a Refresh Lock instance.
 *
 * This pattern is used to handle multiple concurrent 401 responses.
 * Only one refresh happens, and all other requests wait for it.
 */
function createRefreshLock(): RefreshLockState {
	return {
		isRefreshing: false,
		refreshPromise: null,
		waitQueue: [],
	};
}

/**
 * Waits for an ongoing refresh to complete.
 *
 * @param lock - The refresh lock state
 * @returns Promise that resolves with the new token or null if refresh failed
 */
function waitForRefresh(lock: RefreshLockState): Promise<string | null> {
	return new Promise((resolve) => {
		lock.waitQueue.push({ resolve });
	});
}

/**
 * Notifies all waiting requests about the refresh result.
 *
 * @param lock - The refresh lock state
 * @param token - The new token, or null if refresh failed
 */
function notifyWaiters(lock: RefreshLockState, token: string | null): void {
	for (const waiter of lock.waitQueue) {
		waiter.resolve(token);
	}
	lock.waitQueue = [];
}

/**
 * Auth Interceptor
 *
 * Handles 401 responses by automatically refreshing the token and retrying
 * the request. Uses a refresh lock to ensure only one refresh happens
 * at a time, even with multiple concurrent 401 responses.
 *
 * @example
 * ```typescript
 * import { authInterceptor } from "@/http/interceptors/auth-interceptor";
 * import { HttpClient } from "@/http/core/http-client";
 *
 * const client = new HttpClient(baseUrl, [
 *   authMiddleware(getToken),
 * ], {
 *   interceptors: [
 *     authInterceptor({
 *       onRefreshFailed: async () => {
 *         router.push("/login");
 *       },
 *     }),
 *   ],
 * });
 * ```
 */
export function authInterceptor(options: AuthInterceptorOptions = {}) {
	const { getOnRefreshFailed } = options;

	// Shared refresh lock across all requests
	const lock = createRefreshLock();

	return {
		/**
		 * Response interceptor that handles 401 Unauthorized responses.
		 *
		 * Flow:
		 * 1. If response is not 401, pass through
		 * 2. If refresh is already in progress, wait for it
		 * 3. If refresh is not in progress, start a new one
		 * 4. After refresh, retry the request with new token
		 * 5. If refresh fails, call onRefreshFailed callback
		 */
		onResponse: async (
			response: Response,
			config: HttpRequestConfig,
		): Promise<Response | null> => {
			// Not 401? Pass through
			if (response.status !== 401) {
				return response;
			}

			// Already retried? Prevent infinite loop
			if ((config as any).__retried) {
				return response;
			}

			// No token provider? Can't refresh
			const getToken = getTokenProvider();
			if (!getToken) {
				return response;
			}

			// Lazy read: get the handler at call time, not at construction time
			const onRefreshFailed = getOnRefreshFailed?.() ?? undefined;

			// ── Race Condition Handler ──────────────────────

			// If refresh is already in progress, wait for it
			if (lock.isRefreshing && lock.refreshPromise) {
				console.log("[AuthInterceptor] Waiting for existing refresh...");

				try {
					const token = await waitForRefresh(lock);

					if (!token) {
						// Refresh failed
						await onRefreshFailed?.();
						return response;
					}

					// Retry with new token
					return retryWithNewToken(config, token);
				} catch {
					// Refresh failed
					await onRefreshFailed?.();
					return response;
				}
			}

			// ── Start New Refresh ──────────────────────────

			console.log("[AuthInterceptor] Starting new refresh...");

			lock.isRefreshing = true;
			lock.refreshPromise = getToken();

			try {
				const token = await lock.refreshPromise;

				if (!token) {
					// Refresh failed — notify all waiters
					notifyWaiters(lock, null);
					await onRefreshFailed?.();
					return response;
				}

				// Refresh successful — notify all waiters
				notifyWaiters(lock, token);

				// Retry with new token
				return retryWithNewToken(config, token);
			} catch {
				// Refresh error — notify all waiters
				notifyWaiters(lock, null);
				await onRefreshFailed?.();
				return response;
			} finally {
				lock.isRefreshing = false;
				lock.refreshPromise = null;
			}
		},
	};
}

/**
 * Updates the request config with a new token and marks it as retried.
 *
 * @param config - The original request config
 * @param token - The new authentication token
 * @returns null to signal retry
 */
function retryWithNewToken(
	config: HttpRequestConfig,
	token: string,
): null {
	// Update config with new token
	const headers = new Headers(config.headers);
	headers.set("Authorization", `Bearer ${token}`);

	// Mark as retried to prevent infinite loop
	(config as any).headers = headers;
	(config as any).__retried = true;

	// Return null to signal retry
	return null;
}
