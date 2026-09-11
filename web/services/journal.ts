import { authHttpClient } from "@/lib/http-client";
import {
	ApiResponse,
	JournalEntryListResponse,
	JournalEntryListItem,
	CreateJournalDto,
} from "@/types";

const API_BASE = "/api/v1";

export const journalService = {
	getAllScrollView: async (
		request: {
			limit: number;
			cursor?: string;
			startDate?: Date;
			endDate?: Date;
			accountIds?: string[];
		},
	): Promise<JournalEntryListResponse> => {
		const client = authHttpClient();

		const params = new URLSearchParams();
		params.set("limit", String(request.limit));
		if (request.cursor) params.set("cursor", request.cursor);
		if (request.startDate)
			params.set("startDate", request.startDate.toISOString());
		if (request.endDate) params.set("endDate", request.endDate.toISOString());
		if (request.accountIds && request.accountIds.length > 0) {
			params.set("accountIds", request.accountIds.join(","));
		}

		const res = await client.get<ApiResponse<JournalEntryListResponse>>(
			`${API_BASE}/journals?${params}`,
		);

		return res.data.data;
	},

	getTotalSummary: async (): Promise<{
		totalDebit: number;
		totalCredit: number;
		transactionCount: number;
	}> => {
		const client = authHttpClient();
		const res = await client.get<
			ApiResponse<{
				totalDebit: number;
				totalCredit: number;
				transactionCount: number;
			}>
		>(`${API_BASE}/journals/summary`);
		return res.data.data;
	},

	create: async (
		dto: CreateJournalDto,
	): Promise<JournalEntryListItem> => {
		const client = authHttpClient();
		const res = await client.post<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals`,
			dto,
		);
		return res.data.data;
	},

	update: async (
		id: string,
		dto: CreateJournalDto,
	): Promise<JournalEntryListItem> => {
		const client = authHttpClient();
		const res = await client.put<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals/${id}`,
			dto,
		);
		return res.data.data;
	},
};
