"use client";

import { useState } from "react";
import { Info } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { assetService } from "@/services/asset";
import { createLogger } from "@/lib/logger";
import { queryKeys } from "@/lib/query-keys";

const logger = createLogger("asset.adjust-sheet");

interface AdjustBalanceSheetProps {
	open: boolean;
	onOpenChange: (v: boolean) => void;
	asset: {
		id: number;
		code: string;
		name: string;
		balance: number;
	} | null;
}

export function AdjustBalanceSheet({
	open,
	onOpenChange,
	asset,
}: AdjustBalanceSheetProps) {
	const t = useTranslations("assetPage.adjustBalance");
	const queryClient = useQueryClient();
	const [balance, setBalance] = useState(asset?.balance.toString() ?? "");
	const [note, setNote] = useState("");
	const [date, setDate] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	async function handleSubmit() {
		if (!asset) return;

		const balanceNum = parseFloat(balance);
		if (!balance || isNaN(balanceNum) || balanceNum < 0) {
			toast.error(t("toastValidation"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);
		logger.debug("Adjusting asset balance", {
			assetId: asset.id,
			desiredBalance: balanceNum,
			note,
		});

		try {
			const result = await assetService.adjustBalance(
				asset.id,
				balanceNum,
				note || undefined,
				date || undefined,
			);

			toast.success(
				t("toastSuccess", { balance: result.balance.toLocaleString("id-ID") }),
				{ id: toastId },
			);
			logger.info("Asset balance adjusted", {
				assetId: asset.id,
				journalEntryId: result.journalEntryId,
				newBalance: result.balance,
			});

			// Invalidate asset queries
			queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });

			setBalance("");
			setNote("");
			setDate("");
			onOpenChange(false);
		} catch (err) {
			logger.error("Failed to adjust asset balance", { error: err });
			toast.error(t("toastError"), { id: toastId });
		} finally {
			setIsSubmitting(false);
		}
	}

	// Reset form when sheet opens with new asset
	const handleOpenChange = (v: boolean) => {
		if (v && asset) {
			setBalance(asset.balance.toString());
			setNote("");
			setDate("");
		}
		onOpenChange(v);
	};

	if (!asset) return null;

	const labelClass =
		"text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";

	return (
		<Sheet open={open} onOpenChange={handleOpenChange}>
			<SheetContent
				side="bottom"
				className="p-0 rounded-t-3xl max-w-[390px] mx-auto max-h-[92dvh] overflow-y-auto"
			>
				{/* Handle */}
				<div className="sheet-handle" />

				<div className="px-5 pt-3 pb-8">
					<h2 className="text-[17px] font-bold text-secondary-900 mb-1">
						{t("title")}
					</h2>
					<p className="text-[13px] text-secondary-500 mb-4">
						{asset.code} — {asset.name}
					</p>

					<div className="flex flex-col gap-4">
						{/* Current balance info */}
						<div className="flex items-center justify-between rounded-xl bg-secondary-50 px-4 py-3">
							<span className="text-[13px] text-secondary-500">
								{t("currentBalance")}
							</span>
							<span className="text-[15px] font-semibold text-secondary-900">
								Rp {asset.balance.toLocaleString("id-ID")}
							</span>
						</div>

						{/* Desired balance */}
						<div>
							<label className={labelClass}>{t("desiredBalance")}</label>
							<input
								type="number"
								value={balance}
								onChange={(e) => setBalance(e.target.value)}
								placeholder="0"
								min="0"
								step="1000"
								className="w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 placeholder:text-secondary-300"
							/>
						</div>

						{/* Note */}
						<div>
							<label className={labelClass}>
								{t("note")}{" "}
								<span className="text-secondary-300">({t("optional")})</span>
							</label>
							<input
								type="text"
								value={note}
								onChange={(e) => setNote(e.target.value)}
								placeholder={t("notePlaceholder")}
								maxLength={500}
								className="w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 placeholder:text-secondary-300"
							/>
						</div>

						{/* Date */}
						<div>
							<label className={labelClass}>
								{t("date")}{" "}
								<span className="text-secondary-300">({t("optional")})</span>
							</label>
							<input
								type="datetime-local"
								value={date}
								onChange={(e) => setDate(e.target.value)}
								className="w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400"
							/>
						</div>

						{/* Info hint */}
						<div className="flex items-start gap-2 rounded-xl bg-info-50 px-3 py-2.5">
							<Info size={14} className="text-info-600 shrink-0 mt-[1px]" />
							<p className="text-[12px] leading-snug text-info-700">
								{t("hints.info")}
							</p>
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
