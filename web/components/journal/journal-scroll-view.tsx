"use client";
import { JournalEntry } from "@/types";
import { JournalCard, JournalCardSkeleton } from "./journal-card";
import { EditJournalSheet } from "./edit-journal-sheet";
import { EmptyState } from "@/components/shared/empty-state";
import { ReceiptText } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { journalService } from "@/services/journal";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";

interface JournalScrollViewProps {
	startDate?: Date;
	endDate?: Date;
	accountIds?: string[];
}

function groupByDate(journals: JournalEntry[]) {
	const map = new Map<string, JournalEntry[]>();
	for (const j of journals) {
		const key = j.date.slice(0, 10);
		if (!map.has(key)) map.set(key, []);
		map.get(key)!.push(j);
	}
	return map;
}

export function JournalScrollView({
	startDate,
	endDate,
	accountIds,
}: JournalScrollViewProps) {
	const observerRef = useRef<HTMLDivElement>(null);
	const LIMIT = 3;
	const [selectedJournal, setSelectedJournal] = useState<JournalEntry | null>(
		null,
	);
	const [editOpen, setEditOpen] = useState(false);

	const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } =
		useInfiniteQuery({
			queryKey: ["journals", "scroll-view", { startDate, endDate, accountIds }],
			queryFn: ({ pageParam }) => {
				return journalService.getAllScrollView({
					page: pageParam,
					limit: LIMIT,
					startDate,
					endDate,
					accountIds,
				});
			},
			initialPageParam: 1,
			getNextPageParam: (lastPage, allPages) => {
				if (!lastPage || lastPage.length < LIMIT) return undefined;
				return allPages.length + 1;
			},
		});

	const journals = useMemo(() => {
		if (!data) return [];
		return data.pages.flat();
	}, [data]);

	const grouped = useMemo(() => groupByDate(journals), [journals]);

	useEffect(() => {
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
					fetchNextPage();
				}
			},
			{
				threshold: 0,
				rootMargin: "100px",
			},
		);

		if (observerRef.current) {
			observer.observe(observerRef.current);
		}

		return () => observer.disconnect();
	}, [hasNextPage, isFetchingNextPage, fetchNextPage]);

	function handleJournalClick(journal: JournalEntry) {
		setSelectedJournal(journal);
		setEditOpen(true);
	}

	return (
		<>
			<div className="flex flex-col gap-2 px-3 pb-4">
				<JournalScrollViewContent
					isLoading={isLoading}
					journals={journals}
					grouped={grouped}
					isFetchingNextPage={isFetchingNextPage}
					observerRef={observerRef}
					onJournalClick={handleJournalClick}
				/>
			</div>

			<EditJournalSheet
				journal={selectedJournal}
				open={editOpen}
				onOpenChange={setEditOpen}
			/>
		</>
	);
}

function JournalScrollViewContent({
	isLoading,
	journals,
	grouped,
	isFetchingNextPage,
	observerRef,
	onJournalClick,
}: {
	isLoading: boolean;
	journals: JournalEntry[];
	grouped: Map<string, JournalEntry[]>;
	isFetchingNextPage: boolean;
	observerRef: React.RefObject<HTMLDivElement | null>;
	onJournalClick: (journal: JournalEntry) => void;
}) {
	const format = useFormatter();
	const t = useTranslations("journalPage");

	if (isLoading) {
		return (
			<div className="flex flex-col gap-2">
				{Array.from({ length: 3 }).map((_, i) => (
					<JournalCardSkeleton key={i} />
				))}
			</div>
		);
	}

	if (journals.length === 0) {
		return (
			<EmptyState
				icon={ReceiptText}
				title={t("emptyTitle")}
				description={t("emptyDescription")}
			/>
		);
	}

	return (
		<>
			{Array.from(grouped.entries()).map(([date, entries]) => (
				<div key={date}>
					<p className="text-[11px] font-semibold text-secondary-400 tracking-wide px-1 pt-3 pb-1.5">
						{format.dateTime(new Date(date), {
							weekday: "long",
							day: "numeric",
							month: "long",
							year: "numeric",
						})}
					</p>
					<div className="flex flex-col gap-2">
						{entries.map((journal) => (
							<JournalCard
								key={journal.id}
								journal={journal}
								onClick={() => onJournalClick(journal)}
							/>
						))}
					</div>
				</div>
			))}
			{isFetchingNextPage && (
				<p className="text-center text-[12px] text-secondary-400 py-2">
					{t("loadingMore")}
				</p>
			)}
			<div ref={observerRef} className="h-1" />
		</>
	);
}
