import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface ReportCardSkeletonProps {
	className?: string;
}

export function ReportCardSkeleton({ className }: ReportCardSkeletonProps) {
	return (
		<div className={cn("card-default shadow-card", className)}>
			<div className="flex items-center gap-3 px-4 py-3.5">
				<Skeleton className="w-10 h-10 rounded-xl shrink-0" />
				<div className="flex-1 space-y-1.5">
					<Skeleton className="h-[14px] w-[120px] rounded" />
					<Skeleton className="h-[11px] w-[80px] rounded" />
				</div>
				<Skeleton className="h-[13px] w-[90px] rounded" />
			</div>
		</div>
	);
}
