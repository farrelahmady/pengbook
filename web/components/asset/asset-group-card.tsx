"use client";

import { AssetGroup } from "@/types";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { Wallet, Building2, Clock, Package, ChevronUp } from "lucide-react";

const iconMap: Record<string, React.ComponentType<{ size?: number; className?: string }>> = {
	wallet: Wallet,
	building: Building2,
	clock: Clock,
	package: Package,
};

interface AssetGroupCardProps {
	group: AssetGroup;
}

export function AssetGroupCard({ group }: AssetGroupCardProps) {
	const [expanded, setExpanded] = useState(true);
	const currencyFormat = useCurrencyFormatter();
	const Icon = iconMap[group.icon] ?? Wallet;

	return (
		<div className="card-default shadow-card overflow-hidden">
			{/* Header */}
			<button
				onClick={() => setExpanded(!expanded)}
				className="w-full flex items-center gap-3 px-4 py-3.5 text-left active:bg-secondary-50 transition-colors"
			>
				<div className="w-10 h-10 rounded-xl bg-secondary-100 flex items-center justify-center shrink-0">
					<Icon size={20} className="text-secondary-500" />
				</div>
				<div className="flex-1 min-w-0">
					<p className="font-mono text-[10px] text-secondary-400">{group.code}</p>
					<p className="text-[14px] font-semibold text-secondary-900 leading-snug">
						{group.name}
					</p>
				</div>
				<div className="text-right shrink-0">
					<p
						className={cn(
							"font-mono text-[15px] font-semibold",
							group.totalBalance >= 0 ? "text-success-600" : "text-danger-600",
						)}
					>
						{currencyFormat(group.totalBalance, { compact: true, currency: "Rp" })}
					</p>
					<p className="font-mono text-[10px] text-secondary-400">
						{group.accounts.length} sub-akun
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

			{/* Sub-accounts */}
			{expanded && group.accounts.length > 0 && (
				<div className="border-t border-black/[0.06]">
					<div className="flex flex-col">
						{group.accounts.map((account) => (
							<div
								key={account.id}
								className="flex items-center justify-between px-4 py-3 border-b border-black/[0.04] last:border-b-0"
							>
								<div className="flex items-center gap-2 min-w-0">
									<div className="w-0.5 h-4 bg-secondary-200 rounded-full shrink-0" />
									<div>
										<p className="font-mono text-[10px] text-secondary-400">
											{account.code}
										</p>
										<p className="text-[13px] text-secondary-700">{account.name}</p>
									</div>
								</div>
								<div className="text-right shrink-0">
									<p
										className={cn(
											"font-mono text-[13px] font-medium",
											account.balance >= 0
												? "text-success-600"
												: "text-danger-600",
										)}
									>
										{currencyFormat(account.balance, { compact: true, currency: "Rp" })}
									</p>
									{account.isPosting && (
										<span className="inline-block mt-0.5 px-1.5 py-0.5 rounded text-[9px] font-semibold bg-success-50 text-success-700">
											Posting
										</span>
									)}
								</div>
							</div>
						))}
					</div>

					{/* Subtotal */}
					<div className="flex items-center justify-between px-4 py-2.5 bg-secondary-50/50 border-t border-black/[0.06]">
						<p className="text-[11px] text-secondary-400">Subtotal debit bersih</p>
						<p className="font-mono text-[12px] font-medium text-secondary-600">
							DR {currencyFormat(group.totalBalance, { compact: true, currency: "Rp" })}
						</p>
					</div>
				</div>
			)}
		</div>
	);
}
