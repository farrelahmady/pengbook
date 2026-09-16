"use client";

import { accountService } from "@/services/account";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { AccountTypeGroupCard } from "./account-type-group";
import { AccountTypeGroupSkeleton } from "./account-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { BookOpen, Search, SearchX, X } from "lucide-react";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";
import { filterAccountTree } from "./account-search";

const logger = createLogger("account.list");

export default function AccountList() {
	const t = useTranslations("accountPage");
	const [query, setQuery] = useState("");
	const { data, isLoading } = useQuery({
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

			<div className="flex flex-col gap-2 px-3 pb-4">
				{isLoading && (
					<>
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
					</>
				)}

				{!isLoading && (!data || data.groups.length === 0) && !searching && (
					<EmptyState
						icon={BookOpen}
						title={t("emptyTitle")}
						description={t("emptyDescription")}
					/>
				)}

				{!isLoading && data && searching && totalMatches === 0 && (
					<EmptyState
						icon={SearchX}
						title={t("search.emptyTitle")}
						description={t("search.emptyDescription")}
					/>
				)}

				{!isLoading && data && (!searching || totalMatches > 0) && (
					<>
						{(searching ? filteredGroups.filter((g) => g.accounts.length > 0) : filteredGroups).map((group) => (
							<AccountTypeGroupCard
								key={group.type}
								group={group}
								forceExpanded={searching}
								query={query}
							/>
						))}
					</>
				)}
			</div>
		</div>
	);
}
