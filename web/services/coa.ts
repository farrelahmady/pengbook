import { dummyFullCoa } from "@/lib/dummy-data";
import { CoaSummary } from "@/types";

export const coaService = {
	getSummary: async (): Promise<CoaSummary> => {
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyFullCoa);
	},
};
