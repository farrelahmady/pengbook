import { NextIntlClientProvider } from "next-intl";
import { getMessages } from "next-intl/server";
import { routing } from "@/i18n/routing";
import { notFound } from "next/navigation";
import { Syne } from "next/font/google";
import { Metadata, Viewport } from "next";
import { MobileLayout } from "@/components/layout/mobile-layout";
import { Providers } from "./providers";
import "../globals.css";

export const metadata: Metadata = {
	title: "Pengbook",
	description: "Aplikasi Pencatatan Keuangan",
};

export const viewport: Viewport = {
	width: "device-width",
	initialScale: 1,
	maximumScale: 1,
	userScalable: false,
	themeColor: "#1a1a2e",
};

const syne = Syne({
	subsets: ["latin"],
	weight: ["400", "500", "600", "700"],
	variable: "--font-sans",
	display: "swap",
});

export default async function LocaleLayout({
	children,
	params,
}: {
	children: React.ReactNode;
	params: Promise<{ locale: string }>;
}) {
	const { locale } = await params;
	if (!routing.locales.includes(locale as any)) {
		notFound();
	}

	const messages = await getMessages();

	return (
		<html lang={locale}>
			<body className={`bg-secondary-100 antialiased ${syne.variable}`}>
				<MobileLayout>
					<NextIntlClientProvider messages={messages}>
						<Providers>{children}</Providers>
					</NextIntlClientProvider>
				</MobileLayout>
			</body>
		</html>
	);
}
