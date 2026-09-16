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
import { Download } from "lucide-react";
import { toast } from "sonner";
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

	async function handleDownload() {
		logger.debug("Account export download initiated");
		const toastId = toast.loading(t("export.toastLoading"));
		try {
			const blob = await accountService.exportExcel();

			const url = URL.createObjectURL(blob);
			const a = document.createElement("a");
			a.href = url;
			a.download = `akun-${new Date().toISOString().slice(0, 10)}.xlsx`;
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);

			logger.info("Account export download completed");
			toast.success(t("export.toastSuccess"), { id: toastId });
		} catch (error) {
			logger.error("Account export download failed", { error });
			toast.error(t("export.toastError"), { id: toastId });
		}
	}

	return (
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
