/**
 * Global token provider for authenticated HTTP requests.
 *
 * This module stores a reference to the getToken function from AuthContext,
 * allowing services to access auth tokens without explicit parameter passing.
 *
 * Usage:
 *   // In AuthProvider (set once):
 *   setTokenProvider(getToken);
 *
 *   // In services (use anywhere):
 *   const client = authHttpClient();
 */

type GetTokenFn = () => Promise<string | null>;

let globalGetToken: GetTokenFn | null = null;

/**
 * Set the global token provider. Called once by AuthProvider.
 */
export function setTokenProvider(provider: GetTokenFn) {
	globalGetToken = provider;
}

/**
 * Get the global token provider. Returns null if not initialized.
 */
export function getTokenProvider(): GetTokenFn | null {
	return globalGetToken;
}
