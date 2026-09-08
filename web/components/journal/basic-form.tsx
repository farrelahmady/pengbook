"use client";

import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { journalService } from "@/services/journal";

const POSTING_ACCOUNTS = [
  { id: "a1", code: "1.01.01.01", name: "Mandiri - Main" },
  { id: "a2", code: "1.01.01.02", name: "BCA - Main" },
  { id: "a3", code: "1.01.02.01", name: "Kas" },
  { id: "a4", code: "1.01.03.01", name: "Gopay" },
  { id: "a5", code: "4.01.01.01", name: "Pendapatan Jasa" },
  { id: "a6", code: "5.01.01.04", name: "Biaya Perlengkapan" },
  { id: "a7", code: "5.04.01.01", name: "Beban Gaji" },
  { id: "a8", code: "2.01.01.00", name: "Hutang Dagang" },
];

interface BasicFormProps {
  onSuccess: () => void;
}

export function BasicForm({ onSuccess }: BasicFormProps) {
  const t = useTranslations("journalPage.basic");
  const queryClient = useQueryClient();
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [description, setDesc] = useState("");
  const [fromAccount, setFrom] = useState("");
  const [toAccount, setTo] = useState("");
  const [amount, setAmount] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit() {
    if (!fromAccount || !toAccount || !amount) {
      toast.error(t("toastValidation"));
      return;
    }
    if (fromAccount === toAccount) {
      toast.error(t("toastSameAccount"));
      return;
    }

    const toastId = toast.loading(t("toastLoading"));
    setIsSubmitting(true);

    try {
      const from = POSTING_ACCOUNTS.find((a) => a.id === fromAccount);
      const to = POSTING_ACCOUNTS.find((a) => a.id === toAccount);

      await journalService.create({
        date: new Date(date).toISOString(),
        description: description || undefined,
        lines: [
          { accountId: toAccount, debit: parseFloat(amount), credit: 0 },
          { accountId: fromAccount, debit: 0, credit: parseFloat(amount) },
        ],
      });

      toast.success(t("toastSuccess"), { id: toastId });
      queryClient.invalidateQueries({ queryKey: ["journals"] });
      queryClient.invalidateQueries({ queryKey: ["journalSummary"] });
      onSuccess();
    } catch (err) {
      toast.error(t("toastError"), { id: toastId });
    } finally {
      setIsSubmitting(false);
    }
  }

  const selectClass =
    "w-full bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 appearance-none cursor-pointer";
  const labelClass =
    "text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";

  return (
    <div className="flex flex-col gap-4">
      {/* Date */}
      <div>
        <label className={labelClass}>{t("date")}</label>
        <input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className={selectClass}
        />
      </div>

      {/* Description */}
      <div>
        <label className={labelClass}>{t("description")}</label>
        <input
          type="text"
          placeholder={t("descriptionPlaceholder")}
          value={description}
          onChange={(e) => setDesc(e.target.value)}
          className={selectClass}
        />
      </div>

      {/* From */}
      <div>
        <label className={labelClass}>{t("fromAccount")}</label>
        <select
          value={fromAccount}
          onChange={(e) => setFrom(e.target.value)}
          className={selectClass}
        >
          <option value="">{t("fromPlaceholder")}</option>
          {POSTING_ACCOUNTS.map((a) => (
            <option key={a.id} value={a.id}>
              {a.code} · {a.name}
            </option>
          ))}
        </select>
      </div>

      {/* Arrow visual */}
      <div className="flex items-center gap-3 -my-1">
        <div className="flex-1 border-t-2 border-dashed border-secondary-200" />
        <div className="w-8 h-8 rounded-full bg-primary-50 flex items-center justify-center shrink-0">
          <ArrowRight size={14} className="text-primary-500" />
        </div>
        <div className="flex-1 border-t-2 border-dashed border-secondary-200" />
      </div>

      {/* To */}
      <div>
        <label className={labelClass}>{t("toAccount")}</label>
        <select
          value={toAccount}
          onChange={(e) => setTo(e.target.value)}
          className={selectClass}
        >
          <option value="">{t("toPlaceholder")}</option>
          {POSTING_ACCOUNTS.map((a) => (
            <option key={a.id} value={a.id}>
              {a.code} · {a.name}
            </option>
          ))}
        </select>
      </div>

      {/* Amount */}
      <div>
        <label className={labelClass}>{t("amount")}</label>
        <div className="relative">
          <span className="absolute left-3.5 top-1/2 -translate-y-1/2 font-mono text-[13px] text-secondary-400">
            Rp
          </span>
          <input
            type="number"
            placeholder="0"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className={`${selectClass} pl-10 font-mono text-[16px]`}
          />
        </div>
      </div>

      {/* Auto hint */}
      <div className="bg-primary-50 rounded-xl px-3.5 py-3 text-[12px] text-primary-700 leading-relaxed">
        <span className="font-semibold">{t("hint")}</span>
        <br />
        <span className="font-mono text-[11px]">
          {t("hintDetail")}
        </span>
      </div>

      {/* Submit */}
      <button
        onClick={handleSubmit}
        disabled={isSubmitting}
        className="w-full bg-primary-500 text-white font-semibold text-[15px] py-3.5 rounded-xl
                   active:scale-[0.98] transition-transform mt-1 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {t("submit")}
      </button>
    </div>
  );
}
