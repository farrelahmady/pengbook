import type { ComponentProps } from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

interface AppInputProps extends ComponentProps<"input"> {
	variant?: "default" | "compact";
}

/**
 * App-themed input built on the shadcn `Input` primitive.
 *
 * Overrides the stock shadcn sizing (`h-8 text-xs rounded-none`) with the
 * project's form tokens (`rounded-xl bg-secondary-50 text-[13px]`), so every
 * form shares one style definition instead of a copy-pasted `inputClass`.
 */
export function AppInput({ className, variant = "default", ...props }: AppInputProps) {
	return (
		<Input
			className={cn(
				variant === "compact"
					? "h-auto w-full bg-secondary-50 border-secondary-200 rounded-lg px-2 py-1.5 text-[11px] text-secondary-700 placeholder:text-secondary-300 focus-visible:border-primary-400 focus-visible:ring-0"
					: "h-auto w-full bg-secondary-50 border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 placeholder:text-secondary-300 focus-visible:border-primary-400 focus-visible:ring-0",
				className,
			)}
			{...props}
		/>
	);
}
