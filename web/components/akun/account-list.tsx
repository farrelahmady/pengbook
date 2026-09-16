"use client";

import { accountService } from "@/services/account";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { AccountTypeGroupCard } from "./account-type-group";
import { AccountTypeGroupSkeleton } from "./account-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { BookOpen, Search, SearchX, TriangleAlert, X } from "lucide-react";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";
import { filterAccountTree } from "./account-search";

const logger = createLogger("account.list");

export default function AccountList() {
	const t = useTranslations("accountPage");
	const [query, setQuery] = useState("");
	const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());
	const { data, isLoading, isError, refetch } = useQuery({
		queryKey: queryKeys.accounts.tree,
		queryFn: () => accountService.getTree(),
	});

	const searching = query.trim() !== "";
	const filteredGroups =
		data?.groups.map((group) => {
			const { filtered, matchCount } = filterAccountTree(
				group.accounts,
				query,
			);
			return { ...group, accounts: filtered, matchCount };
		}) ?? [];
	const totalMatches = filteredGroups.reduce(
		(sum, g) => sum + g.matchCount,
		0,
	);

	function toggleGroup(type: string) {
		setExpandedGroups((prev) => {
			const next = new Set(prev);
			if (next.has(type)) {
				next.delete(type);
			} else {
				next.add(type);
			}
			logger.debug("Account group toggled", { type, expanded: next.has(type) });
			return next;
		});
	}

	const visibleGroups = searching
		? filteredGroups.filter((g) => g.accounts.length > 0)
		: filteredGroups;

	function expandAll() {
		setExpandedGroups(new Set(visibleGroups.map((g) => g.type)));
		logger.debug("Account groups expanded all");
	}

	function collapseAll() {
		setExpandedGroups(new Set());
		logger.debug("Account groups collapsed all");
	}

	return (
		<div className="px-1">
			<p className="section-label">{t("sectionTitle")}</p>

			<div className="px-3 pb-2">
				<div className="relative">
					<Search
						size={16}
						className="absolute left-3 top-1/2 -translate-y-1/2 text-secondary-400 pointer-events-none"
					/>
					<input
						type="text"
						value={query}
						onChange={(e) => {
							setQuery(e.target.value);
							logger.debug("Account search", { query: e.target.value });
						}}
						placeholder={t("search.placeholder")}
						aria-label={t("search.placeholder")}
						className="w-full bg-secondary-50 border border-black/[0.06] rounded-xl pl-9 pr-9 py-2.5 text-[13px] text-secondary-800 placeholder:text-secondary-400 outline-none focus:border-primary-300 transition-colors"
					/>
					{query !== "" && (
						<button
							type="button"
							onClick={() => setQuery("")}
							aria-label={t("search.clear")}
							className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-lg text-secondary-400 hover:text-secondary-600 hover:bg-secondary-100 active:scale-95 transition-all no-tap"
						>
							<X size={14} />
						</button>
					)}
				</div>
				{searching && !isLoading && data && (
					<p className="text-[11px] text-secondary-400 mt-1.5 px-1">
						{t("search.resultCount", { count: totalMatches, query: query.trim() })}
					</p>
				)}
			</div>

			{!searching && !isLoading && !isError && data && data.groups.length > 0 && (
				<div className="flex justify-end px-4 pb-1">
					<button
						type="button"
						onClick={expandAll}
						className="text-[12px] font-semibold text-primary-500 no-tap"
					>
						{t("expandAll")}
					</button>
					<span className="text-[12px] text-secondary-300 mx-1.5">·</span>
					<button
						type="button"
						onClick={collapseAll}
						className="text-[12px] font-semibold text-primary-500 no-tap"
					>
						{t("collapseAll")}
					</button>
				</div>
			)}

			<div className="flex flex-col gap-2 px-3 pb-4">
				{isLoading && (
					<>
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
					</>
				)}

				{!isLoading && isError && (
					<div className="card-default shadow-card flex flex-col items-center justify-center py-12 px-8 text-center">
						<div className="w-14 h-14 rounded-2xl bg-danger-50 flex items-center justify-center mb-4">
							<TriangleAlert size={24} className="text-danger-600" strokeWidth={1.5} />
						</div>
						<p className="text-[14px] font-semibold text-secondary-700 mb-1">
							{t("error.title")}
						</p>
						<p className="text-[12px] text-secondary-400 leading-relaxed">
							{t("error.description")}
						</p>
						<button
							type="button"
							onClick={() => refetch()}
							className="mt-4 px-4 py-2 rounded-xl bg-primary-500 text-white text-[13px] font-semibold active:scale-95 transition-all no-tap"
						>
							{t("error.retry")}
						</button>
					</div>
				)}

				{!isLoading && !isError && (!data || data.groups.length === 0) && !searching && (
					<EmptyState
						icon={BookOpen}
						title={t("emptyTitle")}
						description={t("emptyDescription")}
					/>
				)}

				{!isLoading && !isError && data && searching && totalMatches === 0 && (
					<EmptyState
						icon={SearchX}
						title={t("search.emptyTitle")}
						description={t("search.emptyDescription")}
					/>
				)}

				{!isLoading && !isError && data && (!searching || totalMatches > 0) && (
					<>
						{visibleGroups.map((group) => (
							<AccountTypeGroupCard
								key={group.type}
								group={group}
								expanded={searching ? true : expandedGroups.has(group.type)}
								onToggle={() => toggleGroup(group.type)}
								query={query}
							/>
						))}
					</>
				)}
			</div>
		</div>
	);
}
