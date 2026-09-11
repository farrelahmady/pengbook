import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import {
	ApiResponse,
	JournalEntryListResponse,
	JournalEntryListItem,
	CreateJournalDto,
} from "@/types";

const logger = createLogger("journal.service");
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

		logger.debug("Fetching journals (scroll view)", {
			limit: request.limit,
			hasCursor: !!request.cursor,
		});

		const res = await client.get<ApiResponse<JournalEntryListResponse>>(
			`${API_BASE}/journals?${params}`,
		);

		logger.debug("Journals fetched", {
			count: res.data.data.data.length,
			hasNextCursor: !!res.data.data.nextCursor,
		});

		return res.data.data;
	},

	getTotalSummary: async (): Promise<{
		totalDebit: number;
		totalCredit: number;
		transactionCount: number;
	}> => {
		const client = authHttpClient();

		logger.debug("Fetching journal summary");

		const res = await client.get<
			ApiResponse<{
				totalDebit: number;
				totalCredit: number;
				transactionCount: number;
			}>
		>(`${API_BASE}/journals/summary`);

		logger.info("Journal summary fetched", {
			totalDebit: res.data.data.totalDebit,
			totalCredit: res.data.data.totalCredit,
			transactionCount: res.data.data.transactionCount,
		});

		return res.data.data;
	},

	create: async (
		dto: CreateJournalDto,
	): Promise<JournalEntryListItem> => {
		const client = authHttpClient();

		logger.debug("Creating journal entry");

		const res = await client.post<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals`,
			dto,
		);

		logger.info("Journal entry created", {
			id: res.data.data.id,
		});

		return res.data.data;
	},

	update: async (
		id: string,
		dto: CreateJournalDto,
	): Promise<JournalEntryListItem> => {
		const client = authHttpClient();

		logger.debug("Updating journal entry", { id });

		const res = await client.put<ApiResponse<JournalEntryListItem>>(
			`${API_BASE}/journals/${id}`,
			dto,
		);

		logger.info("Journal entry updated", { id });

		return res.data.data;
	},
};
