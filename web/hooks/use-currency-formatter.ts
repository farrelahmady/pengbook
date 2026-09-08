// hooks/formatters/use-currency-formatter.ts
import {
	useNumberFormatter,
	type NumberFormatOptions,
} from "./use-number-formatter";

export type CurrencyFormatOptions = NumberFormatOptions & {
	currency?: string; // default "Rp"
};

export function useCurrencyFormatter() {
	const formatNumber = useNumberFormatter();

	return function formatCurrency(
		value: number | string,
		options: CurrencyFormatOptions = {},
	): string {
		const { currency = "Rp", ...numberOptions } = options;
		const num = typeof value === "string" ? parseFloat(value) : value;

		if (isNaN(num)) return `${currency} —`;

		const formatted = formatNumber(num, numberOptions);

		// formatted sudah include sign/minus, currency disisipkan setelah sign
		if (formatted.startsWith("+") || formatted.startsWith("-")) {
			return `${formatted[0]}${currency} ${formatted.slice(1)}`;
		}

		return `${currency} ${formatted}`;
	};
}
