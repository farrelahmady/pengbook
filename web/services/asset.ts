import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import {
	AccountResponse,
	ApiResponse,
	AssetGroups,
	AssetSummary,
} from "@/types";

const logger = createLogger("asset.service");

const API_BASE = "/api/v1/accounts/assets";

/**
 * Asset Service
 *
 * Read endpoints nested di modul account (keputusan desain 16 Sep 2026):
 * - GET /api/v1/accounts/assets/summary (ringan, untuk Topbar)
 * - GET /api/v1/accounts/assets/groups (berat, untuk List)
 *
 * Mutation endpoints:
 * - POST /api/v1/accounts/assets (create asset)
 * - POST /api/v1/accounts/assets/{id}/adjust-balance (adjust balance)
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

	/**
	 * Create a new asset posting account.
	 * POST /api/v1/accounts/assets
	 */
	create: async (parentId: number, name: string): Promise<AccountResponse> => {
		logger.debug("Creating asset", { parentId, name });
		const client = authHttpClient();
		const res = await client.post<ApiResponse<AccountResponse>>(
			`${API_BASE}`,
			{ parentId, name },
		);
		logger.info("Asset created", {
			id: res.data.data.id,
			code: res.data.data.code,
		});
		return res.data.data;
	},

	/**
	 * Adjust the balance of an asset posting account to the desired balance.
	 * Server calculates delta and creates balanced journal entry.
	 * POST /api/v1/accounts/assets/{id}/adjust-balance
	 */
	adjustBalance: async (
		assetId: number,
		balance: number,
		note?: string,
		date?: string,
	): Promise<{ journalEntryId: number; balance: number }> => {
		logger.debug("Adjusting asset balance", {
			assetId,
			desiredBalance: balance,
			note,
		});
		const client = authHttpClient();
		const res = await client.post<
			ApiResponse<{ journalEntryId: number; balance: number }>
		>(`${API_BASE}/${assetId}/adjust-balance`, {
			balance,
			...(note && { note }),
			...(date && { date }),
		});
		logger.info("Asset balance adjusted", {
			assetId,
			journalEntryId: res.data.data.journalEntryId,
			newBalance: res.data.data.balance,
		});
		return res.data.data;
	},
};
