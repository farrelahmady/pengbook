"use client";

import { accountService } from "@/services/account";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { AccountTypeGroupCard } from "./account-type-group";
import { AccountTypeGroupSkeleton } from "./account-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { BookOpen } from "lucide-react";
import { queryKeys } from "@/lib/query-keys";

export default function AccountList() {
	const t = useTranslations("coaPage");
	const { data, isLoading } = useQuery({
		queryKey: queryKeys.accounts.summary,
		queryFn: () => accountService.getSummary(),
	});

	return (
		<div className="px-1">
			<p className="section-label">{t("sectionTitle")}</p>

			<div className="flex flex-col gap-2 px-3 pb-4">
				{isLoading && (
					<>
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
						<AccountTypeGroupSkeleton />
					</>
				)}

				{!isLoading && (!data || data.groups.length === 0) && (
					<EmptyState
						icon={BookOpen}
						title={t("emptyTitle")}
						description={t("emptyDescription")}
					/>
				)}

				{!isLoading && data && (
					<>
						{data.groups.map((group) => (
							<AccountTypeGroupCard key={group.type} group={group} />
						))}
					</>
				)}
			</div>
		</div>
	);
}
