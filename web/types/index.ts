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

// Account (Chart of Accounts)
export type AccountType =
	| "ASSET"
	| "LIABILITY"
	| "EQUITY"
	| "REVENUE"
	| "EXPENSE"
	| "OTHER";

export interface Account {
	id: number;
	code: string;
	name: string;
	type: AccountType;
	level: number;
	isPosting: boolean;
	parentId: number | null;
	createdAt: string;
	updatedAt: string;
}

export interface AccountWithChildren extends Account {
	children: AccountWithChildren[];
}

export interface CreateAccountDto {
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
	account?: Account;
}

export interface JournalEntry {
	id: number;
	date: string;
	description: string | null;
	lines: JournalEntryLine[];
	createdAt: string;
	updatedAt: string;
}

// Journal List View (optimized for UI display)
export interface JournalEntryListItem {
	id: number;
	datetime: string;
	description: string | null;
	netEffect: number;
	lines: JournalLineItem[];
}

export interface JournalLineItem {
	id: number;
	debit: number;
	credit: number;
	accountDisplay: string;
}

export interface JournalEntryListResponse {
	data: JournalEntryListItem[];
	nextCursor: string | null;
}

export interface JournalLineDto {
	accountId?: number;
	accountCode?: string;
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

// Posting Account (simplified for journal forms)
export interface PostingAccount {
	id: number;
	code: string;
	name: string;
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

// Account Page
export interface AccountTypeGroup {
	type: AccountType;
	label: string;
	icon: string;
	accounts: AccountWithChildren[];
	count: number;
}

export interface AccountSummary {
	totalAccounts: number;
	postingAccounts: number;
	headerAccounts: number;
	groups: AccountTypeGroup[];
}
