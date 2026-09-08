"use client";

import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { cn } from "@/lib/utils";
import { useState, ReactNode } from "react";
import { ChevronDown } from "lucide-react";

interface ReportCardProps {
	icon: ReactNode;
	iconBg: string;
	name: string;
	subtitle: string;
	value: React.ReactNode;
	valueColor?: string;
	children?: ReactNode;
	defaultExpanded?: boolean;
}

export function ReportCard({
	icon,
	iconBg,
	name,
	subtitle,
	value,
	valueColor = "text-success-600",
	children,
	defaultExpanded = false,
}: ReportCardProps) {
	const [expanded, setExpanded] = useState(defaultExpanded);

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
						iconBg,
					)}
				>
					{icon}
				</div>
				<div className="flex-1 min-w-0">
					<p className="text-[14px] font-semibold text-secondary-900 leading-snug">
						{name}
					</p>
					<p className="text-[11px] text-secondary-400">{subtitle}</p>
				</div>
				<div className="text-right shrink-0">
					<p className={cn("font-mono text-[13px] font-semibold", valueColor)}>
						{value}
					</p>
				</div>
				<ChevronDown
					size={16}
					className={cn(
						"text-secondary-400 transition-transform duration-200 shrink-0",
						expanded ? "rotate-180" : "",
					)}
				/>
			</button>

			{/* Content */}
			{expanded && children && (
				<div className="border-t border-black/[0.06] px-4 py-3">
					{children}
				</div>
			)}
		</div>
	);
}
