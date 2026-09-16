"use client";

import { AssetGroup } from "@/types";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { ChevronUp } from "lucide-react";

interface AssetGroupCardProps {
	group: AssetGroup & { matchCount?: number };
	expanded: boolean;
	onToggle: () => void;
}

export function AssetGroupCard({ group, expanded, onToggle }: AssetGroupCardProps) {
	const t = useTranslations("assetPage");
	const currencyFormat = useCurrencyFormatter();

	return (
		<div className="card-default shadow-card overflow-hidden">
			{/* Header */}
			<button
				onClick={onToggle}
				aria-expanded={expanded}
				aria-controls={`asset-group-${group.id}`}
				className="w-full flex items-center gap-3 px-4 py-3.5 text-left active:bg-secondary-50 transition-colors"
			>
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
						{t("groupSubtitle", { count: group.accounts.length })}
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
				<div
					id={`asset-group-${group.id}`}
					className="border-t border-black/[0.06]"
				>
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
											{t("badge.posting")}
										</span>
									)}
								</div>
							</div>
						))}
					</div>

					{/* Subtotal */}
					<div className="flex items-center justify-between px-4 py-2.5 bg-secondary-50/50 border-t border-black/[0.06]">
						<p className="text-[11px] text-secondary-400">
							{t("subtotalLabel")}
						</p>
						<p className="font-mono text-[12px] font-medium text-secondary-600">
							{t("subtotalValue", {
								sign: "DR",
								value: currencyFormat(group.totalBalance, { compact: true, currency: "Rp" }),
							})}
						</p>
					</div>
				</div>
			)}
		</div>
	);
}
