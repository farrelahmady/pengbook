"use client";
import {
	PeriodBadge,
	Topbar,
	TopbarSummaryCard,
} from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { reportService } from "@/services/report";
import { useQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";

export default function ReportTopbar() {
	const t = useTranslations("reportPage");
	const format = useFormatter();
	const currencyFormat = useCurrencyFormatter();
	const { data, isLoading } = useQuery({
		queryKey: ["reportSummary"],
		queryFn: () => reportService.getSummary(),
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
			<div className="grid grid-cols-2 gap-2 mt-4">
				<TopbarSummaryCard
					label={t("summaryNetIncome.label")}
					value={
						isLoading ? (
							<Skeleton className="w-16 h-3 rounded" />
						) : (
							currencyFormat(data?.netIncome ?? 0, {
								compact: true,
								currency: "Rp",
							})
						)
					}
					sub={t("summaryNetIncome.sub")}
					accent
				/>
				<TopbarSummaryCard
					label={t("summaryTotalAsset.label")}
					value={
						isLoading ? (
							<Skeleton className="w-16 h-3 rounded" />
						) : (
							currencyFormat(data?.totalAsset ?? 0, {
								compact: true,
								currency: "Rp",
							})
						)
					}
					sub={t("summaryTotalAsset.sub")}
				/>
				<TopbarSummaryCard
					label={t("summaryNetCashFlow.label")}
					value={
						isLoading ? (
							<Skeleton className="w-16 h-3 rounded" />
						) : (
							currencyFormat(data?.netCashFlow ?? 0, {
								compact: true,
								currency: "Rp",
							})
						)
					}
					sub={t("summaryNetCashFlow.sub")}
				/>
				<TopbarSummaryCard
					label={t("summaryTrialBalance.label")}
					value={
						isLoading ? (
							<Skeleton className="w-16 h-3 rounded" />
						) : data?.isTrialBalanceBalanced ? (
							"Balance"
						) : (
							"Unbalanced"
						)
					}
					sub={t("summaryTrialBalance.sub")}
				/>
			</div>
		</Topbar>
	);
}
