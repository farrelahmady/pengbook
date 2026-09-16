import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface AssetGroupCardSkeletonProps {
	className?: string;
}

export function AssetGroupCardSkeleton({ className }: AssetGroupCardSkeletonProps) {
	return (
		<div className={cn("card-default shadow-card", className)}>
			{/* Header */}
			<div className="flex items-center gap-3 px-4 py-3.5">
				<div className="flex-1 space-y-1.5">
					<Skeleton className="h-[10px] w-[60px] rounded" />
					<Skeleton className="h-[14px] w-[120px] rounded" />
				</div>
				<div className="text-right space-y-1.5">
					<Skeleton className="h-[15px] w-[90px] rounded" />
					<Skeleton className="h-[10px] w-[50px] rounded" />
				</div>
			</div>

			{/* Sub-account rows */}
			<div className="border-t border-black/[0.06]">
				{Array.from({ length: 3 }).map((_, i) => (
					<div
						key={i}
						className="flex items-center justify-between px-4 py-3 border-b border-black/[0.04] last:border-b-0"
					>
						<div className="flex items-center gap-2 min-w-0">
							<Skeleton className="w-0.5 h-4 rounded-full shrink-0" />
							<div className="space-y-1">
								<Skeleton className="h-[10px] w-[80px] rounded" />
								<Skeleton className="h-[13px] w-[140px] rounded" />
							</div>
						</div>
						<div className="text-right shrink-0 space-y-1">
							<Skeleton className="h-[13px] w-[80px] rounded" />
							<Skeleton className="h-[16px] w-[32px] rounded" />
						</div>
					</div>
				))}

				{/* Subtotal */}
				<div className="flex items-center justify-between px-4 py-2.5 bg-secondary-50/50 border-t border-black/[0.06]">
					<Skeleton className="h-[11px] w-[120px] rounded" />
					<Skeleton className="h-[12px] w-[80px] rounded" />
				</div>
			</div>
		</div>
	);
}
