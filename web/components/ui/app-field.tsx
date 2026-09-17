import { cloneElement, isValidElement, useId, type ReactElement, type ReactNode } from "react";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface AppFieldProps {
	label: ReactNode;
	children: ReactNode;
	/** Overrides the auto-generated id linkage. */
	htmlFor?: string;
	hint?: ReactNode;
	error?: ReactNode;
	className?: string;
}

/**
 * Shared label + control + hint/error wrapper.
 *
 * Standardizes the `labelClass` string that was copy-pasted across journal,
 * account, and asset forms, and wires `label ↔ control` via `htmlFor`/`id`
 * for screen readers and larger mobile tap targets.
 */
export function AppField({ label, children, htmlFor, hint, error, className }: AppFieldProps) {
	const autoId = useId();
	const fieldId = htmlFor ?? autoId;
	const hintId = hint ? `${fieldId}-hint` : undefined;
	const errorId = error ? `${fieldId}-error` : undefined;

	let control = children;
	let linked = false;
	if (isValidElement(children)) {
		const childProps = (children as ReactElement<{ id?: string }>).props;
		control = cloneElement(
			children as ReactElement<{ id?: string; "aria-describedby"?: string }>,
			{
				id: childProps.id ?? fieldId,
				"aria-describedby": [hintId, errorId].filter(Boolean).join(" ") || undefined,
			},
		);
		linked = true;
	}

	return (
		<div className={className}>
			<Label
				htmlFor={htmlFor ?? (linked ? fieldId : undefined)}
				className="text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block"
			>
				{label}
			</Label>
			{control}
			{hint && !error && (
				<p id={hintId} className="text-[11px] text-secondary-400 mt-1.5">
					{hint}
				</p>
			)}
			{error && (
				<p id={errorId} className={cn("text-[11px] text-danger-600 mt-1.5")}>
					{error}
				</p>
			)}
		</div>
	);
}
