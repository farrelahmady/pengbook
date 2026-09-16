import {
	AssetSummary,
	TrialBalanceEntry,
	IncomeStatement,
	BalanceSheet,
	CashFlow,
} from "@/types";

// ── Dummy balance per account (DR - CR) ───────────────────────────────────
export const dummyBalances: Record<string, number> = {
	a1: 49100000, // Mandiri
	a2: 50000000, // BCA
	a3: 3000000, // Jago
	a4: 2500000, // Wallet Cash
	a5: 800000, // Gopay
	a6: 350000, // Dana
	a7: 12000000, // Stockbit
	a8: 5000000, // Pluang
	a9: 8500000, // Gold
	a10: 15000000, // Emergency Fund
};

// ── Dummy Laporan data ────────────────────────────────────────────────────
export const dummyTrialBalance: TrialBalanceEntry[] = [
	{ code: "1101", name: "Kas", type: "ASSET", debit: 5500000, credit: 0 },
	{ code: "1102", name: "Bank BCA", type: "ASSET", debit: 8700000, credit: 0 },
	{
		code: "1110",
		name: "Piutang Usaha",
		type: "ASSET",
		debit: 1000000,
		credit: 0,
	},
	{ code: "1201", name: "Peralatan", type: "ASSET", debit: 6000000, credit: 0 },
	{
		code: "1209",
		name: "Akum. Penyusutan",
		type: "ASSET",
		debit: 0,
		credit: 1500000,
	},
	{
		code: "2001",
		name: "Hutang Dagang",
		type: "LIABILITY",
		debit: 0,
		credit: 3000000,
	},
	{ code: "3001", name: "Modal", type: "EQUITY", debit: 0, credit: 13000000 },
	{
		code: "4001",
		name: "Pendapatan Jasa",
		type: "REVENUE",
		debit: 0,
		credit: 6000000,
	},
	{
		code: "6001",
		name: "Biaya Perlengkapan",
		type: "EXPENSE",
		debit: 500000,
		credit: 0,
	},
	{
		code: "6002",
		name: "Beban Gaji",
		type: "EXPENSE",
		debit: 1500000,
		credit: 0,
	},
	{
		code: "6003",
		name: "Beban Lain-lain",
		type: "EXPENSE",
		debit: 300000,
		credit: 0,
	},
];

export const dummyIncomeStatement: IncomeStatement = {
	revenues: [{ name: "Pendapatan Jasa", amount: 6000000 }],
	expenses: [
		{ name: "Biaya Perlengkapan", amount: 500000 },
		{ name: "Beban Gaji", amount: 1500000 },
		{ name: "Beban Lain-lain", amount: 300000 },
	],
	totalRevenue: 6000000,
	totalExpense: 2300000,
	netIncome: 3700000,
};

export const dummyBalanceSheet: BalanceSheet = {
	currentAssets: [
		{ name: "Kas", balance: 5500000 },
		{ name: "Bank BCA", balance: 8700000 },
		{ name: "Piutang Usaha", balance: 1000000 },
	],
	nonCurrentAssets: [
		{ name: "Peralatan", balance: 6000000 },
		{ name: "Akum. Penyusutan", balance: -1500000 },
	],
	liabilities: [{ name: "Hutang Dagang", balance: 3000000 }],
	equity: [
		{ name: "Modal", balance: 13000000 },
		{ name: "Laba Periode Ini", balance: 3700000 },
	],
	totalCurrentAssets: 15200000,
	totalNonCurrentAssets: 4500000,
	totalAssets: 19700000,
	totalLiabilities: 3000000,
	totalEquity: 16700000,
};

export const dummyCashFlow: CashFlow = {
	operating: {
		lines: [
			{ label: "Penerimaan dari pelanggan", amount: 6000000 },
			{ label: "Pembayaran beban operasi", amount: -2300000 },
			{ label: "Pembayaran hutang dagang", amount: -2500000 },
		],
		total: 1200000,
	},
	investing: {
		lines: [{ label: "Pembelian instrumen investasi", amount: -5000000 }],
		total: -5000000,
	},
	financing: {
		lines: [{ label: "Setoran modal", amount: 50000000 }],
		total: 50000000,
	},
	netCashFlow: 46200000,
};

// ── Dummy Aset (preview halaman Aset, sementara sebelum API asli) ──────────
// Kontrak mengikuti `AssetSummary` di `@/types` agar Topbar/List/Card bisa
// dirender tanpa backend. Akan dibuang saat GET /api/v1/assets/* jadi.
export const dummyAssetSummary: AssetSummary = {
	totalAsset: 132250000,
	currentAsset: 105750000,
	groups: [
		{
			id: 1,
			code: "1.01",
			name: "Kas & Tunai",
			icon: "wallet",
			totalBalance: 101600000,
			accounts: [
				{
					id: 101,
					code: "1.01.01.01",
					name: "Mandiri - Main",
					balance: 49100000,
					isPosting: true,
				},
				{
					id: 102,
					code: "1.01.01.02",
					name: "BCA - Operasional",
					balance: 50000000,
					isPosting: true,
				},
				{
					id: 103,
					code: "1.01.01.03",
					name: "Wallet Cash",
					balance: 2500000,
					isPosting: true,
				},
			],
		},
		{
			id: 2,
			code: "1.02",
			name: "Bank & E-Wallet",
			icon: "building",
			totalBalance: 4150000,
			accounts: [
				{
					id: 104,
					code: "1.02.01.01",
					name: "Jago",
					balance: 3000000,
					isPosting: true,
				},
				{
					id: 105,
					code: "1.02.01.02",
					name: "GoPay",
					balance: 800000,
					isPosting: true,
				},
				{
					id: 106,
					code: "1.02.01.03",
					name: "DANA",
					balance: 350000,
					isPosting: true,
				},
			],
		},
		{
			id: 3,
			code: "1.03",
			name: "Piutang",
			icon: "clock",
			totalBalance: 1000000,
			accounts: [
				{
					id: 107,
					code: "1.03.01.01",
					name: "Tagihan Klien A",
					balance: 1000000,
					isPosting: true,
				},
			],
		},
		{
			id: 4,
			code: "1.04",
			name: "Investasi",
			icon: "package",
			totalBalance: 25500000,
			accounts: [
				{
					id: 108,
					code: "1.04.01.01",
					name: "Stockbit",
					balance: 12000000,
					isPosting: true,
				},
				{
					id: 109,
					code: "1.04.01.02",
					name: "Pluang",
					balance: 5000000,
					isPosting: true,
				},
				{
					id: 110,
					code: "1.04.01.03",
					name: "Emas",
					balance: 8500000,
					isPosting: true,
				},
			],
		},
	],
};
