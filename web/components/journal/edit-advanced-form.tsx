"use client";

import { useState } from "react";
import { Plus, Minus } from "lucide-react";
import { dateInputToISO, formatNumber, parseDecimal } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient, useQuery } from "@tanstack/react-query";
import { journalService } from "@/services/journal";
import { accountService } from "@/services/account";
import { JournalEntryListItem } from "@/types";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";
import { AppField } from "@/components/ui/app-field";
import { AppInput } from "@/components/ui/app-input";
import { AppSubmitButton } from "@/components/ui/app-submit-button";
import { AccountSelect } from "@/components/akun/account-select";

const logger = createLogger("EditAdvancedForm");

interface JournalLine {
	accountId: string;
	debit: string;
	credit: string;
}

interface EditAdvancedFormProps {
	journal: JournalEntryListItem;
	onSuccess: () => void;
}

export function EditAdvancedForm({ journal, onSuccess }: EditAdvancedFormProps) {
	const t = useTranslations("journalPage.advanced");
	const queryClient = useQueryClient();

	const { data: postingAccounts = [], isLoading: isLoadingAccounts } = useQuery({
		queryKey: queryKeys.accounts.posting,
		queryFn: () => accountService.getPostingAccounts(),
	});

	const [date, setDate] = useState(journal.datetime.slice(0, 10));
	const [description, setDesc] = useState(journal.description ?? "");
	const [lines, setLines] = useState<JournalLine[]>(
		journal.lines.map((l) => ({
			accountId: l.accountId?.toString() ?? "",
			debit: l.debit !== 0 ? String(l.debit) : "",
			credit: l.credit !== 0 ? String(l.credit) : "",
		})),
	);
	const [isSubmitting, setIsSubmitting] = useState(false);

	const totalDr = lines.reduce((s, l) => s + parseDecimal(l.debit), 0);
	const totalCr = lines.reduce((s, l) => s + parseDecimal(l.credit), 0);
	const isBalanced = totalDr === totalCr && totalDr > 0;

	function updateLine(i: number, field: keyof JournalLine, value: string) {
		setLines((prev) =>
			prev.map((l, idx) => (idx === i ? { ...l, [field]: value } : l)),
		);
	}

	function addLine() {
		setLines((prev) => [...prev, { accountId: "", debit: "", credit: "" }]);
	}

	function removeLine(i: number) {
		if (lines.length <= 2) return;
		setLines((prev) => prev.filter((_, idx) => idx !== i));
	}

	async function handleSubmit() {
		if (!isBalanced) {
			toast.error(t("toastUnbalanced"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);

		try {
			await journalService.update(journal.id, {
				date: dateInputToISO(date),
				description: description || undefined,
				lines: lines.map((l) => ({
					accountId: Number(l.accountId),
					debit: parseDecimal(l.debit),
					credit: parseDecimal(l.credit),
				})),
			});

			toast.success(t("toastSuccess"), { id: toastId });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
			onSuccess();
		} catch {
			toast.error(t("toastError"), { id: toastId });
		} finally {
			setIsSubmitting(false);
		}
	}

	return (
		<div className="flex flex-col gap-4">
			{/* Date */}
			<AppField label={t("date")}>
				<AppInput
					type="date"
					value={date}
					onChange={(e) => setDate(e.target.value)}
				/>
			</AppField>

			{/* Description */}
			<AppField label={t("description")}>
				<AppInput
					type="text"
					placeholder={t("descriptionPlaceholder")}
					value={description}
					onChange={(e) => setDesc(e.target.value)}
				/>
			</AppField>

			{/* Journal lines table */}
			<AppField label={t("lines")}>
				<div className="border border-secondary-200 rounded-xl overflow-hidden">
					{/* Head */}
					<div className="grid grid-cols-[1fr_80px_80px_28px] gap-1 px-3 py-2 bg-secondary-50 border-b border-secondary-200">
						{[t("account"), t("debit"), t("credit"), ""].map((h, i) => (
							<span
								key={i}
								className={cn(
									"text-[10px] font-semibold uppercase tracking-[0.4px] text-secondary-400",
									i > 0 && "text-right",
								)}
							>
								{h}
							</span>
						))}
					</div>

					{/* Rows */}
					{lines.map((line, i) => (
						<div
							key={i}
							className="grid grid-cols-[1fr_80px_80px_28px] gap-1 px-2 py-2 border-b border-secondary-100 last:border-none items-center"
						>
						<AccountSelect
							value={line.accountId}
							onChange={(v) => updateLine(i, "accountId", v)}
							accounts={postingAccounts}
							isLoading={isLoadingAccounts}
							placeholder={t("accountPlaceholder")}
							loadingText="Loading..."
							grouped={false}
							display="code"
							variant="compact"
						/>

							<input
								type="number"
								placeholder="0"
								value={line.debit}
								onChange={(e) => updateLine(i, "debit", e.target.value)}
								className="text-[11px] font-mono bg-secondary-50 border border-secondary-200 rounded-lg px-2 py-1.5 text-right w-full outline-none focus:border-success-400 text-success-700"
							/>
							<input
								type="number"
								placeholder="0"
								value={line.credit}
								onChange={(e) => updateLine(i, "credit", e.target.value)}
								className="text-[11px] font-mono bg-secondary-50 border border-secondary-200 rounded-lg px-2 py-1.5 text-right w-full outline-none focus:border-danger-400 text-danger-700"
							/>

							<button
								onClick={() => removeLine(i)}
								disabled={lines.length <= 2}
								className="w-7 h-7 rounded-lg bg-danger-50 flex items-center justify-center text-danger-500 disabled:opacity-30"
							>
								<Minus size={13} />
							</button>
						</div>
					))}
				</div>

				{/* Add row */}
				<button
					onClick={addLine}
					className="w-full mt-2 flex items-center justify-center gap-2 border-2 border-dashed border-secondary-200
                     rounded-xl py-2.5 text-[13px] font-medium text-secondary-400 hover:border-primary-300
                     hover:text-primary-500 transition-colors"
				>
					<Plus size={14} />
					{t("addRow")}
				</button>
			</AppField>

			{/* Balance strip */}
			<div
				className={cn(
					"rounded-xl px-4 py-3 flex items-center justify-between border",
					isBalanced
						? "bg-success-50 border-success-200"
						: "bg-secondary-50 border-secondary-200",
				)}
			>
				<span className="text-[12px] font-semibold text-secondary-600">
					{t("balance")}
				</span>
				<div className="flex gap-4 items-center">
					<div className="text-right">
						<p className="text-[9px] uppercase tracking-wide text-secondary-400">
							{t("debit")}
						</p>
						<p className="font-mono text-[12px] font-semibold text-success-700">
							{formatNumber(totalDr)}
						</p>
					</div>
					<div className="text-right">
						<p className="text-[9px] uppercase tracking-wide text-secondary-400">
							{t("credit")}
						</p>
						<p className="font-mono text-[12px] font-semibold text-danger-700">
							{formatNumber(totalCr)}
						</p>
					</div>
					<div
						className={cn(
							"text-[12px] font-bold font-mono",
							isBalanced ? "text-success-600" : "text-secondary-400",
						)}
					>
						{totalDr === 0 && totalCr === 0
							? "—"
							: isBalanced
								? t("balanced")
								: `Δ ${formatNumber(Math.abs(totalDr - totalCr))}`}
					</div>
				</div>
			</div>

			{/* Submit */}
			<AppSubmitButton
				variant="stateful"
				ready={isBalanced && !isSubmitting}
				onClick={handleSubmit}
				disabled={!isBalanced || isSubmitting}
			>
				{t("submit")}
			</AppSubmitButton>
		</div>
	);
}
