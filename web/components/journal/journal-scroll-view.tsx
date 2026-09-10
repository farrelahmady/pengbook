"use client";
import { JournalEntryListItem } from "@/types";
import { JournalCard, JournalCardSkeleton } from "./journal-card";
import { EditJournalSheet } from "./edit-journal-sheet";
import { EmptyState } from "@/components/shared/empty-state";
import { ReceiptText } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { journalService } from "@/services/journal";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";
import { useAuth } from "@/lib/auth-context";
import { toast } from "sonner";

interface JournalScrollViewProps {
	startDate?: Date;
	endDate?: Date;
	accountIds?: string[];
}

function groupByDate(journals: JournalEntryListItem[]) {
	const map = new Map<string, JournalEntryListItem[]>();
	for (const j of journals) {
		const key = j.datetime.slice(0, 10);
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
	const { getToken } = useAuth();
	const LIMIT = 10;
	const [selectedJournal, setSelectedJournal] =
		useState<JournalEntryListItem | null>(null);
	const [editOpen, setEditOpen] = useState(false);

	const {
		data,
		fetchNextPage,
		hasNextPage,
		isFetchingNextPage,
		isLoading,
		isError,
		error,
	} = useInfiniteQuery({
		queryKey: ["journals", "scroll-view", { startDate, endDate, accountIds }],
		queryFn: ({ pageParam }) => {
			return journalService.getAllScrollView(
				{
					limit: LIMIT,
					cursor: pageParam,
					startDate,
					endDate,
					accountIds,
				},
				getToken,
			);
		},
		initialPageParam: undefined as string | undefined,
		getNextPageParam: (lastPage) => {
			return lastPage.nextCursor ?? undefined;
		},
		retry: false,
	});

	const journals = useMemo(() => {
		if (!data) return [];
		return data.pages.flatMap((page) => page.data);
	}, [data]);

	const grouped = useMemo(() => groupByDate(journals), [journals]);

	// Use ref for fetchNextPage to avoid re-creating observer on every render
	const fetchNextPageRef = useRef(fetchNextPage);
	fetchNextPageRef.current = fetchNextPage;

	useEffect(() => {
		const observer = new IntersectionObserver(
			(entries) => {
				if (
					entries[0].isIntersecting &&
					hasNextPage &&
					!isFetchingNextPage &&
					!isError
				) {
					fetchNextPageRef.current();
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
	}, [hasNextPage, isFetchingNextPage, isError]);

	// Handle error toast
	useEffect(() => {
		if (isError) {
			toast.error(
				`Failed to load journals. ${error instanceof Error ? error.message : "Unknown error"}`,
			);
		}
	}, [isError, error]);

	function handleJournalClick(journal: JournalEntryListItem) {
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
	journals: JournalEntryListItem[];
	grouped: Map<string, JournalEntryListItem[]>;
	isFetchingNextPage: boolean;
	observerRef: React.RefObject<HTMLDivElement | null>;
	onJournalClick: (journal: JournalEntryListItem) => void;
}) {
	const format = useFormatter();
	const t = useTranslations("journalPage");

	if (isLoading) {
		return <JournalScrollViewSkeletons />;
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
			{isFetchingNextPage && <JournalScrollViewSkeletons />}
			<div ref={observerRef} className="h-1" />
		</>
	);
}

function JournalScrollViewSkeletons() {
	return (
		<div className="flex flex-col gap-2">
			{Array.from({ length: 3 }).map((_, i) => (
				<JournalCardSkeleton key={i} />
			))}
		</div>
	);
}
