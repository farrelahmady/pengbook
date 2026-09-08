import { cn } from "@/lib/utils";
import { LucideIcon } from "lucide-react";

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center py-16 px-8 text-center",
        className,
      )}
    >
      <div className="w-14 h-14 rounded-2xl bg-secondary-100 flex items-center justify-center mb-4">
        <Icon size={24} className="text-secondary-400" strokeWidth={1.5} />
      </div>
      <p className="text-[14px] font-semibold text-secondary-700 mb-1">
        {title}
      </p>
      {description && (
        <p className="text-[12px] text-secondary-400 leading-relaxed">
          {description}
        </p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
