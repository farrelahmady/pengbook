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

export const reportService = {
	getSummary: async (): Promise<ReportSummary> => {
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
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyTrialBalance);
	},

	getIncomeStatement: async (): Promise<IncomeStatement> => {
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyIncomeStatement);
	},

	getBalanceSheet: async (): Promise<BalanceSheet> => {
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyBalanceSheet);
	},

	getCashFlow: async (): Promise<CashFlow> => {
		await new Promise((resolve) => setTimeout(resolve, 1000));
		return Promise.resolve(dummyCashFlow);
	},
};
