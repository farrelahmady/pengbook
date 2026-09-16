"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { accountService } from "@/services/account";
import { createLogger } from "@/lib/logger";
import { queryKeys } from "@/lib/query-keys";
import { AccountWithChildren } from "@/types";

const logger = createLogger("account.edit-sheet");

interface EditAccountSheetProps {
	account: AccountWithChildren | null;
	open: boolean;
	onOpenChange: (v: boolean) => void;
}

export function EditAccountSheet({
	account,
	open,
	onOpenChange,
}: EditAccountSheetProps) {
	const t = useTranslations("accountPage.edit");
	const queryClient = useQueryClient();
	const [name, setName] = useState(account?.name ?? "");
	const [isSubmitting, setIsSubmitting] = useState(false);

	async function handleSubmit() {
		if (!account) return;
		if (!name.trim()) {
			toast.error(t("toastValidation"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);
		logger.debug("Updating account", { id: account.id, name });

		try {
			const updated = await accountService.update(account.id, {
				name: name.trim(),
			});

			toast.success(t("toastSuccess", { code: updated.code }), {
				id: toastId,
			});
			logger.info("Account updated", { id: updated.id, code: updated.code });
			queryClient.invalidateQueries({ queryKey: queryKeys.accounts.all });
			onOpenChange(false);
		} catch (err) {
			logger.error("Failed to update account", { error: err });
			toast.error(t("toastError"), { id: toastId });
		} finally {
			setIsSubmitting(false);
		}
	}

	const labelClass =
		"text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent
				side="bottom"
				className="p-0 rounded-t-3xl max-w-[390px] mx-auto max-h-[92dvh] overflow-y-auto"
			>
				{/* Handle */}
				<div className="sheet-handle" />

				<div className="px-5 pt-3 pb-8">
					<h2 className="text-[17px] font-bold text-secondary-900 mb-4">
						{t("title")}
					</h2>

					<div className="flex flex-col gap-4">
						{/* Code (read-only) */}
						<div>
							<span className={labelClass}>{t("code")}</span>
							<p className="font-mono text-[13px] text-secondary-500">
								{account?.code}
							</p>
						</div>

						{/* Name */}
						<div>
							<label className={labelClass}>{t("name")}</label>
							<input
								type="text"
								value={name}
								maxLength={200}
								onChange={(e) => setName(e.target.value)}
								className="w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 placeholder:text-secondary-300"
							/>
						</div>

						<button
							type="button"
							onClick={handleSubmit}
							disabled={isSubmitting}
							className="w-full py-3.5 rounded-2xl bg-primary-500 text-white font-semibold text-[14px] shadow-[0_4px_20px_rgba(59,79,212,0.4)] active:scale-[0.98] transition-transform no-tap disabled:opacity-50"
						>
							{t("submit")}
						</button>
					</div>
				</div>
			</SheetContent>
		</Sheet>
	);
}
