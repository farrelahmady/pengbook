"use client";

import { useState } from "react";
import { Link } from "@/i18n/navigation";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useAuth } from "@/lib/auth-context";

export function RegisterForm() {
	const t = useTranslations("auth.register");
	const { register } = useAuth();
	const [name, setName] = useState("");
	const [username, setUsername] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	const inputClass =
		"w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400";
	const labelClass =
		"text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();

		if (!name || !username || !email || !password) {
			toast.error(t("toastValidation"));
			return;
		}

		if (password.length < 8) {
			toast.error(t("toastPasswordMin"));
			return;
		}

		setIsSubmitting(true);
		const toastId = toast.loading(t("toastLoading"));

		try {
			await register(name, username, email, password);
			toast.success(t("toastSuccess"), { id: toastId });
		} catch (err: any) {
			const msg =
				err?.response?.message ||
				err?.message ||
				t("toastError");
			toast.error(msg, { id: toastId });
		} finally {
			setIsSubmitting(false);
		}
	}

	return (
		<div className="min-h-dvh flex flex-col">
			{/* Gradient header */}
			<div
				className="px-5 pt-12 pb-8"
				style={{
					background:
						"linear-gradient(160deg, #1a1a2e 70%, #23309a 130%)",
				}}
			>
				<h1 className="text-xl font-bold tracking-tight text-white leading-none">
					Peng<span className="text-primary-300">book</span>
				</h1>
				<p className="text-[11px] text-white/40 font-medium mt-0.5 tracking-wide">
					{t("subtitle")}
				</p>
			</div>

			{/* Form card */}
			<div className="flex-1 px-4 -mt-4">
				<div className="bg-white rounded-2xl border border-black/[0.08] p-5">
					<h2 className="text-[17px] font-bold text-secondary-900 mb-4">
						{t("title")}
					</h2>

					<form onSubmit={handleSubmit} className="flex flex-col gap-4">
						{/* Name */}
						<div>
							<label className={labelClass}>{t("name")}</label>
							<input
								type="text"
								placeholder={t("namePlaceholder")}
								value={name}
								onChange={(e) => setName(e.target.value)}
								className={inputClass}
								autoComplete="name"
							/>
						</div>

						{/* Username */}
						<div>
							<label className={labelClass}>{t("username")}</label>
							<input
								type="text"
								placeholder={t("usernamePlaceholder")}
								value={username}
								onChange={(e) => setUsername(e.target.value)}
								className={inputClass}
								autoComplete="username"
							/>
						</div>

						{/* Email */}
						<div>
							<label className={labelClass}>{t("email")}</label>
							<input
								type="email"
								placeholder={t("emailPlaceholder")}
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								className={inputClass}
								autoComplete="email"
							/>
						</div>

						{/* Password */}
						<div>
							<label className={labelClass}>{t("password")}</label>
							<input
								type="password"
								placeholder={t("passwordPlaceholder")}
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								className={inputClass}
								autoComplete="new-password"
							/>
							<p className="text-[11px] text-secondary-400 mt-1.5">
								{t("passwordHint")}
							</p>
						</div>

						{/* Submit */}
						<button
							type="submit"
							disabled={isSubmitting}
							className="w-full bg-primary-500 text-white font-semibold text-[15px] py-3.5 rounded-xl
                                       active:scale-[0.98] transition-transform mt-1 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{t("submit")}
						</button>
					</form>
				</div>

				{/* Login link */}
				<p className="text-center text-[13px] text-secondary-500 mt-4">
					{t("hasAccount")}{" "}
					<Link
						href="/login"
						className="text-primary-500 font-semibold hover:text-primary-600 transition-colors"
					>
						{t("loginLink")}
					</Link>
				</p>
			</div>
		</div>
	);
}
