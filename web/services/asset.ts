import { dummyAssetSummary } from "@/lib/dummy-data";
import { AssetSummary } from "@/types";

export const assetService = {
	getSummary: async (): Promise<AssetSummary> => {
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyAssetSummary);
	},
};
