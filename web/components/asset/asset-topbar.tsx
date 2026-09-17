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
import { Download } from "lucide-react";
import { toast } from "sonner";
import { useEffect } from "react";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";

const logger = createLogger("asset.topbar");

export default function AssetTopbar() {
	const t = useTranslations("assetPage");
	const format = useFormatter();
	const currencyFormat = useCurrencyFormatter();
	const { data, isLoading, isError } = useQuery({
		queryKey: queryKeys.assets.summary,
		queryFn: () => assetService.getSummary(),
	});

	useEffect(() => {
		if (isError) {
			logger.error("Failed to load asset summary");
		}
	}, [isError]);

	const labelPeriod = format.dateTime(new Date(), {
		month: "short",
		year: "numeric",
	});

	async function handleDownload() {
		logger.debug("Asset export download initiated");
		const toastId = toast.loading(t("export.toastLoading"));
		try {
			const blob = await assetService.exportExcel();

			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = `aset-${new Date().toISOString().slice(0, 10)}.xlsx`;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);

			logger.info("Asset export download completed");
			toast.success(t("export.toastSuccess"), { id: toastId });
		} catch (error) {
			logger.error("Asset export download failed", { error });
			toast.error(t("export.toastError"), { id: toastId });
		}
	}

	return (
		<>
			{/* ── Topbar ── */}
			<Topbar
				subtitle={t("title")}
				right={
					<div className="flex items-center gap-2">
						<button
							onClick={handleDownload}
							className="w-8 h-8 rounded-full flex items-center justify-center
                           bg-white/10 hover:bg-white/20 transition-colors"
							aria-label={t("export.button")}
						>
							<Download size={14} className="text-white/60" />
						</button>
						<PeriodBadge label={labelPeriod} />
					</div>
				}
			>
				<div className="flex gap-2 mt-4">
					<TopbarSummaryCard
						label={t("summaryTotalAsset.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : isError ? (
								"–"
							) : (
								currencyFormat(data?.totalAsset ?? 0, {
									compact: true,
									currency: "Rp",
								})
							)
						}
						sub={
							isLoading || isError
								? undefined
								: `${data?.accountCount ?? 0} ${t("summaryTotalAsset.sub")}`
						}
					/>
					<TopbarSummaryCard
						label={t("summaryCurrentAsset.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : isError ? (
								"–"
							) : (
								currencyFormat(data?.currentAsset ?? 0, {
									compact: true,
									currency: "Rp",
								})
							)
						}
						sub={isError ? undefined : t("summaryCurrentAsset.sub")}
						accent
					/>
				</div>
			</Topbar>
		</>
	);
}
