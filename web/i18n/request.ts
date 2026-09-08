import { getRequestConfig } from "next-intl/server";
import { routing } from "./routing";
import deepmerge from "deepmerge"; // npm install deepmerge

export default getRequestConfig(async ({ requestLocale }) => {
	let locale = await requestLocale;
	if (!locale || !routing.locales.includes(locale as any)) {
		locale = routing.defaultLocale;
	}

	const defaultMessages = (
		await import(`../messages/${routing.defaultLocale}.json`)
	).default;

	// Kalau locale aktif = default locale, gak perlu merge
	const messages =
		locale === routing.defaultLocale
			? defaultMessages
			: deepmerge(
					defaultMessages,
					(await import(`../messages/${locale}.json`)).default,
				);

	return {
		locale,
		messages,
	};
});
