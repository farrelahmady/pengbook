"use client";

import { useState } from "react";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { EditBasicForm } from "./edit-basic-form";
import { EditAdvancedForm } from "./edit-advanced-form";
import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { JournalEntryListItem } from "@/types";
import { journalService } from "@/services/journal";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";

const logger = createLogger("EditJournalSheet");

interface EditJournalSheetProps {
	journal: JournalEntryListItem | null;
	open: boolean;
	onOpenChange: (v: boolean) => void;
}

type Mode = "basic" | "advanced";

export function EditJournalSheet({
	journal,
	open,
	onOpenChange,
}: EditJournalSheetProps) {
	const t = useTranslations("journalPage.sheet");
	const queryClient = useQueryClient();
	const hasMoreThan2Lines = journal ? journal.lines.length > 2 : false;

	const [mode, setMode] = useState<Mode>(hasMoreThan2Lines ? "advanced" : "basic");
	const [confirming, setConfirming] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);

	if (!journal) return null;
	const entry = journal;

	const showModeToggle = entry.lines.length <= 2;
	const useAdvanced = hasMoreThan2Lines || mode === "advanced";

	function handleOpenChange(v: boolean) {
		if (!v) setConfirming(false);
		onOpenChange(v);
	}

	async function handleDelete() {
		const toastId = toast.loading(t("toastDeleting"));
		setIsDeleting(true);

		try {
			await journalService.delete(entry.id);
			toast.success(t("toastDeleteSuccess"), { id: toastId });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
			queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
			onOpenChange(false);
		} catch (err) {
			logger.error("Failed to delete journal", { id: entry.id, error: err });
			toast.error(t("toastDeleteError"), { id: toastId });
		} finally {
			setIsDeleting(false);
			setConfirming(false);
		}
	}

	return (
		<Sheet open={open} onOpenChange={handleOpenChange}>
			<SheetContent
				side="bottom"
				className="p-0 rounded-t-3xl max-w-[390px] mx-auto max-h-[92dvh] overflow-y-auto"
			>
				{/* Handle */}
				<div className="sheet-handle" />

				{/* Header */}
				<div className="px-5 pt-3 pb-0">
					<div className="flex items-center justify-between mb-3">
						<h2 className="text-[17px] font-bold text-secondary-900">
							{t("editTitle")}
						</h2>
					</div>

					{/* Mode toggle - only show if <= 2 lines */}
					{showModeToggle && (
						<div className="flex bg-secondary-100 rounded-xl p-1 gap-1 mb-4">
							{(["basic", "advanced"] as Mode[]).map((m) => (
								<button
									key={m}
									onClick={() => setMode(m)}
									className={cn(
										"flex-1 py-2 rounded-lg text-[13px] font-semibold transition-all no-tap capitalize",
										mode === m
											? "bg-white text-secondary-900 shadow-sm"
											: "text-secondary-400 hover:text-secondary-600",
									)}
								>
									{t(`modes.${m}`)}
								</button>
							))}
						</div>
					)}

					<div className="h-px bg-black/[0.06] -mx-5 mb-4" />
				</div>

				{/* Form */}
				<div className="px-5 pb-8">
					{useAdvanced ? (
						<EditAdvancedForm
							journal={entry}
							onSuccess={() => onOpenChange(false)}
						/>
					) : (
						<EditBasicForm
							journal={entry}
							onSuccess={() => onOpenChange(false)}
						/>
					)}

					{/* Delete zone */}
					<div className="mt-5 border-t border-black/[0.06] pt-4">
						{!confirming ? (
							<button
								onClick={() => setConfirming(true)}
								className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-[14px] font-semibold text-danger-500 border border-danger-200 bg-danger-50 active:scale-[0.98] transition-all no-tap"
							>
								<Trash2 size={15} />
								{t("delete")}
							</button>
						) : (
							<div className="bg-danger-50 border border-danger-200 rounded-xl p-3.5">
								<p className="text-[13px] font-semibold text-secondary-900">
									{t("deleteConfirmTitle")}
								</p>
								<p className="text-[11px] text-secondary-500 mt-0.5">
									{t("deleteConfirmDesc")}
								</p>
								<div className="flex gap-2 mt-3">
									<button
										onClick={() => setConfirming(false)}
										disabled={isDeleting}
										className="flex-1 py-2.5 rounded-lg text-[13px] font-semibold bg-white text-secondary-600 border border-secondary-200 disabled:opacity-50 no-tap"
									>
										{t("deleteCancel")}
									</button>
									<button
										onClick={handleDelete}
										disabled={isDeleting}
										className="flex-1 py-2.5 rounded-lg text-[13px] font-semibold bg-danger-500 text-white disabled:opacity-50 active:scale-[0.98] transition-all no-tap"
									>
										{t("deleteConfirm")}
									</button>
								</div>
							</div>
						)}
					</div>
				</div>
			</SheetContent>
		</Sheet>
	);
}
