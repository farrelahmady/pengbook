import { cn } from "@/lib/utils";

interface TopbarProps {
	subtitle?: string;
	right?: React.ReactNode;
	children?: React.ReactNode; // slot untuk summary strip, dll
	className?: string;
}

export function Topbar({ subtitle, right, children, className }: TopbarProps) {
	return (
		<div
			className={cn("bg-secondary-900 px-5 pt-12 pb-4", className)}
			style={{
				background: "linear-gradient(160deg, #1a1a2e 70%, #23309a 130%)",
			}}
		>
			<div className="flex items-start justify-between mb-1">
				<div>
					<h1 className="text-xl font-bold tracking-tight text-white leading-none">
						Peng<span className="text-primary-300">book</span>
					</h1>
					{subtitle && (
						<p className="text-[11px] text-white/40 font-medium mt-0.5 tracking-wide">
							{subtitle}
						</p>
					)}
				</div>
				{right && <div className="mt-0.5">{right}</div>}
			</div>
			{children}
		</div>
	);
}

// Period badge reusable
export function PeriodBadge({ label }: { label: string }) {
	return (
		<span
			className="font-mono text-[11px] px-3 py-1 rounded-full border text-white/60"
			style={{
				background: "rgba(255,255,255,0.08)",
				borderColor: "rgba(255,255,255,0.12)",
			}}
		>
			{label}
		</span>
	);
}

// Summary card di dalam topbar
export function TopbarSummaryCard({
	label,
	value,
	sub,
	accent,
}: {
	label: string;
	// Value is component so it can be skeleton loader or actual value
	value: React.ReactNode;
	sub?: string;
	accent?: boolean;
}) {
	return (
		<div
			className="flex-1 rounded-xl px-3 py-2.5"
			style={{
				background: "rgba(255,255,255,0.07)",
				border: "1px solid rgba(255,255,255,0.10)",
			}}
		>
			<p className="text-[10px] uppercase tracking-[0.5px] text-white/40 mb-1">
				{label}
			</p>
			<div
				className={cn(
					"font-mono text-[15px] font-medium leading-none",
					accent ? "text-primary-300" : "text-white",
				)}
			>
				{value}
			</div>
			{sub && <p className="font-mono text-[10px] text-white/30 mt-1">{sub}</p>}
		</div>
	);
}
