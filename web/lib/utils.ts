import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { format, parseISO } from "date-fns";
import { id } from "date-fns/locale";
import { useTranslations } from "next-intl";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export function formatNumber(value: number | string, decimals = 0): string {
	const num = typeof value === "string" ? parseFloat(value) : value;
	if (isNaN(num)) return "—";
	return new Intl.NumberFormat("id-ID", {
		minimumFractionDigits: decimals,
		maximumFractionDigits: decimals,
	}).format(num);
}

export function parseDecimal(
	value: string | number | null | undefined,
): number {
	if (value == null) return 0;
	const num = typeof value === "string" ? parseFloat(value) : value;
	return isNaN(num) ? 0 : num;
}

// ── Date formatter ─────────────────────────────────────────────────
export function formatDate(value: string | Date, fmt = "d MMMM yyyy"): string {
	try {
		const date = typeof value === "string" ? parseISO(value) : value;
		return format(date, fmt, { locale: id });
	} catch {
		return String(value);
	}
}

export function formatDateShort(value: string | Date): string {
	return formatDate(value, "d MMM yyyy");
}

export function formatTime(value: string | Date): string {
	return formatDate(value, "HH:mm");
}

export function formatDateGroup(value: string | Date): string {
	return formatDate(value, "EEEE, d MMMM yyyy");
}

// ── Accounting helpers ─────────────────────────────────────────────
export function parseDecimalSafe(v: string | number | null | undefined) {
	return parseDecimal(v);
}

export function calcJournalTotals(
	lines: Array<{ debit: number | string; credit: number | string }>,
) {
	const totalDebit = lines.reduce((s, l) => s + parseDecimal(l.debit), 0);
	const totalCredit = lines.reduce((s, l) => s + parseDecimal(l.credit), 0);
	return {
		totalDebit,
		totalCredit,
		isBalanced: Math.abs(totalDebit - totalCredit) < 0.01,
	};
}

export function isAssetAccount(code: string): boolean {
	return code.startsWith("1");
}

// ── Date input helpers ─────────────────────────────────────────────
// <input type="date"> works with local calendar dates ("yyyy-MM-dd"),
// while the API expects RFC3339 instants. These helpers bridge the gap
// without timezone off-by-one errors (GMT+7 safe).
//
// NOTE: never use `new Date().toISOString().slice(0, 10)` for the default
// value — toISOString() is UTC, so it shows yesterday after 00:00 local.
// And never use `new Date("yyyy-MM-dd")` for submit — a date-only string
// parses as UTC midnight per spec, shifting the instant by the local offset.

/** Today as "yyyy-MM-dd" in the *local* timezone (for <input type="date">). */
export function todayLocalDate(): string {
	return formatDate(new Date(), "yyyy-MM-dd");
}

/**
 * Convert a "yyyy-MM-dd" date-input value to an RFC3339 string for the API.
 * Uses local *noon*: the resulting instant stays on the same calendar date
 * in UTC and in every timezone from UTC-12 to UTC+12. (Local midnight would
 * shift the UTC date to the previous day, e.g. Sept 16 Jakarta becomes
 * Sept 15 17:00Z — breaking raw `slice(0, 10)` reads like the edit forms use.)
 */
export function dateInputToISO(date: string): string {
	return new Date(`${date}T12:00:00`).toISOString();
}

/**
 * Client's IANA timezone name (e.g. "Asia/Jakarta").
 * Centralized so every API call needing calendar rendering in the client's
 * zone reads it from one place. Falls back to "UTC" when unavailable (SSR).
 */
export function getTimeZone(): string {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
	} catch {
		return "UTC";
	}
}

/**
 * Full ISO-8601 instant like toISOString(), but keeping the *local* offset
 * (`+07:00`) instead of converting to UTC (`Z`).
 *
 * Use this when the server must interpret the date in the *client's* calendar
 * (e.g. "which month does the user mean?"). toISOString() loses the offset,
 * so a 1 Oct 00:30 +07:00 instant arrives as 30 Sep 17:30Z and a naive
 * UTC-based month split would land in the wrong month.
 */
export function localISO(date: Date): string {
	const pad = (n: number) => String(n).padStart(2, "0");
	const offsetMin = -date.getTimezoneOffset();
	const sign = offsetMin >= 0 ? "+" : "-";
	const abs = Math.abs(offsetMin);
	return (
		`${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
		`T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}` +
		`${sign}${pad(Math.floor(abs / 60))}:${pad(abs % 60)}`
	);
}
