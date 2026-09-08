import { cn } from "@/lib/utils";

export function DrBadge({ className }: { className?: string }) {
  return <span className={cn("badge-dr", className)}>DR</span>;
}

export function CrBadge({ className }: { className?: string }) {
  return <span className={cn("badge-cr", className)}>CR</span>;
}

export function DrCrDot({ type }: { type: "dr" | "cr" }) {
  return (
    <span
      className={cn(
        "inline-block w-1.5 h-1.5 rounded-full shrink-0",
        type === "dr" ? "bg-danger-500" : "bg-success-500"
      )}
    />
  );
}