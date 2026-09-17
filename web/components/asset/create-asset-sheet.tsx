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
import { AppField } from "@/components/ui/app-field";
import { AppInput } from "@/components/ui/app-input";
import { AppSubmitButton } from "@/components/ui/app-submit-button";
import { AccountSelect } from "@/components/akun/account-select";

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
						<AccountSelect
							label={t("parent")}
							value={parentId}
							onChange={setParentId}
							accounts={assetParents}
							isLoading={isLoadingParents}
							placeholder={t("parentPlaceholder")}
							loadingText="Loading..."
							grouped={false}
						/>

						{/* Name */}
						<AppField label={t("name")} hint={t("hints.autoCode")}>
							<AppInput
								type="text"
								value={name}
								maxLength={200}
								onChange={(e) => setName(e.target.value)}
								placeholder={t("namePlaceholder")}
							/>
						</AppField>

						{/* Contextual hint */}
						<div className="flex items-start gap-2 rounded-xl bg-info-50 px-3 py-2.5">
							<Info size={14} className="text-info-600 shrink-0 mt-[1px]" />
							<p className="text-[12px] leading-snug text-info-700">
								{t("hints.assetPosting")}
							</p>
						</div>

						<AppSubmitButton
							type="button"
							onClick={handleSubmit}
							disabled={isSubmitting}
						>
							{t("submit")}
						</AppSubmitButton>
					</div>
				</div>
			</SheetContent>
		</Sheet>
	);
}
