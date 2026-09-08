"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { journalService } from "@/services/journal";
import { POSTING_ACCOUNTS } from "@/lib/constants";
import { JournalEntry } from "@/types";

interface EditBasicFormProps {
	journal: JournalEntry;
	onSuccess: () => void;
}

export function EditBasicForm({ journal, onSuccess }: EditBasicFormProps) {
	const t = useTranslations("journalPage.basic");
	const queryClient = useQueryClient();

	// Parse existing journal lines to get from/to accounts
	const existingLines = journal.lines;
	const debitLine = existingLines.find((l) => parseFloat(l.debit) > 0);
	const creditLine = existingLines.find((l) => parseFloat(l.credit) > 0);

	const [date, setDate] = useState(journal.date.slice(0, 10));
	const [description, setDesc] = useState(journal.description ?? "");
	const [fromAccount, setFrom] = useState(creditLine?.accountId ?? "");
	const [toAccount, setTo] = useState(debitLine?.accountId ?? "");
	const [amount, setAmount] = useState(
		debitLine ? String(parseFloat(debitLine.debit)) : "",
	);
	const [isSubmitting, setIsSubmitting] = useState(false);

	function handleSubmit() {
		if (!date || !fromAccount || !toAccount || !amount) {
			toast.error(t("toastValidation"));
			return;
		}
		if (fromAccount === toAccount) {
			toast.error(t("toastSameAccount"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);

		journalService
			.update(journal.id, {
				date: new Date(date).toISOString(),
				description: description || undefined,
				lines: [
					{
						accountId: toAccount,
						debit: parseFloat(amount),
						credit: 0,
					},
					{
						accountId: fromAccount,
						debit: 0,
						credit: parseFloat(amount),
					},
				],
			})
			.then(() => {
				toast.success(t("toastSuccess"), { id: toastId });
				queryClient.invalidateQueries({ queryKey: ["journals"] });
				queryClient.invalidateQueries({ queryKey: ["journalSummary"] });
				onSuccess();
			})
			.catch(() => {
				toast.error(t("toastError"), { id: toastId });
			})
			.finally(() => {
				setIsSubmitting(false);
			});
	}

	const labelClass =
		"text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";
	const inputClass =
		"bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 w-full";

	return (
		<div className="flex flex-col gap-4">
			{/* Date */}
			<div>
				<label className={labelClass}>{t("date")}</label>
				<input
					type="date"
					value={date}
					onChange={(e) => setDate(e.target.value)}
					className={inputClass}
				/>
			</div>

			{/* Description */}
			<div>
				<label className={labelClass}>{t("description")}</label>
				<input
					type="text"
					placeholder={t("descriptionPlaceholder")}
					value={description}
					onChange={(e) => setDesc(e.target.value)}
					className={inputClass}
				/>
			</div>

			{/* From Account */}
			<div>
				<label className={labelClass}>{t("fromAccount")}</label>
				<select
					value={fromAccount}
					onChange={(e) => setFrom(e.target.value)}
					className={`${inputClass} appearance-none`}
				>
					<option value="">{t("fromPlaceholder")}</option>
					{POSTING_ACCOUNTS.map((a) => (
						<option key={a.id} value={a.id}>
							{a.code} · {a.name}
						</option>
					))}
				</select>
			</div>

			{/* Arrow */}
			<div className="flex justify-center py-1">
				<div className="w-8 h-8 rounded-full bg-primary-50 flex items-center justify-center">
					<ArrowRight size={16} className="text-primary-500 rotate-90" />
				</div>
			</div>

			{/* To Account */}
			<div>
				<label className={labelClass}>{t("toAccount")}</label>
				<select
					value={toAccount}
					onChange={(e) => setTo(e.target.value)}
					className={`${inputClass} appearance-none`}
				>
					<option value="">{t("toPlaceholder")}</option>
					{POSTING_ACCOUNTS.map((a) => (
						<option key={a.id} value={a.id}>
							{a.code} · {a.name}
						</option>
					))}
				</select>
			</div>

			{/* Amount */}
			<div>
				<label className={labelClass}>{t("amount")}</label>
				<input
					type="number"
					placeholder="0"
					value={amount}
					onChange={(e) => setAmount(e.target.value)}
					className={`${inputClass} font-mono`}
				/>
			</div>

			{/* Hint */}
			<div className="bg-secondary-50 rounded-xl px-4 py-3 border border-secondary-100">
				<p className="text-[11px] font-semibold text-secondary-500 mb-1">
					{t("hint")}
				</p>
				<p className="text-[11px] text-secondary-400">{t("hintDetail")}</p>
			</div>

			{/* Submit */}
			<button
				onClick={handleSubmit}
				disabled={!date || !fromAccount || !toAccount || !amount || isSubmitting}
				className={`w-full font-semibold text-[15px] py-3.5 rounded-xl transition-all mt-1 ${
					date && fromAccount && toAccount && amount && !isSubmitting
						? "bg-primary-500 text-white active:scale-[0.98]"
						: "bg-secondary-100 text-secondary-400 cursor-not-allowed"
				}`}
			>
				{t("submit")}
			</button>
		</div>
	);
}
