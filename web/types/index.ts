// API wrapper
export interface ApiResponse<T> {
	success: boolean;
	message: string;
	data: T;
	meta?: { timestamp: string; path?: string };
}

// COA
export type AccountType =
	| "ASSET"
	| "LIABILITY"
	| "EQUITY"
	| "REVENUE"
	| "EXPENSE"
	| "OTHER";

export interface ChartOfAccount {
	id: string;
	code: string;
	name: string;
	type: AccountType;
	isPosting: boolean;
	parentId: string | null;
	createdAt: string;
	updatedAt: string;
}

export interface ChartOfAccountWithChildren extends ChartOfAccount {
	children: ChartOfAccountWithChildren[];
}

export interface CreateCoaDto {
	code: string;
	name: string;
	type: AccountType;
	parentId?: string;
}

// Journal
export interface JournalEntryLine {
	id: string;
	journalEntryId: string;
	accountId: string;
	debit: string;
	credit: string;
	account?: ChartOfAccount;
}

export interface JournalEntry {
	id: string;
	date: string;
	description: string | null;
	lines: JournalEntryLine[];
	createdAt: string;
	updatedAt: string;
}

export interface JournalLineDto {
	accountId: string;
	debit: number;
	credit: number;
}

export interface CreateJournalDto {
	date: string;
	description?: string;
	lines: JournalLineDto[];
}

export interface JournalSummary {
	totalDebit: number;
	totalCredit: number;
	transactionCount: number;
}

export type JournalMode = "basic" | "advanced";
export type NavTab = "jurnal" | "aset" | "laporan" | "akun";

// Asset
export interface AssetAccount {
	id: string;
	code: string;
	name: string;
	balance: number;
	isPosting: boolean;
}

export interface AssetGroup {
	id: string;
	code: string;
	name: string;
	icon: string;
	accounts: AssetAccount[];
	totalBalance: number;
}

export interface AssetSummary {
	totalAsset: number;
	currentAsset: number;
	groups: AssetGroup[];
}

// Report
export interface TrialBalanceEntry {
	code: string;
	name: string;
	type: AccountType;
	debit: number;
	credit: number;
}

export interface TrialBalanceGroup {
	label: string;
	entries: TrialBalanceEntry[];
}

export interface IncomeStatementItem {
	name: string;
	amount: number;
}

export interface IncomeStatement {
	revenues: IncomeStatementItem[];
	expenses: IncomeStatementItem[];
	totalRevenue: number;
	totalExpense: number;
	netIncome: number;
}

export interface BalanceSheetItem {
	name: string;
	balance: number;
}

export interface BalanceSheet {
	currentAssets: BalanceSheetItem[];
	nonCurrentAssets: BalanceSheetItem[];
	liabilities: BalanceSheetItem[];
	equity: BalanceSheetItem[];
	totalCurrentAssets: number;
	totalNonCurrentAssets: number;
	totalAssets: number;
	totalLiabilities: number;
	totalEquity: number;
}

export interface CashFlowLine {
	label: string;
	amount: number;
}

export interface CashFlowSection {
	lines: CashFlowLine[];
	total: number;
}

export interface CashFlow {
	operating: CashFlowSection;
	investing: CashFlowSection;
	financing: CashFlowSection;
	netCashFlow: number;
}

export interface ReportSummary {
	netIncome: number;
	totalAsset: number;
	netCashFlow: number;
	isTrialBalanceBalanced: boolean;
}

// COA Page
export interface CoaTypeGroup {
	type: AccountType;
	label: string;
	icon: string;
	accounts: ChartOfAccountWithChildren[];
	count: number;
}

export interface CoaSummary {
	totalAccounts: number;
	postingAccounts: number;
	headerAccounts: number;
	groups: CoaTypeGroup[];
}
