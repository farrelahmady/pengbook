"use client";
import { JournalScrollView } from "@/components/journal/journal-scroll-view";
import { useTranslations } from "next-intl";
import { useCallback, useState } from "react";
import { CalendarDays, X } from "lucide-react";
import { POSTING_ACCOUNTS } from "@/lib/constants";

const QUICK_FILTERS = ["all", "today", "week", "month"] as const;

export default function JournalList() {
	const t = useTranslations("journalPage");
	const [activeQuickFilter, setActiveQuickFilter] = useState<string>("today");
	const [dateFrom, setDateFrom] = useState("");
	const [dateTo, setDateTo] = useState("");
	const [selectedAccountIds, setSelectedAccountIds] = useState<string[]>([]);
	const [showDatePicker, setShowDatePicker] = useState(false);
	const [showCoaPicker, setShowCoaPicker] = useState(false);

	const hasCustomDate = dateFrom || dateTo;
	const hasCoaFilter = selectedAccountIds.length > 0;

	const getDateRange = useCallback(() => {
		if (hasCustomDate) {
			return {
				startDate: dateFrom ? new Date(dateFrom) : undefined,
				endDate: dateTo
					? new Date(dateTo + "T23:59:59.999")
					: undefined,
			};
		}

		const today = new Date();
		const endOfToday = new Date(
			today.getFullYear(),
			today.getMonth(),
			today.getDate(),
			23,
			59,
			59,
			999,
		);

		switch (activeQuickFilter) {
			case "all":
				return { startDate: undefined, endDate: undefined };
			case "today":
				return { startDate: today, endDate: endOfToday };
			case "week": {
				const firstDay = new Date(today);
				firstDay.setDate(today.getDate() - today.getDay());
				return { startDate: firstDay, endDate: endOfToday };
			}
			case "month": {
				const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
				return { startDate: firstDay, endDate: endOfToday };
			}
			default:
				return { startDate: undefined, endDate: undefined };
		}
	}, [activeQuickFilter, hasCustomDate, dateFrom, dateTo]);

	const { startDate, endDate } = getDateRange();

	function toggleAccount(accountId: string) {
		setSelectedAccountIds((prev) =>
			prev.includes(accountId)
				? prev.filter((id) => id !== accountId)
				: [...prev, accountId],
		);
	}

	function clearAllFilters() {
		setDateFrom("");
		setDateTo("");
		setSelectedAccountIds([]);
		setActiveQuickFilter("today");
	}

	const hasActiveFilters = hasCustomDate || hasCoaFilter;

	return (
		<>
			{/* ── Quick filter chips ── */}
			<div className="flex gap-2 px-4 py-3 overflow-x-auto scrollbar-none">
				{QUICK_FILTERS.map((f) => (
					<button
						key={f}
						onClick={() => {
							setActiveQuickFilter(f);
							setDateFrom("");
							setDateTo("");
						}}
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
						setShowCoaPicker(false);
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

				{/* COA filter button */}
				<button
					onClick={() => {
						setShowCoaPicker(!showCoaPicker);
						setShowDatePicker(false);
					}}
					className={[
						"shrink-0 flex items-center gap-1.5 px-4 py-1.5 rounded-full text-[12px] font-semibold border transition-all no-tap",
						hasCoaFilter
							? "bg-primary-500 text-white border-primary-500"
							: "bg-white text-secondary-500 border-secondary-200 hover:border-primary-300",
					].join(" ")}
				>
					{t("filter.coa")}
					{hasCoaFilter && (
						<span className="w-4 h-4 rounded-full bg-white/20 flex items-center justify-center text-[10px]">
							{selectedAccountIds.length}
						</span>
					)}
				</button>

				{/* Clear filters */}
				{hasActiveFilters && (
					<button
						onClick={clearAllFilters}
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
									onChange={(e) => {
										setDateFrom(e.target.value);
										setActiveQuickFilter("");
									}}
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
									onChange={(e) => {
										setDateTo(e.target.value);
										setActiveQuickFilter("");
									}}
									className="w-full bg-secondary-50 border border-secondary-200 rounded-lg px-3 py-2 text-[12px] text-secondary-700 outline-none focus:border-primary-400"
								/>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* ── COA multi-select ── */}
			{showCoaPicker && (
				<div className="px-4 pb-3">
					<div className="card-default shadow-card p-3 max-h-60 overflow-y-auto">
						<p className="text-[11px] font-semibold uppercase tracking-wide text-secondary-400 mb-2">
							{t("filter.selectAccount")}
						</p>
						<div className="flex flex-col gap-1">
							{POSTING_ACCOUNTS.map((account) => {
								const isSelected = selectedAccountIds.includes(account.id);
								return (
									<button
										key={account.id}
										onClick={() => toggleAccount(account.id)}
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
					</div>
				</div>
			)}

			{/* ── List ── */}
			<div className="px-1">
				<p className="section-label">{t("listTitle")}</p>
				<JournalScrollView
					startDate={startDate}
					endDate={endDate}
					accountIds={selectedAccountIds.length > 0 ? selectedAccountIds : undefined}
				/>
			</div>
		</>
	);
}
