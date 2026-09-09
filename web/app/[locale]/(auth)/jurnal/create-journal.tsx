"use client";
import { CreateJournalSheet } from "@/components/journal/create-journal-sheet";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useTranslations } from "next-intl";

export default function CreateJournal() {
	const t = useTranslations("journalPage");
	const [sheetOpen, setSheetOpen] = useState(false);

	return (
		<>
			{/* ── FAB ── */}
			<button
				onClick={() => setSheetOpen(true)}
				className="fixed bottom-22 left-1/2 -translate-x-1/2 z-30
                   flex items-center gap-2 px-6 py-3.5 rounded-2xl
                   bg-primary-500 text-white font-semibold text-[14px]
                   shadow-[0_4px_20px_rgba(59,79,212,0.4)]
                   active:scale-95 transition-transform no-tap"
			>
				<Plus size={18} strokeWidth={2.5} />
				{t("fab")}
			</button>

			{/* ── Sheet ── */}
			<CreateJournalSheet open={sheetOpen} onOpenChange={setSheetOpen} />
		</>
	);
}
