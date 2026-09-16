import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import { ApiResponse, AssetGroups, AssetSummary } from "@/types";

const logger = createLogger("asset.service");

const API_BASE = "/api/v1/accounts/assets";

/**
 * Asset Service
 *
 * Read endpoints nested di modul account (keputusan desain 16 Sep 2026):
 * - GET /api/v1/accounts/assets/summary (ringan, untuk Topbar)
 * - GET /api/v1/accounts/assets/groups (berat, untuk List)
 *
 * Uses authHttpClient for authenticated endpoints.
 */
export const assetService = {
	/**
	 * Get asset summary aggregates (lightweight, no groups).
	 * Real API call to GET /api/v1/accounts/assets/summary
	 */
	getSummary: async (): Promise<AssetSummary> => {
		logger.debug("Fetching asset summary");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<AssetSummary>>(
			`${API_BASE}/summary`,
		);
		logger.info("Asset summary fetched", {
			totalAsset: res.data.data.totalAsset,
			accountCount: res.data.data.accountCount,
		});
		return res.data.data;
	},

	/**
	 * Get asset groups with posting accounts and balances.
	 * Real API call to GET /api/v1/accounts/assets/groups
	 */
	getGroups: async (): Promise<AssetGroups> => {
		logger.debug("Fetching asset groups");
		const client = authHttpClient();
		const res = await client.get<ApiResponse<AssetGroups>>(
			`${API_BASE}/groups`,
		);
		logger.info("Asset groups fetched", {
			groups: res.data.data.groups.length,
		});
		return res.data.data;
	},
};
