"use client";
import {
	PeriodBadge,
	Topbar,
	TopbarSummaryCard,
} from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { coaService } from "@/services/coa";
import { useQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";
import { queryKeys } from "@/lib/query-keys";

export default function CoaTopbar() {
	const t = useTranslations("coaPage");
	const format = useFormatter();
	const { data, isLoading } = useQuery({
		queryKey: queryKeys.coa.summary,
		queryFn: () => coaService.getSummary(),
	});

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
						) : (
							`${data?.headerAccounts ?? 0}`
						)
					}
				/>
			</div>
		</Topbar>
	);
}
