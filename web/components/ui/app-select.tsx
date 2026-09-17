"use client";

import { useId, type SelectHTMLAttributes } from "react";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

export interface AppSelectOption {
	value: string;
	label: string;
}

export interface AppSelectGroup {
	label: string;
	options: AppSelectOption[];
}

interface AppSelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, "onChange" | "value"> {
	value: string;
	onChange: (value: string) => void;
	/** Flat options. Ignored when `groups` is provided. */
	options?: AppSelectOption[];
	/** Grouped options, rendered as `<optgroup>`. Maps 1:1 to a future Radix `SelectGroup`. */
	groups?: AppSelectGroup[];
	placeholder?: string;
	isLoading?: boolean;
	loadingText?: string;
	variant?: "default" | "compact";
}

/**
 * Shared dropdown built on the native `<select>`.
 *
 * Deliberately native (not Radix/shadcn `Select`): inside bottom `Sheet`s on
 * mobile the OS picker is the better UX, and it avoids Portal/z-index and
 * focus-trap issues of a custom popover. The `groups` prop mirrors the Radix
 * `SelectGroup + SelectLabel` structure, so a future migration only swaps
 * this file's internals without touching call sites.
 */
export function AppSelect({
	value,
	onChange,
	options = [],
	groups,
	placeholder,
	isLoading = false,
	loadingText = "Loading...",
	variant = "default",
	className,
	disabled,
	id,
	...props
}: AppSelectProps) {
	const autoId = useId();
	const selectId = id ?? autoId;
	const showLoading = isLoading && loadingText;
	const isCompact = variant === "compact";

	return (
		<div className="relative">
			<select
				id={selectId}
				value={value}
				onChange={(e) => onChange(e.target.value)}
				disabled={disabled || isLoading}
				className={cn(
					isCompact
						? "w-full bg-secondary-50 border border-secondary-200 rounded-lg pl-2 pr-8 py-1.5 text-[11px] text-secondary-700 outline-none focus:border-primary-400 appearance-none cursor-pointer disabled:opacity-50"
						: "w-full bg-secondary-50 border border-secondary-200 rounded-xl pl-3.5 pr-10 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 appearance-none cursor-pointer disabled:opacity-50",
					className,
				)}
				{...props}
			>
				<option value="">{showLoading ? loadingText : (placeholder ?? "")}</option>
				{groups
					? groups.map((group) => (
							<optgroup key={group.label} label={group.label}>
								{group.options.map((opt) => (
									<option key={opt.value} value={opt.value}>
										{opt.label}
									</option>
								))}
							</optgroup>
						))
					: options.map((opt) => (
							<option key={opt.value} value={opt.value}>
								{opt.label}
							</option>
						))}
			</select>
			<ChevronDown
				size={isCompact ? 12 : 14}
				className="pointer-events-none absolute top-1/2 -translate-y-1/2 text-secondary-400 right-3"
				aria-hidden
			/>
		</div>
	);
}
