"use client";

import { CoaTypeGroup, ChartOfAccountWithChildren } from "@/types";
import { cn } from "@/lib/utils";
import { useState } from "react";
import {
	Wallet,
	CreditCard,
	Landmark,
	TrendingUp,
	TrendingDown,
	ChevronUp,
} from "lucide-react";

const iconMap: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
	wallet: Wallet,
	"credit-card": CreditCard,
	landmark: Landmark,
	"trending-up": TrendingUp,
	"trending-down": TrendingDown,
};

const typeColorMap: Record<string, string> = {
	ASSET: "bg-primary-50 text-primary-600",
	LIABILITY: "bg-danger-50 text-danger-600",
	EQUITY: "bg-info-50 text-info-600",
	REVENUE: "bg-success-50 text-success-600",
	EXPENSE: "bg-warning-50 text-warning-600",
};

interface CoaTypeGroupCardProps {
	group: CoaTypeGroup;
}

function countPostingAccounts(accounts: ChartOfAccountWithChildren[]): number {
	let count = 0;
	for (const acc of accounts) {
		if (acc.isPosting) count++;
		count += countPostingAccounts(acc.children);
	}
	return count;
}

function CoaAccountTree({ accounts, depth = 0 }: { accounts: ChartOfAccountWithChildren[]; depth?: number }) {
	return (
		<>
			{accounts.map((acc) => (
				<div key={acc.id}>
					<div
						className={cn(
							"flex items-center justify-between px-4 py-2.5 border-b border-black/[0.04] last:border-b-0",
							depth > 0 && "bg-secondary-50/30",
						)}
						style={{ paddingLeft: `${16 + depth * 16}px` }}
					>
						<div className="flex items-center gap-2 min-w-0">
							{depth > 0 && (
								<div className="w-0.5 h-4 bg-secondary-200 rounded-full shrink-0" />
							)}
							<div>
								<p className="font-mono text-[10px] text-secondary-400">{acc.code}</p>
								<p
									className={cn(
										"text-[13px] leading-snug",
										acc.isPosting
											? "text-secondary-800 font-medium"
											: "text-secondary-600 font-semibold",
									)}
								>
									{acc.name}
								</p>
							</div>
						</div>
						<div className="shrink-0">
							{acc.isPosting ? (
								<span className="inline-block px-1.5 py-0.5 rounded text-[9px] font-semibold bg-success-50 text-success-700">
									Posting
								</span>
							) : (
								<span className="inline-block px-1.5 py-0.5 rounded text-[9px] font-semibold bg-secondary-100 text-secondary-500">
									Header
								</span>
							)}
						</div>
					</div>
					{acc.children.length > 0 && (
						<CoaAccountTree accounts={acc.children} depth={depth + 1} />
					)}
				</div>
			))}
		</>
	);
}

export function CoaTypeGroupCard({ group }: CoaTypeGroupCardProps) {
	const [expanded, setExpanded] = useState(false);
	const Icon = iconMap[group.icon] ?? Wallet;
	const postingCount = countPostingAccounts(group.accounts);

	return (
		<div className="card-default shadow-card overflow-hidden">
			{/* Header */}
			<button
				onClick={() => setExpanded(!expanded)}
				className="w-full flex items-center gap-3 px-4 py-3.5 text-left active:bg-secondary-50 transition-colors"
			>
				<div
					className={cn(
						"w-10 h-10 rounded-xl flex items-center justify-center shrink-0",
						typeColorMap[group.type] ?? "bg-secondary-100 text-secondary-500",
					)}
				>
					<Icon size={20} />
				</div>
				<div className="flex-1 min-w-0">
					<p className="text-[14px] font-semibold text-secondary-900 leading-snug">
						{group.label}
					</p>
					<p className="text-[11px] text-secondary-400">
						{group.count} akun · {postingCount} posting
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

			{/* Accounts tree */}
			{expanded && (
				<div className="border-t border-black/[0.06]">
					<CoaAccountTree accounts={group.accounts} />
				</div>
			)}
		</div>
	);
}
