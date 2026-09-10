import { createHttpClient } from "@/lib/http-client";
import {
	ApiResponse,
	User,
	TokenResponse,
	LoginRequest,
	RegisterRequest,
} from "@/types";

function authClient(getToken?: () => Promise<string | null>) {
	return createHttpClient(getToken);
}

export const authService = {
	/**
	 * Register a new user.
	 */
	register: async (
		data: RegisterRequest,
		getToken?: () => Promise<string | null>,
	): Promise<TokenResponse> => {
		const client = authClient(getToken || (async () => null));
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/register",
			data,
		);
		return res.data.data;
	},

	/**
	 * Login with email and password.
	 */
	login: async (data: LoginRequest): Promise<TokenResponse> => {
		const client = authClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/login",
			data,
		);
		return res.data.data;
	},

	/**
	 * Refresh access token using a refresh token.
	 */
	refresh: async (refreshToken: string): Promise<TokenResponse> => {
		const client = authClient();
		const res = await client.post<ApiResponse<TokenResponse>>(
			"/api/v1/auth/refresh",
			{ refreshToken },
		);
		return res.data.data;
	},

	/**
	 * Logout (revoke refresh token).
	 */
	logout: async (refreshToken: string): Promise<void> => {
		const client = authClient();
		await client.post<ApiResponse<{ message: string }>>("/api/v1/auth/logout", {
			refreshToken,
		});
	},

	/**
	 * Get current user profile (requires access token).
	 */
	me: async (getToken: () => Promise<string | null>): Promise<User> => {
		const client = authClient(getToken);
		const res = await client.get<ApiResponse<User>>("/api/v1/auth/me");
		return res.data.data;
	},
};
