import { httpClient, authHttpClient } from "@/lib/http-client";
import {
	ApiResponse,
	User,
	TokenResponse,
	LoginRequest,
	RegisterRequest,
} from "@/types";

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
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/register",
			data,
		);
		return res.data.data;
	},

	/**
	 * Login with email and password.
	 * No auth required — uses httpClient (singleton).
	 */
	login: async (data: LoginRequest): Promise<TokenResponse> => {
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/login",
			data,
		);
		return res.data.data;
	},

	/**
	 * Refresh access token using a refresh token.
	 * No auth required — uses httpClient (singleton).
	 */
	refresh: async (refreshToken: string): Promise<TokenResponse> => {
		const client = httpClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/refresh",
			{ refreshToken },
		);
		return res.data.data;
	},

	/**
	 * Logout (revoke refresh token).
	 * No auth required — uses httpClient (singleton).
	 */
	logout: async (refreshToken: string): Promise<void> => {
		const client = httpClient();
		await client.post<ApiResponse<{ message: string }>>(
			"/api/v1/auth/logout",
			{ refreshToken },
		);
	},

	/**
	 * Get current user profile.
	 * Auth required — uses authHttpClient (singleton with auth).
	 * Token is automatically added via global token provider.
	 */
	me: async (): Promise<User> => {
		const client = authHttpClient();
		const res = await client.get<ApiResponse<User>>("/api/v1/auth/me");
		return res.data.data;
	},
};
