import createMiddleware from "next-intl/middleware";
import { NextRequest, NextResponse } from "next/server";
import { routing } from "./i18n/routing";

/**
 * Proxy: Security Layer (Next.js 16)
 * - Handle locale detection (next-intl)
 * - Basic auth check (redirect if no token)
 * - Protect guest/auth routes
 */

const intlMiddleware = createMiddleware(routing);

// ── Route Definitions ──────────────────────────────────────

/** Guest-only routes: accessible only when NOT logged in */
const GUEST_ROUTES = ["/login", "/register"];

/** Auth-only routes: accessible only when logged in */
const AUTH_ROUTES = ["/jurnal", "/aset", "/laporan", "/akun", "/download"];

// ── Helpers ────────────────────────────────────────────────

/** Strip locale prefix from pathname */
function stripLocale(pathname: string): string {
	return pathname.replace(/^\/(?:en|id)(\/|$)/, "$1");
}

/** Check if pathname is a guest-only route */
function isGuestRoute(pathname: string): boolean {
	const clean = stripLocale(pathname);
	return GUEST_ROUTES.some((r) => clean.startsWith(r));
}

/** Check if pathname is an auth-only route */
function isAuthRoute(pathname: string): boolean {
	const clean = stripLocale(pathname);
	return AUTH_ROUTES.some((r) => clean.startsWith(r));
}

/** Extract locale from pathname */
function getLocale(pathname: string): string {
	const match = pathname.match(/^\/(en|id)(\/|$)/);
	return match ? match[1] : routing.defaultLocale;
}

/** Check if user has access token */
function hasToken(request: NextRequest): boolean {
	return !!request.cookies.get("access_token")?.value;
}

// ── Proxy ──────────────────────────────────────────────────

export default function proxy(request: NextRequest) {
	const { pathname } = request.nextUrl;
	const isLoggedIn = hasToken(request);
	const locale = getLocale(pathname);

	// Guest-only routes: redirect to /jurnal if already logged in
	if (isGuestRoute(pathname)) {
		if (isLoggedIn) {
			return NextResponse.redirect(new URL(`/${locale}/jurnal`, request.url));
		}
		// Not logged in → allow access, continue to intlMiddleware
	}

	// Auth-only routes: redirect to /login if not logged in
	if (isAuthRoute(pathname)) {
		if (!isLoggedIn) {
			const loginUrl = new URL(`/${locale}/login`, request.url);
			loginUrl.searchParams.set("redirect", pathname);
			return NextResponse.redirect(loginUrl);
		}
		// Logged in → allow access, continue to intlMiddleware
	}

	// Run intlMiddleware for locale handling
	return intlMiddleware(request);
}

export const config = {
	matcher: ["/((?!api|trpc|_next|_vercel|.*\\..*).*)"],
};
