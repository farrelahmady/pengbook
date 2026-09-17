"use client";

import { useState } from "react";
import { Info } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { accountService } from "@/services/account";
import { assetService } from "@/services/asset";
import { createLogger } from "@/lib/logger";
import { queryKeys } from "@/lib/query-keys";
import type { ParentListItem } from "@/types";

const logger = createLogger("asset.create-sheet");

interface CreateAssetSheetProps {
	open: boolean;
	onOpenChange: (v: boolean) => void;
}

export function CreateAssetSheet({
	open,
	onOpenChange,
}: CreateAssetSheetProps) {
	const t = useTranslations("assetPage.create");
	const queryClient = useQueryClient();
	const [parentId, setParentId] = useState("");
	const [name, setName] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);
	const type = "ASSET";

	// Fetch level-2 ASSET parents only
	const { data: parents = [], isLoading: isLoadingParents } = useQuery({
		queryKey: queryKeys.accounts.parentsWithType(2, type), // level 2 for getting level-2 children
		queryFn: () => accountService.getParents(2, type),
		enabled: open,
	});

	// Filter to only ASSET type level-2 parents
	const assetParents = parents.filter(
		(p: ParentListItem) => p.type === "ASSET" && p.level === 2,
	);

	async function handleSubmit() {
		if (!parentId || !name.trim()) {
			toast.error(t("toastValidation"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);
		logger.debug("Creating asset", { parentId, name });

		try {
			const created = await assetService.create(Number(parentId), name.trim());

			toast.success(t("toastSuccess", { code: created.code }), { id: toastId });
			logger.info("Asset created", { id: created.id, code: created.code });

			// Invalidate asset queries and account queries (CoA changed)
			queryClient.invalidateQueries({ queryKey: queryKeys.assets.all });
			queryClient.invalidateQueries({ queryKey: queryKeys.accounts.all });

			setParentId("");
			setName("");
			onOpenChange(false);
		} catch (err) {
			logger.error("Failed to create asset", { error: err });
			toast.error(t("toastError"), { id: toastId });
		} finally {
			setIsSubmitting(false);
		}
	}

	const selectClass =
		"w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 appearance-none cursor-pointer disabled:opacity-50";
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
						{/* Parent */}
						<div>
							<label className={labelClass}>{t("parent")}</label>
							<select
								value={parentId}
								onChange={(e) => setParentId(e.target.value)}
								disabled={isLoadingParents}
								className={selectClass}
							>
								<option value="">{t("parentPlaceholder")}</option>
								{assetParents.map((p: ParentListItem) => (
									<option key={p.id} value={String(p.id)}>
										{p.code} · {p.name}
									</option>
								))}
							</select>
						</div>

						{/* Name */}
						<div>
							<label className={labelClass}>{t("name")}</label>
							<input
								type="text"
								value={name}
								maxLength={200}
								onChange={(e) => setName(e.target.value)}
								placeholder={t("namePlaceholder")}
								className="w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 placeholder:text-secondary-300"
							/>
							<p className="text-[11px] text-secondary-400 mt-1.5">
								{t("hints.autoCode")}
							</p>
						</div>

						{/* Contextual hint */}
						<div className="flex items-start gap-2 rounded-xl bg-info-50 px-3 py-2.5">
							<Info size={14} className="text-info-600 shrink-0 mt-[1px]" />
							<p className="text-[12px] leading-snug text-info-700">
								{t("hints.assetPosting")}
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
