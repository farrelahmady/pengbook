"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient, useQuery } from "@tanstack/react-query";
import { journalService } from "@/services/journal";
import { accountService } from "@/services/account";
import { dateInputToISO, parseDecimal, todayLocalDate } from "@/lib/utils";
import { queryKeys } from "@/lib/query-keys";
import { AppField } from "@/components/ui/app-field";
import { AppInput } from "@/components/ui/app-input";
import { AppSubmitButton } from "@/components/ui/app-submit-button";
import { AccountSelect } from "@/components/akun/account-select";

interface BasicFormProps {
	onSuccess: () => void;
}

export function BasicForm({ onSuccess }: BasicFormProps) {
	const t = useTranslations("journalPage.basic");
	const queryClient = useQueryClient();
	const [date, setDate] = useState(todayLocalDate());
	const [description, setDesc] = useState("");
	const [fromAccount, setFrom] = useState("");
	const [toAccount, setTo] = useState("");
	const [amount, setAmount] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	const { data: postingAccounts = [], isLoading: isLoadingAccounts } = useQuery(
		{
			queryKey: queryKeys.accounts.posting,
			queryFn: () => accountService.getPostingAccounts(),
		},
	);

	async function handleSubmit() {
		if (!fromAccount || !toAccount || !amount) {
			toast.error(t("toastValidation"));
			return;
		}
		if (fromAccount === toAccount) {
			toast.error(t("toastSameAccount"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);

		try {
			await journalService.create({
				date: dateInputToISO(date),
				description: description || undefined,
				lines: [
					{
						accountId: Number(toAccount),
						debit: parseDecimal(amount),
						credit: 0,
					},
					{
						accountId: Number(fromAccount),
						debit: 0,
						credit: parseDecimal(amount),
					},
				],
			});

			toast.success(t("toastSuccess"), { id: toastId });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
			onSuccess();
		} catch (err) {
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

			{/* From */}
			<AccountSelect
				label={t("fromAccount")}
				value={fromAccount}
				onChange={setFrom}
				accounts={postingAccounts}
				isLoading={isLoadingAccounts}
				placeholder={t("fromPlaceholder")}
				loadingText="Loading..."
			/>

			{/* Arrow visual */}
			<div className="flex items-center gap-3 -my-1">
				<div className="flex-1 border-t-2 border-dashed border-secondary-200" />
				<div className="w-8 h-8 rounded-full bg-primary-50 flex items-center justify-center shrink-0">
					<ArrowRight size={14} className="text-primary-500" />
				</div>
				<div className="flex-1 border-t-2 border-dashed border-secondary-200" />
			</div>

			{/* To */}
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
				<div className="relative">
					<span className="absolute left-3.5 top-1/2 -translate-y-1/2 font-mono text-[13px] text-secondary-400">
						Rp
					</span>
					<AppInput
						type="number"
						placeholder="0"
						value={amount}
						onChange={(e) => setAmount(e.target.value)}
						className="pl-10 font-mono text-[16px]"
					/>
				</div>
			</AppField>

			{/* Auto hint */}
			<div className="bg-primary-50 rounded-xl px-3.5 py-3 text-[12px] text-primary-700 leading-relaxed">
				<span className="font-semibold">{t("hint")}</span>
				<br />
				<span className="font-mono text-[11px]">{t("hintDetail")}</span>
			</div>

			{/* Submit */}
			<AppSubmitButton
				variant="form"
				onClick={handleSubmit}
				disabled={isSubmitting}
			>
				{t("submit")}
			</AppSubmitButton>
		</div>
	);
}
