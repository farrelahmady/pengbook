import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import { ApiResponse, AccountSummary } from "@/types";

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
	 * Get account summary with hierarchical tree.
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
};
