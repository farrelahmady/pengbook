"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient, useQuery } from "@tanstack/react-query";
import { journalService } from "@/services/journal";
import { accountService } from "@/services/account";
import { JournalEntryListItem } from "@/types";
import { dateInputToISO } from "@/lib/utils";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";
import { AppField } from "@/components/ui/app-field";
import { AppInput } from "@/components/ui/app-input";
import { AppSubmitButton } from "@/components/ui/app-submit-button";
import { AccountSelect } from "@/components/akun/account-select";

const logger = createLogger("EditBasicForm");

interface EditBasicFormProps {
	journal: JournalEntryListItem;
	onSuccess: () => void;
}

export function EditBasicForm({ journal, onSuccess }: EditBasicFormProps) {
	const t = useTranslations("journalPage.basic");
	const queryClient = useQueryClient();

	const { data: postingAccounts = [], isLoading: isLoadingAccounts } = useQuery({
		queryKey: queryKeys.accounts.posting,
		queryFn: () => accountService.getPostingAccounts(),
	});

	// Parse existing journal lines to get from/to accounts
	const existingLines = journal.lines;
	const debitLine = existingLines.find((l) => l.debit > 0);
	const creditLine = existingLines.find((l) => l.credit > 0);

	const [date, setDate] = useState(journal.datetime.slice(0, 10));
	const [description, setDesc] = useState(journal.description ?? "");
	const [fromAccount, setFrom] = useState(creditLine?.accountId?.toString() ?? "");
	const [toAccount, setTo] = useState(debitLine?.accountId?.toString() ?? "");
	const [amount, setAmount] = useState(
		debitLine ? String(debitLine.debit) : "",
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
				date: dateInputToISO(date),
				description: description || undefined,
				lines: [
					{
						accountId: Number(toAccount),
						debit: parseFloat(amount),
						credit: 0,
					},
					{
						accountId: Number(fromAccount),
						debit: 0,
						credit: parseFloat(amount),
					},
				],
			})
			.then(() => {
				toast.success(t("toastSuccess"), { id: toastId });
				queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
				queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
				onSuccess();
			})
			.catch(() => {
				toast.error(t("toastError"), { id: toastId });
			})
			.finally(() => {
				setIsSubmitting(false);
			});
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

			{/* From Account */}
			<AccountSelect
				label={t("fromAccount")}
				value={fromAccount}
				onChange={setFrom}
				accounts={postingAccounts}
				isLoading={isLoadingAccounts}
				placeholder={t("fromPlaceholder")}
				loadingText="Loading..."
			/>

			{/* Arrow */}
			<div className="flex justify-center py-1">
				<div className="w-8 h-8 rounded-full bg-primary-50 flex items-center justify-center">
					<ArrowRight size={16} className="text-primary-500 rotate-90" />
				</div>
			</div>

			{/* To Account */}
			<AccountSelect
				label={t("toAccount")}
				value={toAccount}
				onChange={setTo}
				accounts={postingAccounts}
				isLoading={isLoadingAccounts}
				placeholder={t("toPlaceholder")}
				loadingText="Loading..."
			/>

			{/* Amount */}
			<AppField label={t("amount")}>
				<AppInput
					type="number"
					placeholder="0"
					value={amount}
					onChange={(e) => setAmount(e.target.value)}
					className="font-mono"
				/>
			</AppField>

			{/* Hint */}
			<div className="bg-secondary-50 rounded-xl px-4 py-3 border border-secondary-100">
				<p className="text-[11px] font-semibold text-secondary-500 mb-1">
					{t("hint")}
				</p>
				<p className="text-[11px] text-secondary-400">{t("hintDetail")}</p>
			</div>

			{/* Submit */}
			<AppSubmitButton
				variant="stateful"
				ready={Boolean(date && fromAccount && toAccount && amount && !isSubmitting)}
				onClick={handleSubmit}
				disabled={!date || !fromAccount || !toAccount || !amount || isSubmitting}
			>
				{t("submit")}
			</AppSubmitButton>
		</div>
	);
}
