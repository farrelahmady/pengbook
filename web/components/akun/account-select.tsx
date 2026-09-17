"use client";

import type { ReactNode } from "react";
import { AppField } from "@/components/ui/app-field";
import { AppSelect } from "@/components/ui/app-select";
import { groupByType } from "@/lib/group-accounts";

/** Minimal shape shared by `PostingAccount` and `ParentListItem`. */
export interface AccountSelectItem {
	id: number;
	code: string;
	name: string;
	type: string;
}

interface AccountSelectProps {
	/** Selected account id as string. Convert with `Number()` at the submit boundary. */
	value: string;
	onChange: (value: string) => void;
	accounts: AccountSelectItem[];
	placeholder: string;
	/** When provided, the select is wrapped in an `AppField` label. */
	label?: ReactNode;
	hint?: ReactNode;
	isLoading?: boolean;
	loadingText?: string;
	/** Group options by account type (`<optgroup>`). Defaults to true. */
	grouped?: boolean;
	/** `code-name` renders "code · name"; `code` renders code only (compact rows). */
	display?: "code-name" | "code";
	variant?: "default" | "compact";
	disabled?: boolean;
	id?: string;
}

/**
 * Domain dropdown for picking a chart-of-accounts entry.
 *
 * Replaces the per-form `groupedParents` Map blocks and
 * `{code} · {name}` option rendering that were duplicated across the journal,
 * account, and asset forms.
 */
export function AccountSelect({
	value,
	onChange,
	accounts,
	placeholder,
	label,
	hint,
	isLoading = false,
	loadingText,
	grouped = true,
	display = "code-name",
	variant = "default",
	disabled,
	id,
}: AccountSelectProps) {
	const toOption = (a: AccountSelectItem) => ({
		value: String(a.id),
		label: display === "code" ? a.code : `${a.code} · ${a.name}`,
	});

	const select = (
		<AppSelect
			id={id}
			value={value}
			onChange={onChange}
			placeholder={placeholder}
			isLoading={isLoading}
			loadingText={loadingText}
			variant={variant}
			disabled={disabled}
			groups={grouped ? groupByType(accounts).map(([type, list]) => ({ label: type, options: list.map(toOption) })) : undefined}
			options={grouped ? undefined : accounts.map(toOption)}
		/>
	);

	if (label === undefined) {
		return select;
	}

	return (
		<AppField label={label} hint={hint}>
			{select}
		</AppField>
	);
}
