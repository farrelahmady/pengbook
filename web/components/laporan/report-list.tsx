"use client";

import { reportService } from "@/services/report";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useCurrencyFormatter } from "@/hooks/use-currency-formatter";
import { TrialBalanceCard } from "./trial-balance-card";
import { ReportCard } from "./report-card";
import { ReportCardSkeleton } from "./report-card-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { ChartColumn, Square, Clock, FileText } from "lucide-react";

export default function ReportList() {
	const t = useTranslations("reportPage");
	const currencyFormat = useCurrencyFormatter();

	const { data: summary, isLoading: isLoadingSummary } = useQuery({
		queryKey: ["reportSummary"],
		queryFn: () => reportService.getSummary(),
	});

	const { data: trialBalance, isLoading: isLoadingTB } = useQuery({
		queryKey: ["trialBalance"],
		queryFn: () => reportService.getTrialBalance(),
	});

	const { data: incomeStatement, isLoading: isLoadingIS } = useQuery({
		queryKey: ["incomeStatement"],
		queryFn: () => reportService.getIncomeStatement(),
	});

	const { data: balanceSheet, isLoading: isLoadingBS } = useQuery({
		queryKey: ["balanceSheet"],
		queryFn: () => reportService.getBalanceSheet(),
	});

	const { data: cashFlow, isLoading: isLoadingCF } = useQuery({
		queryKey: ["cashFlow"],
		queryFn: () => reportService.getCashFlow(),
	});

	const isLoading = isLoadingSummary || isLoadingTB || isLoadingIS || isLoadingBS || isLoadingCF;

	return (
		<div className="px-1">
			<p className="section-label">{t("sectionTitle")}</p>

			<div className="flex flex-col gap-2 px-3 pb-4">
				{isLoading && (
					<>
						<ReportCardSkeleton />
						<ReportCardSkeleton />
						<ReportCardSkeleton />
						<ReportCardSkeleton />
					</>
				)}

				{!isLoading && !trialBalance && (
					<EmptyState
						icon={FileText}
						title={t("emptyTitle")}
						description={t("emptyDescription")}
					/>
				)}

				{!isLoading && trialBalance && (
					<>
						{/* Trial Balance */}
						<TrialBalanceCard entries={trialBalance} />

						{/* Laba Rugi */}
						{incomeStatement && (
							<ReportCard
								icon={<ChartColumn size={20} className="text-success-600" />}
								iconBg="bg-success-50"
								name={t("reports.incomeStatement.name")}
								subtitle={t("reports.incomeStatement.subtitle")}
								value={`+${currencyFormat(incomeStatement.netIncome, { compact: true, currency: "Rp" })}`}
								valueColor="text-success-600"
							>
								<div className="flex flex-col gap-2">
									{incomeStatement.revenues.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-success-600">
												+{currencyFormat(item.amount)}
											</p>
										</div>
									))}
									{incomeStatement.expenses.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-danger-600">
												-{currencyFormat(item.amount)}
											</p>
										</div>
									))}
									<div className="border-t border-black/[0.06] pt-2 mt-1 flex justify-between">
										<p className="text-[12px] font-semibold text-secondary-900">
											{t("reports.incomeStatement.total")}
										</p>
										<p className="font-mono text-[12px] font-semibold text-success-600">
											+{currencyFormat(incomeStatement.netIncome)}
										</p>
									</div>
								</div>
							</ReportCard>
						)}

						{/* Neraca */}
						{balanceSheet && (
							<ReportCard
								icon={<Square size={20} className="text-primary-600" />}
								iconBg="bg-primary-50"
								name={t("reports.balanceSheet.name")}
								subtitle={t("reports.balanceSheet.subtitle")}
								value="Balance"
								valueColor="text-primary-600"
							>
								<div className="flex flex-col gap-2">
									<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400">
										{t("reports.balanceSheet.assets")}
									</p>
									{balanceSheet.currentAssets.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-success-600">
												{currencyFormat(item.balance)}
											</p>
										</div>
									))}
									{balanceSheet.nonCurrentAssets.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-success-600">
												{currencyFormat(item.balance)}
											</p>
										</div>
									))}
									<div className="border-t border-black/[0.06] pt-2 flex justify-between">
										<p className="text-[11px] font-semibold text-secondary-500">
											{t("reports.balanceSheet.totalAssets")}
										</p>
										<p className="font-mono text-[11px] font-semibold text-secondary-700">
											{currencyFormat(balanceSheet.totalAssets)}
										</p>
									</div>

									<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400 mt-2">
										{t("reports.balanceSheet.liabilities")}
									</p>
									{balanceSheet.liabilities.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-danger-600">
												{currencyFormat(item.balance)}
											</p>
										</div>
									))}
									{balanceSheet.equity.map((item) => (
										<div key={item.name} className="flex justify-between">
											<p className="text-[12px] text-secondary-600">{item.name}</p>
											<p className="font-mono text-[12px] text-danger-600">
												{currencyFormat(item.balance)}
											</p>
										</div>
									))}
									<div className="border-t border-black/[0.06] pt-2 flex justify-between">
										<p className="text-[11px] font-semibold text-secondary-500">
											{t("reports.balanceSheet.totalLiabEquity")}
										</p>
										<p className="font-mono text-[11px] font-semibold text-secondary-700">
											{currencyFormat(
												balanceSheet.totalLiabilities + balanceSheet.totalEquity,
											)}
										</p>
									</div>
								</div>
							</ReportCard>
						)}

						{/* Arus Kas */}
						{cashFlow && (
							<ReportCard
								icon={<Clock size={20} className="text-warning-600" />}
								iconBg="bg-warning-50"
								name={t("reports.cashFlow.name")}
								subtitle={t("reports.cashFlow.subtitle")}
								value={`+${currencyFormat(cashFlow.netCashFlow, { compact: true, currency: "Rp" })}`}
								valueColor="text-success-600"
							>
								<div className="flex flex-col gap-2">
									{(["operating", "investing", "financing"] as const).map((section) => {
										const s = cashFlow[section];
										return (
											<div key={section}>
												<p className="text-[10px] font-semibold uppercase tracking-wide text-secondary-400 mb-1">
													{t(`reports.cashFlow.${section}`)}
												</p>
												{s.lines.map((line) => (
													<div key={line.label} className="flex justify-between">
														<p className="text-[12px] text-secondary-600">{line.label}</p>
														<p
															className={`font-mono text-[12px] ${
																line.amount >= 0 ? "text-success-600" : "text-danger-600"
															}`}
														>
															{line.amount >= 0 ? "+" : ""}
															{currencyFormat(line.amount)}
														</p>
													</div>
												))}
											</div>
										);
									})}
									<div className="border-t border-black/[0.06] pt-2 mt-1 flex justify-between">
										<p className="text-[11px] font-semibold text-secondary-500">
											{t("reports.cashFlow.total")}
										</p>
										<p className="font-mono text-[11px] font-semibold text-success-600">
											+{currencyFormat(cashFlow.netCashFlow)}
										</p>
									</div>
								</div>
							</ReportCard>
						)}
					</>
				)}
			</div>
		</div>
	);
}
