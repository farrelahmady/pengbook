import type { ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

interface AppSubmitButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	/**
	 * - `sheet`: bottom-sheet primary submit (rounded-2xl + shadow).
	 * - `form`: inline-form primary submit (rounded-xl).
	 * - `stateful`: blue when `ready`, muted grey otherwise (validity-driven forms).
	 */
	variant?: "sheet" | "form" | "stateful";
	/** Only used with `stateful`: whether the form is valid submittable. */
	ready?: boolean;
}

/**
 * Shared primary submit button.
 *
 * Centralizes the three submit-button styles copy-pasted across journal,
 * account, and asset forms. Intentionally a styled native `<button>` rather
 * than the shadcn `Button` cva: these styles (py-3.5, rounded-2xl, glow
 * shadow) diverge far from the stock `h-8 rounded-none text-xs` tokens, so
 * overriding the cva would be more fragile than one explicit definition.
 */
export function AppSubmitButton({
	variant = "sheet",
	ready = true,
	className,
	type = "button",
	...props
}: AppSubmitButtonProps) {
	if (variant === "stateful") {
		return (
			<button
				type={type}
				className={cn(
					"w-full font-semibold text-[15px] py-3.5 rounded-xl transition-all mt-1",
					ready
						? "bg-primary-500 text-white active:scale-[0.98]"
						: "bg-secondary-100 text-secondary-400 cursor-not-allowed",
					className,
				)}
				{...props}
			/>
		);
	}

	return (
		<button
			type={type}
			className={cn(
				variant === "form"
					? "w-full bg-primary-500 text-white font-semibold text-[15px] py-3.5 rounded-xl active:scale-[0.98] transition-transform mt-1 disabled:opacity-50 disabled:cursor-not-allowed"
					: "w-full py-3.5 rounded-2xl bg-primary-500 text-white font-semibold text-[14px] shadow-[0_4px_20px_rgba(59,79,212,0.4)] active:scale-[0.98] transition-transform no-tap disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}
