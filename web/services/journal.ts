import { createHttpClient } from "@/lib/http-client";
import {
	ApiResponse,
	JournalEntryListResponse,
	JournalEntryListItem,
	CreateJournalDto,
} from "@/types";

const API_BASE = "/api/v1";

function journalClient(getToken?: () => Promise<string | null>) {
	return createHttpClient(getToken);
}

export const journalService = {
	getAllScrollView: async (
		request: {
			limit: number;
			cursor?: string;
			startDate?: Date;
			endDate?: Date;
			accountIds?: string[];
		},
		getToken?: () => Promise<string | null>,
	): Promise<JournalEntryListResponse> => {
		const client = journalClient(getToken);

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

	getTotalSummary: async (
		getToken?: () => Promise<string | null>,
	): Promise<{
		totalDebit: number;
		totalCredit: number;
		transactionCount: number;
	}> => {
		const client = journalClient(getToken);
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
		getToken?: () => Promise<string | null>,
	): Promise<JournalEntryListItem> => {
		const client = journalClient(getToken);
		const res = await client.post<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals`,
			dto,
		);
		return res.data.data;
	},

	update: async (
		id: string,
		dto: CreateJournalDto,
		getToken?: () => Promise<string | null>,
	): Promise<JournalEntryListItem> => {
		const client = journalClient(getToken);
		const res = await client.put<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals/${id}`,
			dto,
		);
		return res.data.data;
	},
};
