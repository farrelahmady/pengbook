import { authHttpClient } from "@/lib/http-client";
import { createLogger } from "@/lib/logger";
import {
	ApiResponse,
	Account,
	AccountSummary,
	AccountTree,
	CreateAccountDto,
	UpdateAccountDto,
	ParentListItem,
	PostingAccount,
} from "@/types";

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
	 * Get candidate parents at the given parent level (0-2) for account creation.
	 * Real API call to GET /api/v1/accounts/parents?level=N
	 */
	getParents: async (
		level: number,
		type?: string,
	): Promise<ParentListItem[]> => {
		logger.debug("Fetching parent accounts", { level });

		const params = new URLSearchParams();
		params.set("level", String(level));
		if (type) params.set(type, type);

		const client = authHttpClient();
		const res = await client.get<ApiResponse<ParentListItem[]>>(
			`${API_BASE}/accounts/parents?${params}`,
		);
		logger.info("Parent accounts fetched", {
			level,
			type,
			count: res.data.data.length,
		});
		return res.data.data;
	},

	/**
	 * Create a new account. Code is generated server-side from the parent.
	 * Real API call to POST /api/v1/accounts
	 */
	create: async (dto: CreateAccountDto): Promise<Account> => {
		logger.debug("Creating account", {
			name: dto.name,
			parentId: dto.parentId,
		});
		const client = authHttpClient();
		const res = await client.post<ApiResponse<Account>>(
			`${API_BASE}/accounts`,
			dto,
		);
		logger.info("Account created", {
			id: res.data.data.id,
			code: res.data.data.code,
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

	/**
	 * Rename an account. Real API call to PUT /api/v1/accounts/{id}
	 */
	update: async (id: number, dto: UpdateAccountDto): Promise<Account> => {
		logger.debug("Updating account", { id, name: dto.name });
		const client = authHttpClient();
		const res = await client.put<ApiResponse<Account>>(
			`${API_BASE}/accounts/${id}`,
			dto,
		);
		logger.info("Account updated", { id, code: res.data.data.code });
		return res.data.data;
	},

	/**
	 * Download the chart of accounts as an Excel file.
	 * Real API call to GET /api/v1/accounts/export
	 */
	exportExcel: async (): Promise<Blob> => {
		logger.debug("Downloading account export");
		const client = authHttpClient();
		const res = await client.get(`${API_BASE}/accounts/export`, {
			responseType: "blob",
		});
		logger.info("Account export downloaded");
		return res.data as Blob;
	},
};
