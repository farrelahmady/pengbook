import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import { ApiResponse, AccountSummary, AccountTree, PostingAccount } from "@/types";

const logger = createLogger("account.service");

const API_BASE = "/api/v1";

/**
 * Account Service
 *
 * Uses authHttpClient for authenticated endpoints.
 * Replaces the old dummy data implementation.
 */
export const accountService = {
	/**
	 * Get account summary counts (lightweight, no tree).
	 * Real API call to GET /api/v1/accounts/summary
	 */
	getSummary: async (): Promise<AccountSummary> => {
		logger.debug("Fetching account summary");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<AccountSummary>>(
			`${API_BASE}/accounts/summary`,
		);
		logger.info("Account summary fetched", {
			totalAccounts: res.data.data.totalAccounts,
			postingAccounts: res.data.data.postingAccounts,
		});
		return res.data.data;
	},

	/**
	 * Get account hierarchical tree grouped by type.
	 * Real API call to GET /api/v1/accounts/tree
	 */
	getTree: async (): Promise<AccountTree> => {
		logger.debug("Fetching account tree");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<AccountTree>>(
			`${API_BASE}/accounts/tree`,
		);
		logger.info("Account tree fetched", {
			groups: res.data.data.groups.length,
		});
		return res.data.data;
	},

	/**
	 * Get posting accounts (level=3) for journal forms.
	 * Real API call to GET /api/v1/accounts/posting
	 */
	getPostingAccounts: async (): Promise<PostingAccount[]> => {
		logger.debug("Fetching posting accounts");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<PostingAccount[]>>(
			`${API_BASE}/accounts/posting`,
		);
		logger.info("Posting accounts fetched", {
			count: res.data.data.length,
		});
		return res.data.data;
	},
};
