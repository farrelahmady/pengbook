import {
	dummyTrialBalance,
	dummyIncomeStatement,
	dummyBalanceSheet,
	dummyCashFlow,
} from "@/lib/dummy-data";
import {
	TrialBalanceEntry,
	IncomeStatement,
	BalanceSheet,
	CashFlow,
	ReportSummary,
} from "@/types";

/**
 * Report Service
 *
 * @todo Replace with real API implementation when report feature is implemented
 * @todo Backend endpoints needed:
 *   - GET /api/v1/reports/summary
 *   - GET /api/v1/reports/trial-balance
 *   - GET /api/v1/reports/income-statement
 *   - GET /api/v1/reports/balance-sheet
 *   - GET /api/v1/reports/cash-flow
 *
 * Current implementation uses dummy data for development.
 * This should be replaced with actual API calls:
 *
 * ```typescript
 * import { authHttpClient } from "@/lib/http-client";
 * import { ApiResponse, ReportSummary, TrialBalanceEntry, IncomeStatement, BalanceSheet, CashFlow } from "@/types";
 *
 * export const reportService = {
 *   getSummary: async (): Promise<ReportSummary> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<ReportSummary>>("/api/v1/reports/summary");
 *     return res.data.data;
 *   },
 *
 *   getTrialBalance: async (): Promise<TrialBalanceEntry[]> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<TrialBalanceEntry[]>>("/api/v1/reports/trial-balance");
 *     return res.data.data;
 *   },
 *
 *   getIncomeStatement: async (): Promise<IncomeStatement> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<IncomeStatement>>("/api/v1/reports/income-statement");
 *     return res.data.data;
 *   },
 *
 *   getBalanceSheet: async (): Promise<BalanceSheet> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<BalanceSheet>>("/api/v1/reports/balance-sheet");
 *     return res.data.data;
 *   },
 *
 *   getCashFlow: async (): Promise<CashFlow> => {
 *     const client = authHttpClient();
 *     const res = await client.get<ApiResponse<CashFlow>>("/api/v1/reports/cash-flow");
 *     return res.data.data;
 *   },
 * };
 * ```
 */
export const reportService = {
	getSummary: async (): Promise<ReportSummary> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<ReportSummary>>("/api/v1/reports/summary");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 800));
		const totalDebit = dummyTrialBalance.reduce((s, e) => s + e.debit, 0);
		const totalCredit = dummyTrialBalance.reduce((s, e) => s + e.credit, 0);
		return {
			netIncome: dummyIncomeStatement.netIncome,
			totalAsset: dummyBalanceSheet.totalAssets,
			netCashFlow: dummyCashFlow.netCashFlow,
			isTrialBalanceBalanced: Math.abs(totalDebit - totalCredit) < 0.01,
		};
	},

	getTrialBalance: async (): Promise<TrialBalanceEntry[]> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<TrialBalanceEntry[]>>("/api/v1/reports/trial-balance");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyTrialBalance);
	},

	getIncomeStatement: async (): Promise<IncomeStatement> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<IncomeStatement>>("/api/v1/reports/income-statement");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyIncomeStatement);
	},

	getBalanceSheet: async (): Promise<BalanceSheet> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<BalanceSheet>>("/api/v1/reports/balance-sheet");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyBalanceSheet);
	},

	getCashFlow: async (): Promise<CashFlow> => {
		// TODO: Replace with actual API call
		// const client = authHttpClient();
		// const res = await client.get<ApiResponse<CashFlow>>("/api/v1/reports/cash-flow");
		// return res.data.data;

		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyCashFlow);
	},
};
