"use client";

import { TrialBalanceEntry, TrialBalanceGroup } from "@/types";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { cn } from "@/lib/utils";
import { useState, useMemo } from "react";
import { SquareCheck, ChevronUp, Download } from "lucide-react";
import { useTranslations } from "next-intl";

interface TrialBalanceCardProps {
	entries: TrialBalanceEntry[];
}

const GROUP_ORDER = ["ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE"] as const;

const GROUP_LABELS: Record<string, string> = {
	ASSET: "Aset",
	LIABILITY: "Liabilitas & Ekuitas",
	EQUITY: "Liabilitas & Ekuitas",
	REVENUE: "Pendapatan & Biaya",
	EXPENSE: "Pendapatan & Biaya",
};

export function TrialBalanceCard({ entries }: TrialBalanceCardProps) {
	const t = useTranslations("reportPage");
	const [expanded, setExpanded] = useState(true);
	const currencyFormat = useCurrencyFormatter();

	const groups = useMemo(() => {
		const map = new Map<string, TrialBalanceEntry[]>();
		for (const entry of entries) {
			const key = GROUP_LABELS[entry.type] ?? entry.type;
			if (!map.has(key)) map.set(key, []);
			map.get(key)!.push(entry);
		}
		const result: TrialBalanceGroup[] = [];
		const seen = new Set<string>();
		for (const type of GROUP_ORDER) {
			const label = GROUP_LABELS[type];
			if (label && !seen.has(label)) {
				seen.add(label);
				const groupEntries = map.get(label) ?? [];
				if (groupEntries.length > 0) {
					result.push({ label, entries: groupEntries });
				}
			}
		}
		return result;
	}, [entries]);

	const totalDebit = entries.reduce((s, e) => s + e.debit, 0);
	const totalCredit = entries.reduce((s, e) => s + e.credit, 0);
	const isBalanced = Math.abs(totalDebit - totalCredit) < 0.01;

	return (
		<div className="card-default shadow-card overflow-hidden">
			{/* Header */}
			<button
				onClick={() => setExpanded(!expanded)}
				className="w-full flex items-center gap-3 px-4 py-3.5 text-left active:bg-secondary-50 transition-colors"
			>
				<div className="w-10 h-10 rounded-xl bg-secondary-100 flex items-center justify-center shrink-0">
					<SquareCheck size={20} className="text-secondary-500" />
				</div>
				<div className="flex-1 min-w-0">
					<p className="text-[14px] font-semibold text-secondary-900 leading-snug">
						{t("trialBalance.name")}
					</p>
					<p className="text-[11px] text-secondary-400">{t("trialBalance.subtitle")}</p>
				</div>
				<div className="text-right shrink-0">
					<p
						className={cn(
							"font-mono text-[13px] font-semibold",
							isBalanced ? "text-success-600" : "text-danger-600",
						)}
					>
						{isBalanced ? "Balance" : "Unbalanced"}
					</p>
				</div>
				<ChevronUp
					size={16}
					className={cn(
						"text-secondary-400 transition-transform duration-200 shrink-0",
						expanded ? "" : "rotate-180",
					)}
				/>
			</button>

			{/* Table */}
			{expanded && (
				<div className="border-t border-black/[0.06]">
					{/* Column headers */}
					<div className="grid grid-cols-[1fr_100px_100px] px-4 py-2 bg-secondary-50/50">
						<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400">
							{t("trialBalance.headers.account")}
						</p>
						<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400 text-right">
							{t("trialBalance.headers.debit")}
						</p>
						<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400 text-right">
							{t("trialBalance.headers.credit")}
						</p>
					</div>

					{/* Groups */}
					{groups.map((group) => (
						<div key={group.label}>
							<p className="text-[10px] font-semibold uppercase tracking-wide text-primary-600 px-4 py-2 bg-primary-50/30">
								{group.label}
							</p>
							{group.entries.map((entry) => (
								<div
									key={entry.code}
									className="grid grid-cols-[1fr_100px_100px] px-4 py-2.5 border-b border-black/[0.04] items-center"
								>
									<p className="text-[12px] text-secondary-700">
										<span className="font-mono text-secondary-400">{entry.code}</span>
										{" · "}
										{entry.name}
									</p>
									<p
										className={cn(
											"font-mono text-[12px] font-medium text-right",
											entry.debit > 0 ? "text-success-600" : "text-secondary-300",
										)}
									>
										{entry.debit > 0 ? currencyFormat(entry.debit) : "—"}
									</p>
									<p
										className={cn(
											"font-mono text-[12px] font-medium text-right",
											entry.credit > 0 ? "text-danger-600" : "text-secondary-300",
										)}
									>
										{entry.credit > 0 ? currencyFormat(entry.credit) : "—"}
									</p>
								</div>
							))}
						</div>
					))}

					{/* Total */}
					<div className="grid grid-cols-[1fr_100px_100px] px-4 py-3 bg-secondary-900 text-white items-center">
						<p className="text-[12px] font-semibold">Total</p>
						<p className="font-mono text-[12px] font-semibold text-right">
							{currencyFormat(totalDebit)}
						</p>
						<p className="font-mono text-[12px] font-semibold text-right">
							{currencyFormat(totalCredit)}
						</p>
					</div>

					{/* Export button */}
					<div className="flex justify-center px-4 py-3">
						<button className="flex items-center gap-2 px-4 py-2 rounded-xl border border-secondary-200 text-[12px] font-medium text-secondary-600 hover:bg-secondary-50 transition-colors">
							<Download size={14} />
							{t("trialBalance.export")}
						</button>
					</div>
				</div>
			)}
		</div>
	);
}
