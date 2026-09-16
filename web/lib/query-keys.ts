/**
 * Query Keys
 *
 * Centralized query key definitions for React Query.
 * Using namespace pattern for consistent invalidation.
 *
 * Pattern: ["namespace", "resource", { filters? }]
 *
 * @example
 * // Invalidate all journal queries
 * queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
 *
 * // Invalidate specific query
 * queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
 */

// ── Journals ─────────────────────────────────────────────

export const queryKeys = {
	journals: {
		all: ["journals"] as const,
		scrollView: (filters?: {
			startDate?: Date;
			endDate?: Date;
			accountIds?: string[];
		}) => ["journals", "scroll-view", filters] as const,
		summary: ["journals", "summary"] as const,
		summaryForMonth: (month: string) =>
			["journals", "summary", month] as const,
	},

	// ── Accounts (Chart of Accounts) ───────────────────────

	accounts: {
		all: ["accounts"] as const,
		summary: ["accounts", "summary"] as const,
		tree: ["accounts", "tree"] as const,
		posting: ["accounts", "posting"] as const,
	},

	// ── Assets ──────────────────────────────────────────────

	assets: {
		all: ["assets"] as const,
		summary: ["assets", "summary"] as const,
	},

	// ── Reports ─────────────────────────────────────────────

	reports: {
		all: ["reports"] as const,
		summary: ["reports", "summary"] as const,
		trialBalance: ["reports", "trial-balance"] as const,
		incomeStatement: ["reports", "income-statement"] as const,
		balanceSheet: ["reports", "balance-sheet"] as const,
		cashFlow: ["reports", "cash-flow"] as const,
	},
} as const;
