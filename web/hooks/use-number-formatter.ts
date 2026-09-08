// hooks/formatters/use-number-formatter.ts
import { useFormatter, useTranslations } from "next-intl";

export type NumberFormatOptions = {
	compact?: boolean;
	showSign?: boolean;
	decimals?: number;
};

export function useNumberFormatter() {
	const format = useFormatter();
	const t = useTranslations("number");

	return function formatNumber(
		value: number | string,
		options: NumberFormatOptions = {},
	): string {
		const num = typeof value === "string" ? parseFloat(value) : value;
		const { compact = false, showSign = false, decimals = 0 } = options;

		if (isNaN(num)) return "—";

		const sign = showSign && num > 0 ? "+" : "";
		const absNum = Math.abs(num);

		if (compact) {
			if (absNum >= 1_000_000_000_000) {
				const f = format.number(num / 1_000_000_000_000, {
					maximumFractionDigits: 1,
				});
				return `${sign}${f} ${t("trillion")}`;
			}
			if (absNum >= 1_000_000_000) {
				const f = format.number(num / 1_000_000_000, {
					maximumFractionDigits: 1,
				});
				return `${sign}${f} ${t("billion")}`;
			}
			if (absNum >= 1_000_000) {
				const f = format.number(num / 1_000_000, { maximumFractionDigits: 1 });
				return `${sign}${f} ${t("million")}`;
			}
			if (absNum >= 1_000) {
				const f = format.number(num / 1_000, { maximumFractionDigits: 1 });
				return `${sign}${f} ${t("thousand")}`;
			}
		}

		const formatted = format.number(absNum, {
			minimumFractionDigits: decimals,
			maximumFractionDigits: decimals,
		});

		return `${sign}${num < 0 ? "-" : ""}${formatted}`;
	};
}
