"use client";
import { useCallback, useState } from "react";
import JournalTopbar from "./journal-topbar";
import JournalList from "./journal-list";

/**
 * Owns the journal page filter state shared by the topbar (export) and the
 * list (display). The list UI itself stays in JournalList; this component
 * only holds data + handlers so both children observe the same filters.
 */
export default function JurnalPageContent() {
	const [activeQuickFilter, setActiveQuickFilter] = useState<string>("all");
	const [dateFrom, setDateFrom] = useState("");
	const [dateTo, setDateTo] = useState("");
	const [selectedAccountIds, setSelectedAccountIds] = useState<string[]>([]);

	const hasCustomDate = dateFrom || dateTo;

	const getDateRange = useCallback(() => {
		if (hasCustomDate) {
			return {
				startDate: dateFrom ? new Date(dateFrom) : undefined,
				endDate: dateTo ? new Date(dateTo + "T23:59:59.999") : undefined,
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

	function handleQuickFilterChange(f: string) {
		setActiveQuickFilter(f);
		setDateFrom("");
		setDateTo("");
	}

	function handleDateFromChange(v: string) {
		setDateFrom(v);
		setActiveQuickFilter("");
	}

	function handleDateToChange(v: string) {
		setDateTo(v);
		setActiveQuickFilter("");
	}

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

	return (
		<>
			<JournalTopbar
				startDate={startDate}
				endDate={endDate}
				accountIds={
					selectedAccountIds.length > 0 ? selectedAccountIds : undefined
				}
			/>
			<JournalList
				activeQuickFilter={activeQuickFilter}
				dateFrom={dateFrom}
				dateTo={dateTo}
				selectedAccountIds={selectedAccountIds}
				startDate={startDate}
				endDate={endDate}
				onQuickFilterChange={handleQuickFilterChange}
				onDateFromChange={handleDateFromChange}
				onDateToChange={handleDateToChange}
				onToggleAccount={toggleAccount}
				onClearAllFilters={clearAllFilters}
			/>
		</>
	);
}
