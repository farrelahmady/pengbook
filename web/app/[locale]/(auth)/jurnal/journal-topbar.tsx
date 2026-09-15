"use client";
import {
	PeriodBadge,
	Topbar,
	TopbarSummaryCard,
} from "@/components/layout/topbar";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { journalService } from "@/services/journal";
import { useQuery } from "@tanstack/react-query";
import { useFormatter, useTranslations } from "next-intl";
import { Download } from "lucide-react";
import { toast } from "sonner";
import { createLogger } from "@/lib/logger";
import { queryKeys } from "@/lib/query-keys";

const logger = createLogger("JournalTopbar");

interface JournalTopbarProps {
	startDate?: Date;
	endDate?: Date;
	accountIds?: string[];
}

export default function JournalTopbar({
	startDate,
	endDate,
	accountIds,
}: JournalTopbarProps) {
	const t = useTranslations("journalPage");
	const format = useFormatter();
	const currencyFormat = useCurrencyFormatter();
	const { data, isLoading } = useQuery({
		queryKey: queryKeys.journals.summary,
		queryFn: () => {
			logger.debug("Fetching journal summary");
			return journalService.getTotalSummary();
		},
	});

	const labelPeriod = format.dateTime(new Date(), {
		month: "short",
		year: "numeric",
	});

	async function handleDownload() {
		logger.info("Download initiated", {
			hasStartDate: !!startDate,
			hasEndDate: !!endDate,
			accountCount: accountIds?.length ?? 0,
		});
		const toastId = toast.loading(t("download.toastLoading"));
		try {
			const blob = await journalService.downloadTransactions({
				startDate,
				endDate,
				accountIds,
			});

			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = `jurnal-${new Date().toISOString().slice(0, 10)}.xlsx`;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);

			logger.info("Download completed successfully");
			toast.success(t("download.toastSuccess"), { id: toastId });
		} catch (error) {
			logger.error("Download failed", { error });
			toast.error(t("download.toastError"), { id: toastId });
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
							aria-label={t("download.button")}
						>
							<Download size={14} className="text-white/60" />
						</button>
						<PeriodBadge label={labelPeriod} />
					</div>
				}
			>
				<div className="flex gap-2 mt-4">
					<TopbarSummaryCard
						label={t("summaryDebit.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : (
								currencyFormat(data?.totalDebit ?? 0, {
									compact: true,
									currency: "Rp",
								})
							)
						}
					/>
					<TopbarSummaryCard
						label={t("summaryCredit.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : (
								currencyFormat(data?.totalCredit ?? 0, {
									compact: true,
									currency: "Rp",
								})
							)
						}
					/>
					<TopbarSummaryCard
						label={t("summarytransaction.label")}
						value={
							isLoading ? (
								<Skeleton className="w-16 h-3 rounded" />
							) : (
								`${data?.transactionCount ?? 0}`
							)
						}
						accent
					/>
				</div>
			</Topbar>
		</>
	);
}
