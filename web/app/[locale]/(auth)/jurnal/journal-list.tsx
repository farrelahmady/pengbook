"use client";
import { JournalScrollView } from "@/components/journal/journal-scroll-view";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { CalendarDays, X } from "lucide-react";
import { accountService } from "@/services/account";
import { queryKeys } from "@/lib/query-keys";

const QUICK_FILTERS = ["all", "today", "week", "month"] as const;

interface JournalListProps {
	activeQuickFilter: string;
	dateFrom: string;
	dateTo: string;
	selectedAccountIds: string[];
	startDate?: Date;
	endDate?: Date;
	onQuickFilterChange: (f: string) => void;
	onDateFromChange: (v: string) => void;
	onDateToChange: (v: string) => void;
	onToggleAccount: (accountId: string) => void;
	onClearAllFilters: () => void;
}

export default function JournalList({
	activeQuickFilter,
	dateFrom,
	dateTo,
	selectedAccountIds,
	startDate,
	endDate,
	onQuickFilterChange,
	onDateFromChange,
	onDateToChange,
	onToggleAccount,
	onClearAllFilters,
}: JournalListProps) {
	const t = useTranslations("journalPage");
	const [showDatePicker, setShowDatePicker] = useState(false);
	const [showAccountPicker, setShowAccountPicker] = useState(false);

	const { data: postingAccounts = [], isLoading: isLoadingAccounts } = useQuery(
		{
			queryKey: queryKeys.accounts.posting,
			queryFn: () => accountService.getPostingAccounts(),
		},
	);

	const hasCustomDate = dateFrom || dateTo;
	const hasAccountFilter = selectedAccountIds.length > 0;

	const hasActiveFilters = hasCustomDate || hasAccountFilter;

	return (
		<>
			{/* ── Quick filter chips ── */}
			<div className="flex gap-2 px-4 py-3 overflow-x-auto scrollbar-none">
				{QUICK_FILTERS.map((f) => (
					<button
						key={f}
						onClick={() => onQuickFilterChange(f)}
						className={[
							"shrink-0 px-4 py-1.5 rounded-full text-[12px] font-semibold border transition-all no-tap",
							activeQuickFilter === f && !hasCustomDate
								? "bg-primary-500 text-white border-primary-500"
								: "bg-white text-secondary-500 border-secondary-200 hover:border-primary-300",
						].join(" ")}
					>
						{t(`transactionFilterChips.${f}`)}
					</button>
				))}

				{/* Date range button */}
				<button
					onClick={() => {
						setShowDatePicker(!showDatePicker);
						setShowAccountPicker(false);
					}}
					className={[
						"shrink-0 flex items-center gap-1.5 px-4 py-1.5 rounded-full text-[12px] font-semibold border transition-all no-tap",
						hasCustomDate
							? "bg-primary-500 text-white border-primary-500"
							: "bg-white text-secondary-500 border-secondary-200 hover:border-primary-300",
					].join(" ")}
				>
					<CalendarDays size={12} />
					{t("filter.dateRange")}
				</button>

				{/* Account filter button */}
				<button
					onClick={() => {
						setShowAccountPicker(!showAccountPicker);
						setShowDatePicker(false);
					}}
					className={[
						"shrink-0 flex items-center gap-1.5 px-4 py-1.5 rounded-full text-[12px] font-semibold border transition-all no-tap",
						hasAccountFilter
							? "bg-primary-500 text-white border-primary-500"
							: "bg-white text-secondary-500 border-secondary-200 hover:border-primary-300",
					].join(" ")}
				>
					{t("filter.account")}
					{hasAccountFilter && (
						<span className="w-4 h-4 rounded-full bg-white/20 flex items-center justify-center text-[10px]">
							{selectedAccountIds.length}
						</span>
					)}
				</button>

				{/* Clear filters */}
				{hasActiveFilters && (
					<button
						onClick={onClearAllFilters}
						className="shrink-0 flex items-center gap-1 px-3 py-1.5 rounded-full text-[12px] font-semibold
                       border border-danger-200 text-danger-500 bg-danger-50 hover:bg-danger-100 transition-all no-tap"
					>
						<X size={12} />
						{t("filter.clear")}
					</button>
				)}
			</div>

			{/* ── Date range picker ── */}
			{showDatePicker && (
				<div className="px-4 pb-3">
					<div className="card-default shadow-card p-3 flex flex-col gap-2">
						<p className="text-[11px] font-semibold uppercase tracking-wide text-secondary-400 mb-1">
							{t("filter.selectDateRange")}
						</p>
						<div className="flex gap-2">
							<div className="flex-1">
								<label className="text-[10px] text-secondary-400 block mb-1">
									{t("filter.from")}
								</label>
								<input
									type="date"
									value={dateFrom}
									onChange={(e) => onDateFromChange(e.target.value)}
									className="w-full bg-secondary-50 border border-secondary-200 rounded-lg px-3 py-2 text-[12px] text-secondary-700 outline-none focus:border-primary-400"
								/>
							</div>
							<div className="flex-1">
								<label className="text-[10px] text-secondary-400 block mb-1">
									{t("filter.to")}
								</label>
								<input
									type="date"
									value={dateTo}
									onChange={(e) => onDateToChange(e.target.value)}
									className="w-full bg-secondary-50 border border-secondary-200 rounded-lg px-3 py-2 text-[12px] text-secondary-700 outline-none focus:border-primary-400"
								/>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* ── Account multi-select ── */}
			{showAccountPicker && (
				<div className="px-4 pb-3">
					<div className="card-default shadow-card p-3 max-h-60 overflow-y-auto">
						<p className="text-[11px] font-semibold uppercase tracking-wide text-secondary-400 mb-2">
							{t("filter.selectAccount")}
						</p>
						{isLoadingAccounts ? (
							<p className="text-[12px] text-secondary-400 px-3 py-2">
								Loading...
							</p>
						) : (
							<div className="flex flex-col gap-1">
								{postingAccounts.map((account) => {
									const id = account.id.toString();
									const isSelected = selectedAccountIds.includes(id);
									return (
										<button
											key={account.id}
											onClick={() => onToggleAccount(id)}
											className={[
												"flex items-center gap-3 px-3 py-2.5 rounded-lg text-left transition-all",
												isSelected
													? "bg-primary-50 border border-primary-200"
													: "hover:bg-secondary-50 border border-transparent",
											].join(" ")}
										>
											<div
												className={[
													"w-4 h-4 rounded border-2 flex items-center justify-center shrink-0 transition-all",
													isSelected
														? "bg-primary-500 border-primary-500"
														: "border-secondary-300",
												].join(" ")}
											>
												{isSelected && (
													<svg
														width="10"
														height="8"
														viewBox="0 0 10 8"
														fill="none"
													>
														<path
															d="M1 4L3.5 6.5L9 1"
															stroke="white"
															strokeWidth="2"
															strokeLinecap="round"
															strokeLinejoin="round"
														/>
													</svg>
												)}
											</div>
											<div className="flex-1 min-w-0">
												<p className="text-[12px] font-medium text-secondary-700 truncate">
													{account.name}
												</p>
												<p className="font-mono text-[10px] text-secondary-400">
													{account.code}
												</p>
											</div>
										</button>
									);
								})}
							</div>
						)}
					</div>
				</div>
			)}

			{/* ── List ── */}
			<div className="px-1">
				<p className="section-label">{t("listTitle")}</p>
				<JournalScrollView
					startDate={startDate}
					endDate={endDate}
					accountIds={
						selectedAccountIds.length > 0 ? selectedAccountIds : undefined
					}
				/>
			</div>
		</>
	);
}
