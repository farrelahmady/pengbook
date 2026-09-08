"use client";

import { useTheme } from "next-themes";
import { Toaster as Sonner, type ToasterProps } from "sonner";
import {
	CheckCircleIcon,
	InfoIcon,
	WarningIcon,
	XCircleIcon,
	SpinnerIcon,
} from "@phosphor-icons/react";

const Toaster = ({ ...props }: ToasterProps) => {
	const { theme = "system" } = useTheme();

	return (
		<Sonner
			theme={theme as ToasterProps["theme"]}
			className="toaster group font-sans"
			icons={{
				success: <CheckCircleIcon className="size-4" />,
				info: <InfoIcon className="size-4" />,
				warning: <WarningIcon className="size-4" />,
				error: <XCircleIcon className="size-4" />,
				loading: <SpinnerIcon className="size-4 animate-spin" />,
			}}
			style={
				{
					/* ── Default (loading / normal) ─────────────────────── */
					"--normal-bg": "var(--color-popover)",
					"--normal-text": "var(--color-popover-foreground)",
					"--normal-border": "var(--color-border)",

					/* ── Success ────────────────────────────────────────── */
					"--success-bg": "var(--color-success-50)",
					"--success-text": "var(--color-success-700)",
					"--success-border": "var(--color-success-200)",

					/* ── Error / Danger ─────────────────────────────────── */
					"--error-bg": "var(--color-danger-50)",
					"--error-text": "var(--color-danger-700)",
					"--error-border": "var(--color-danger-200)",

					/* ── Info ───────────────────────────────────────────── */
					"--info-bg": "var(--color-info-50)",
					"--info-text": "var(--color-info-700)",
					"--info-border": "var(--color-info-200)",

					/* ── Warning ────────────────────────────────────────── */
					"--warning-bg": "var(--color-warning-50)",
					"--warning-text": "var(--color-warning-700)",
					"--warning-border": "var(--color-warning-200)",

					/* ── Misc ───────────────────────────────────────────── */
					"--border-radius": "var(--radius)",
				} as React.CSSProperties
			}
			toastOptions={{
				classNames: {
					toast: "cn-toast",
					title: "line-clamp-2",
					description: "text-muted-foreground",
				},
			}}
			richColors
			{...props}
		/>
	);
};

export { Toaster };
