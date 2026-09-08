"use client";
import {
	PeriodBadge,
	Topbar,
	TopbarSummaryCard,
} from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { assetService } from "@/services/asset";
import { useQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";

export default function AssetTopbar() {
	const t = useTranslations("assetPage");
	const format = useFormatter();
	const currencyFormat = useCurrencyFormatter();
	const { data, isLoading } = useQuery({
		queryKey: ["assetSummary"],
		queryFn: () => assetService.getSummary(),
	});

	const totalAccount = data?.groups.reduce(
		(sum, g) => sum + g.accounts.length,
		0,
	) ?? 0;

	const labelPeriod = format.dateTime(new Date(), {
		month: "short",
		year: "numeric",
	});

	return (
		<>
			{/* ── Topbar ── */}
			<Topbar
				subtitle={t("title")}
				right={<PeriodBadge label={labelPeriod} />}
			>
				<div className="flex gap-2 mt-4">
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
						sub={
							isLoading ? undefined : `${totalAccount} ${t("summaryTotalAsset.sub")}`
						}
					/>
					<TopbarSummaryCard
						label={t("summaryCurrentAsset.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : (
								currencyFormat(data?.currentAsset ?? 0, {
									compact: true,
									currency: "Rp",
								})
							)
						}
						sub={t("summaryCurrentAsset.sub")}
						accent
					/>
				</div>
			</Topbar>
		</>
	);
}
