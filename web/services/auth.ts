import { httpClient, authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import {
	ApiResponse,
	User,
	TokenResponse,
	LoginRequest,
	RegisterRequest,
} from "@/types";

const logger = createLogger("auth.service");

/**
 * Auth Service
 *
 * Uses two different HTTP clients:
 * - httpClient(): For unauthenticated endpoints (login, register, refresh, logout)
 * - authHttpClient(): For authenticated endpoints (me)
 */
export const authService = {
	/**
	 * Register a new user.
	 * No auth required — uses httpClient (singleton).
	 */
	register: async (data: RegisterRequest): Promise<TokenResponse> => {
		logger.debug("Registering new user", { username: data.username, email: data.email });
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/register",
			data,
		);
		logger.info("User registered successfully");
		return res.data.data;
	},

	/**
	 * Login with email and password.
	 * No auth required — uses httpClient (singleton).
	 */
	login: async (data: LoginRequest): Promise<TokenResponse> => {
		logger.debug("Logging in user", { identifier: data.identifier });
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/login",
			data,
		);
		logger.info("User logged in successfully");
		return res.data.data;
	},

	/**
	 * Refresh access token using a refresh token.
	 * No auth required — uses httpClient (singleton).
	 */
	refresh: async (refreshToken: string): Promise<TokenResponse> => {
		logger.debug("Refreshing access token");
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/refresh",
			{ refreshToken },
		);
		logger.info("Access token refreshed successfully");
		return res.data.data;
	},

	/**
	 * Logout (revoke refresh token).
	 * No auth required — uses httpClient (singleton).
	 */
	logout: async (refreshToken: string): Promise<void> => {
		logger.debug("Revoking refresh token");
		const client = httpClient();
		await client.post<ApiResponse<{ message: string }>>(
			"/api/v1/auth/logout",
			{ refreshToken },
		);
		logger.info("Refresh token revoked");
	},

	/**
	 * Get current user profile.
	 * Auth required — uses authHttpClient (singleton with auth).
	 * Token is automatically added via global token provider.
	 */
	me: async (): Promise<User> => {
		logger.debug("Fetching current user profile");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<User>>("/api/v1/auth/me");
		logger.info("User profile fetched", { userId: res.data.data.id });
		return res.data.data;
	},
};
