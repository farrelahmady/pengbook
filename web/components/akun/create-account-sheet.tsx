"use client";

import { useState } from "react";
import { Info } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { accountService } from "@/services/account";
import { createLogger } from "@/lib/logger";
import { queryKeys } from "@/lib/query-keys";
import { cn } from "@/lib/utils";

const logger = createLogger("account.create-sheet");

interface CreateAccountSheetProps {
	open: boolean;
	onOpenChange: (v: boolean) => void;
}

const LEVELS = [1, 2, 3] as const;

export function CreateAccountSheet({
	open,
	onOpenChange,
}: CreateAccountSheetProps) {
	const t = useTranslations("accountPage.create");
	const queryClient = useQueryClient();
	const [level, setLevel] = useState<number>(3);
	const [parentId, setParentId] = useState("");
	const [name, setName] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	const { data: parents = [], isLoading: isLoadingParents } = useQuery({
		queryKey: queryKeys.accounts.parents(level - 1),
		queryFn: () => accountService.getParents(level - 1),
		enabled: open,
	});

	const groupedParents = (() => {
		const map = new Map<string, typeof parents>();
		for (const p of parents) {
			const list = map.get(p.type) ?? [];
			list.push(p);
			map.set(p.type, list);
		}
		return [...map.entries()];
	})();

	function handleLevelChange(v: number) {
		setLevel(v);
		setParentId("");
		logger.debug("Create level changed", { level: v });
	}

	const selectedParent = parents.find((p) => String(p.id) === parentId);
	const showAssetHint = level === 3 && selectedParent?.type === "ASSET";

	async function handleSubmit() {
		if (!parentId || !name.trim()) {
			toast.error(t("toastValidation"));
			return;
		}

		const toastId = toast.loading(t("toastLoading"));
		setIsSubmitting(true);
		logger.debug("Creating account", { level, parentId, name });

		try {
			const created = await accountService.create({
				name: name.trim(),
				parentId: Number(parentId),
			});

			toast.success(t("toastSuccess", { code: created.code }), { id: toastId });
			logger.info("Account created", { id: created.id, code: created.code });
			queryClient.invalidateQueries({ queryKey: queryKeys.accounts.all });
			setParentId("");
			setName("");
			onOpenChange(false);
		} catch (err) {
			logger.error("Failed to create account", { error: err });
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
						{/* Level */}
						<div>
							<label className={labelClass}>{t("level")}</label>
							<div className="flex bg-secondary-100 rounded-xl p-1 gap-1">
								{LEVELS.map((l) => (
									<button
										key={l}
										type="button"
										onClick={() => handleLevelChange(l)}
										className={cn(
											"flex-1 py-2 rounded-lg text-[13px] font-semibold transition-all no-tap",
											level === l
												? "bg-white text-secondary-900 shadow-sm"
												: "text-secondary-400 hover:text-secondary-600",
										)}
									>
										{t(`level${l}`)}
									</button>
								))}
							</div>
						</div>

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
								{groupedParents.map(([type, list]) => (
									<optgroup key={type} label={type}>
										{list.map((p) => (
											<option key={p.id} value={String(p.id)}>
												{p.code} · {p.name}
											</option>
										))}
									</optgroup>
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
								{level < 3 ? t("hints.header") : t("hints.posting")}
								{showAssetHint && ` ${t("hints.assetBalance")}`}
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
