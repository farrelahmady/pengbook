"use client";

import { JournalEntry } from "@/types";
import { formatTime, parseDecimal } from "@/lib/utils";
import { DrBadge, CrBadge, DrCrDot } from "@/components/shared/dbcr-badge";
import { cn } from "@/lib/utils";
import { Skeleton } from "../ui/skeleton";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";

interface JournalCardProps {
	journal: JournalEntry;
	onClick?: () => void;
}

export function JournalCard({ journal, onClick }: JournalCardProps) {
	const currencyFormat = useCurrencyFormatter();
	// Hitung total debit utama untuk tampilan amount
	const totalDebit = journal.lines.reduce(
		(s, l) => s + parseDecimal(l.debit),
		0,
	);
	const totalCredit = journal.lines.reduce(
		(s, l) => s + parseDecimal(l.credit),
		0,
	);
	// Jika lebih banyak kredit, ini transaksi masuk
	const isInflow = totalCredit > totalDebit;
	const displayAmount = Math.max(totalDebit, totalCredit);

	return (
		<div
			onClick={onClick}
			className="card-default shadow-card active:scale-[0.99] transition-transform cursor-pointer"
		>
			{/* Header */}
			<div className="flex items-start justify-between px-4 pt-3.5 pb-3">
				<div className="flex-1 min-w-0 pr-3">
					<p className="text-[14px] font-semibold text-secondary-900 leading-snug truncate">
						{journal.description ?? "—"}
					</p>
					<p className="font-mono text-[10px] text-secondary-400 mt-0.5">
						{formatTime(journal.date)}
					</p>
				</div>
				<div className="text-right shrink-0">
					<p
						className={cn(
							"font-mono text-[15px] font-semibold",
							isInflow ? "text-success-600" : "text-danger-600",
						)}
					>
						{isInflow ? "+" : "-"}
						{currencyFormat(displayAmount, { compact: true })}
					</p>
				</div>
			</div>

			{/* Lines */}
			<div className="border-t border-black/[0.06] px-4 py-2.5 flex flex-col gap-2">
				{journal.lines.map((line) => {
					const isDr = parseDecimal(line.debit) > 0;
					const amount = isDr
						? parseDecimal(line.debit)
						: parseDecimal(line.credit);
					return (
						<div
							key={line.id}
							className="flex items-center justify-between gap-2"
						>
							<div className="flex items-center gap-2 min-w-0">
								<DrCrDot type={isDr ? "dr" : "cr"} />
								<span className="text-[12px] text-secondary-500 truncate">
									{line.account?.code} · {line.account?.name}
								</span>
							</div>
							<div className="flex items-center gap-1.5 shrink-0">
								{isDr ? <DrBadge /> : <CrBadge />}
								<span
									className={cn(
										"font-mono text-[11px] font-medium",
										isDr ? "text-danger-600" : "text-success-600",
									)}
								>
									{currencyFormat(amount, { compact: true })}
								</span>
							</div>
						</div>
					);
				})}
			</div>
		</div>
	);
}

interface JournalCardSkeletonProps {
	/** Jumlah baris jurnal (debit/kredit) yang ditampilkan, default 2 */
	lineCount?: number;
	className?: string;
}

export function JournalCardSkeleton({
	lineCount = 2,
	className,
}: JournalCardSkeletonProps) {
	return (
		<div className={cn("card-default shadow-card", className)}>
			{/* Header */}
			<div className="flex items-start justify-between px-4 pt-3.5 pb-3">
				<div className="flex-1 min-w-0 pr-3 space-y-1.5">
					{/* description */}
					<Skeleton className="h-[14px] w-[65%] rounded" />
					{/* time */}
					<Skeleton className="h-[10px] w-[40px] rounded" />
				</div>
				<div className="text-right shrink-0">
					{/* amount */}
					<Skeleton className="h-[15px] w-[90px] rounded" />
				</div>
			</div>

			{/* Lines */}
			<div className="border-t border-black/[0.06] px-4 py-2.5 flex flex-col gap-2">
				{Array.from({ length: lineCount }).map((_, i) => (
					<div key={i} className="flex items-center justify-between gap-2">
						<div className="flex items-center gap-2 min-w-0 flex-1">
							{/* DrCrDot */}
							<Skeleton className="h-2 w-2 rounded-full shrink-0" />
							{/* account code · name */}
							<Skeleton className="h-[12px] w-[55%] rounded" />
						</div>
						<div className="flex items-center gap-1.5 shrink-0">
							{/* DrBadge / CrBadge */}
							<Skeleton className="h-[16px] w-[28px] rounded-full" />
							{/* amount */}
							<Skeleton className="h-[11px] w-[60px] rounded" />
						</div>
					</div>
				))}
			</div>
		</div>
	);
}
