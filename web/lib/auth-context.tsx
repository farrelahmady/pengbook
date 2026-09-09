"use client";

import {
	createContext,
	useContext,
	useState,
	useEffect,
	useCallback,
	useMemo,
	type ReactNode,
} from "react";
import { useRouter } from "@/i18n/navigation";
import { authService } from "@/services/auth";
import type { User, TokenResponse } from "@/types";

// ── Cookie helpers ──────────────────────────────────────────

function setCookie(name: string, value: string, maxAge: number) {
	document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAge}; SameSite=Lax`;
}

function getCookie(name: string): string | null {
	const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
	return match ? decodeURIComponent(match[1]) : null;
}

function deleteCookie(name: string) {
	document.cookie = `${name}=; path=/; max-age=0; SameSite=Lax`;
}

// ── Token storage (cookies) ────────────────────────────────

const ACCESS_TOKEN_KEY = "access_token";
const REFRESH_TOKEN_KEY = "refresh_token";

const ACCESS_TOKEN_MAX_AGE = 15 * 60; // 15 minutes
const REFRESH_TOKEN_MAX_AGE = 7 * 24 * 60 * 60; // 7 days

/** Refresh token if less than this many ms left before expiry */
const REFRESH_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes

function getAccessToken(): string | null {
	if (typeof window === "undefined") return null;
	return getCookie(ACCESS_TOKEN_KEY);
}

function getRefreshToken(): string | null {
	if (typeof window === "undefined") return null;
	return getCookie(REFRESH_TOKEN_KEY);
}

function setTokens(access: string, refresh: string) {
	setCookie(ACCESS_TOKEN_KEY, access, ACCESS_TOKEN_MAX_AGE);
	setCookie(REFRESH_TOKEN_KEY, refresh, REFRESH_TOKEN_MAX_AGE);
}

function clearTokens() {
	deleteCookie(ACCESS_TOKEN_KEY);
	deleteCookie(REFRESH_TOKEN_KEY);
}

// ── Token expiry helpers ────────────────────────────────────

/** Decode JWT payload (no signature verification — client-side only) */
function decodeTokenPayload(token: string): Record<string, unknown> | null {
	try {
		const base64 = token.split(".")[1];
		const json = atob(base64.replace(/-/g, "+").replace(/_/g, "/"));
		return JSON.parse(json);
	} catch {
		return null;
	}
}

/** Returns ms until token expires, or 0 if expired / unparseable */
function getTokenExpiresIn(token: string): number {
	const payload = decodeTokenPayload(token);
	if (!payload || typeof payload.exp !== "number") return 0;
	return payload.exp * 1000 - Date.now();
}

/** True if token is expired or will expire within REFRESH_THRESHOLD_MS */
function isTokenExpiringSoon(token: string): boolean {
	const expiresIn = getTokenExpiresIn(token);
	return expiresIn <= REFRESH_THRESHOLD_MS;
}

// ── Context types ───────────────────────────────────────────

interface AuthContextValue {
	user: User | null;
	isLoading: boolean;
	isAuthenticated: boolean;
	getToken: () => Promise<string | null>;
	login: (identifier: string, password: string) => Promise<void>;
	register: (
		name: string,
		username: string,
		email: string,
		password: string,
	) => Promise<void>;
	logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

// ── Provider ────────────────────────────────────────────────

export function AuthProvider({ children }: { children: ReactNode }) {
	const [user, setUser] = useState<User | null>(null);
	const [isLoading, setIsLoading] = useState(true);
	const router = useRouter();

	// ── Core helpers ───────────────────────────────────────

	/** Fetch current user profile */
	const fetchUser = useCallback(async (token: string): Promise<User | null> => {
		try {
			return await authService.me(async () => token);
		} catch {
			return null;
		}
	}, []);

	/** Try to refresh access token using the refresh token */
	const tryRefresh = useCallback(async (): Promise<boolean> => {
		const refreshToken = getRefreshToken();
		if (!refreshToken) return false;

		try {
			const tokens = await authService.refresh(refreshToken);
			setTokens(tokens.accessToken, tokens.refreshToken);
			return true;
		} catch {
			clearTokens();
			return false;
		}
	}, []);

	/**
	 * Returns a valid access token.
	 * If the current token is about to expire, refresh first.
	 * Returns null if no token or refresh fails (token is invalid).
	 * This is the main entry-point used by HTTP auth middleware.
	 */
	const getToken = useCallback(async (): Promise<string | null> => {
		let token = getAccessToken();
		if (!token) return null;

		// Token expired or expiring soon → refresh
		if (isTokenExpiringSoon(token)) {
			const refreshed = await tryRefresh();
			if (refreshed) {
				token = getAccessToken(); // Get new token
			} else {
				// Refresh failed → token is invalid, return null
				return null;
			}
		}

		return token;
	}, [tryRefresh]);

	// ── Initialization ─────────────────────────────────────

	useEffect(() => {
		let cancelled = false;

		async function init() {
			const accessToken = getAccessToken();
			if (!accessToken) {
				setIsLoading(false);
				return;
			}

			// Token expired on mount → try refresh immediately
			if (isTokenExpiringSoon(accessToken)) {
				const refreshed = await tryRefresh();
				if (cancelled) return;

				if (!refreshed) {
					clearTokens();
					setIsLoading(false);
					return;
				}
			}

			// Fetch user with valid token
			const token = getAccessToken();
			if (token) {
				const u = await fetchUser(token);
				if (!cancelled) setUser(u);
			}

			if (!cancelled) setIsLoading(false);
		}

		init();
		return () => { cancelled = true; };
	}, [fetchUser, tryRefresh]);

	// ── Proactive refresh (safety-net interval) ────────────

	useEffect(() => {
		// Check every 60 s — if token expires within 5 min, refresh
		const interval = setInterval(async () => {
			const token = getAccessToken();
			if (!token) return;

			if (isTokenExpiringSoon(token)) {
				const refreshed = await tryRefresh();
				if (refreshed) {
					const newToken = getAccessToken();
					if (newToken) {
						const u = await fetchUser(newToken);
						setUser(u);
					}
				}
			}
		}, 60 * 1000);

		return () => clearInterval(interval);
	}, [fetchUser, tryRefresh]);

	// ── Auth actions ───────────────────────────────────────

	const login = useCallback(
		async (identifier: string, password: string) => {
			const tokens = await authService.login({ identifier, password });
			setTokens(tokens.accessToken, tokens.refreshToken);

			const u = await fetchUser(tokens.accessToken);
			setUser(u);

			router.push("/jurnal");
		},
		[fetchUser, router],
	);

	const register = useCallback(
		async (name: string, username: string, email: string, password: string) => {
			const tokens = await authService.register({
				name,
				username,
				email,
				password,
			});
			setTokens(tokens.accessToken, tokens.refreshToken);

			const u = await fetchUser(tokens.accessToken);
			setUser(u);

			router.push("/jurnal");
		},
		[fetchUser, router],
	);

	const logout = useCallback(async () => {
		const refreshToken = getRefreshToken();
		if (refreshToken) {
			try {
				await authService.logout(refreshToken);
			} catch {
				// Ignore logout API errors — clear local state anyway
			}
		}

		clearTokens();
		setUser(null);
		router.push("/login");
	}, [router]);

	// ── Context value ──────────────────────────────────────

	const value = useMemo<AuthContextValue>(
		() => ({
			user,
			isLoading,
			isAuthenticated: !!user,
			getToken,
			login,
			register,
			logout,
		}),
		[user, isLoading, getToken, login, register, logout],
	);

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// ── Hook ────────────────────────────────────────────────────

export function useAuth() {
	const ctx = useContext(AuthContext);
	if (!ctx) {
		throw new Error("useAuth must be used within an AuthProvider");
	}
	return ctx;
}
