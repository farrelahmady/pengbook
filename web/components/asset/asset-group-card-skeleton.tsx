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
				<Skeleton className="w-10 h-10 rounded-xl shrink-0" />
				<div className="flex-1 space-y-1.5">
					<Skeleton className="h-[10px] w-[60px] rounded" />
					<Skeleton className="h-[14px] w-[120px] rounded" />
				</div>
				<div className="text-right space-y-1.5">
					<Skeleton className="h-[15px] w-[90px] rounded" />
					<Skeleton className="h-[10px] w-[50px] rounded" />
				</div>
			</div>
		</div>
	);
}
