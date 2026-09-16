"use client";
import {
	PeriodBadge,
	Topbar,
	TopbarSummaryCard,
} from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { accountService } from "@/services/account";
import { useQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";
import { useEffect } from "react";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";

const logger = createLogger("account.topbar");

export default function AccountTopbar() {
	const t = useTranslations("accountPage");
	const format = useFormatter();
	const { data, isLoading, isError } = useQuery({
		queryKey: queryKeys.accounts.summary,
		queryFn: () => accountService.getSummary(),
	});

	useEffect(() => {
		if (isError) {
			logger.error("Failed to load account summary");
		}
	}, [isError]);

	const labelPeriod = format.dateTime(new Date(), {
		month: "short",
		year: "numeric",
	});

	return (
		<Topbar
			subtitle={t("title")}
			right={<PeriodBadge label={labelPeriod} />}
		>
			<div className="grid grid-cols-3 gap-2 mt-4">
				<TopbarSummaryCard
					label={t("summaryTotal.label")}
					value={
						isLoading ? (
							<Skeleton className="w-10 h-3 rounded" />
						) : isError ? (
							"–"
						) : (
							`${data?.totalAccounts ?? 0}`
						)
					}
					accent
				/>
				<TopbarSummaryCard
					label={t("summaryPosting.label")}
					value={
						isLoading ? (
							<Skeleton className="w-10 h-3 rounded" />
						) : isError ? (
							"–"
						) : (
							`${data?.postingAccounts ?? 0}`
						)
					}
				/>
				<TopbarSummaryCard
					label={t("summaryHeader.label")}
					value={
						isLoading ? (
							<Skeleton className="w-10 h-3 rounded" />
						) : isError ? (
							"–"
						) : (
							`${data?.headerAccounts ?? 0}`
						)
					}
				/>
			</div>
		</Topbar>
	);
}
