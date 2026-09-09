// API wrapper
export interface ApiResponse<T> {
	success: boolean;
	message: string;
	data: T;
	meta?: { timestamp: string; path?: string };
}

// Auth
export interface User {
	id: number;
	name: string;
	username: string;
	email: string;
	createdAt: string;
}

export interface TokenResponse {
	accessToken: string;
	refreshToken: string;
	expiresIn: number;
	tokenType: string;
}

export interface LoginRequest {
	identifier: string;
	password: string;
}

export interface RegisterRequest {
	name: string;
	username: string;
	email: string;
	password: string;
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
	id: number;
	code: string;
	name: string;
	type: AccountType;
	isPosting: boolean;
	parentId: number | null;
	createdAt: string;
	updatedAt: string;
}

export interface ChartOfAccountWithChildren extends ChartOfAccount {
	children: ChartOfAccountWithChildren[];
}

export interface CreateCoaDto {
	code: string;
	name: string;
	parentId?: number;
}

// Journal
export interface JournalEntryLine {
	id: number;
	journalEntryId: number;
	accountId: number;
	debit: number;
	credit: number;
	account?: ChartOfAccount;
}

export interface JournalEntry {
	id: number;
	date: string;
	description: string | null;
	lines: JournalEntryLine[];
	createdAt: string;
	updatedAt: string;
}

export interface JournalLineDto {
	accountId: number;
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
	id: number;
	code: string;
	name: string;
	balance: number;
	isPosting: boolean;
}

export interface AssetGroup {
	id: number;
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
