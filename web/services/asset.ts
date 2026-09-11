import { dummyAssetSummary } from "@/lib/dummy-data";
import { AssetSummary } from "@/types";

/**
 * Asset Service
 *
 * @todo Replace with real API implementation when asset feature is implemented
 * @todo Backend endpoint: GET /api/v1/assets/summary
 * @todo Expected response: AssetSummary
 *
 * Current implementation uses dummy data for development.
 * This should be replaced with actual API call:
 *
 * ```typescript
 * import { authHttpClient } from "@/lib/http-client";
 * import { ApiResponse, AssetSummary } from "@/types";
 *
 * export const assetService = {
 *   getSummary: async (): Promise<AssetSummary> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<AssetSummary>>("/api/v1/assets/summary");
 *     return res.data.data;
 *   },
 * };
 * ```
 */
export const assetService = {
	getSummary: async (): Promise<AssetSummary> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<AssetSummary>>("/api/v1/assets/summary");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyAssetSummary);
	},
};
