"use client";

import { useState } from "react";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import { EditBasicForm } from "./edit-basic-form";
import { EditAdvancedForm } from "./edit-advanced-form";
import { cn } from "@/lib/utils";
import { useTranslations } from "next-intl";
import { JournalEntry } from "@/types";

interface EditJournalSheetProps {
	journal: JournalEntry | null;
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
	const hasMoreThan2Lines = journal ? journal.lines.length > 2 : false;

	const [mode, setMode] = useState<Mode>(hasMoreThan2Lines ? "advanced" : "basic");

	if (!journal) return null;

	const showModeToggle = journal.lines.length <= 2;
	const useAdvanced = hasMoreThan2Lines || mode === "advanced";

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
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
							journal={journal}
							onSuccess={() => onOpenChange(false)}
						/>
					) : (
						<EditBasicForm
							journal={journal}
							onSuccess={() => onOpenChange(false)}
						/>
					)}
				</div>
			</SheetContent>
		</Sheet>
	);
}
