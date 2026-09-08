"use client";

import { useState } from "react";
import { Plus, Minus } from "lucide-react";
import { formatNumber, parseDecimal } from "@/lib/utils";
import { cn } from "@/lib/utils";
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

interface JournalLine {
  accountId: string;
  debit: string;
  credit: string;
}

interface AdvancedFormProps {
  onSuccess: () => void;
}

export function AdvancedForm({ onSuccess }: AdvancedFormProps) {
  const t = useTranslations("journalPage.advanced");
  const queryClient = useQueryClient();
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [description, setDesc] = useState("");
  const [lines, setLines] = useState<JournalLine[]>([
    { accountId: "", debit: "", credit: "" },
    { accountId: "", debit: "", credit: "" },
  ]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const totalDr = lines.reduce((s, l) => s + parseDecimal(l.debit), 0);
  const totalCr = lines.reduce((s, l) => s + parseDecimal(l.credit), 0);
  const isBalanced = Math.abs(totalDr - totalCr) < 0.01 && totalDr > 0;

  function updateLine(i: number, field: keyof JournalLine, value: string) {
    setLines((prev) =>
      prev.map((l, idx) => (idx === i ? { ...l, [field]: value } : l)),
    );
  }

  function addLine() {
    setLines((prev) => [...prev, { accountId: "", debit: "", credit: "" }]);
  }

  function removeLine(i: number) {
    if (lines.length <= 2) return;
    setLines((prev) => prev.filter((_, idx) => idx !== i));
  }

  async function handleSubmit() {
    if (!isBalanced) {
      toast.error(t("toastUnbalanced"));
      return;
    }

    const toastId = toast.loading(t("toastLoading"));
    setIsSubmitting(true);

    try {
      await journalService.create({
        date: new Date(date).toISOString(),
        description: description || undefined,
        lines: lines.map((l) => ({
          accountId: l.accountId,
          debit: parseDecimal(l.debit),
          credit: parseDecimal(l.credit),
        })),
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

  const labelClass =
    "text-[11px] font-semibold uppercase tracking-[0.5px] text-secondary-400 mb-1.5 block";
  const inputClass =
    "bg-secondary-50 border border-secondary-200 rounded-xl px-3.5 py-3 text-[13px] text-secondary-800 outline-none focus:border-primary-400 w-full";

  return (
    <div className="flex flex-col gap-4">
      {/* Date */}
      <div>
        <label className={labelClass}>{t("date")}</label>
        <input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className={inputClass}
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
          className={inputClass}
        />
      </div>

      {/* Journal lines table */}
      <div>
        <label className={labelClass}>{t("lines")}</label>
        <div className="border border-secondary-200 rounded-xl overflow-hidden">
          {/* Head */}
          <div className="grid grid-cols-[1fr_80px_80px_28px] gap-1 px-3 py-2 bg-secondary-50 border-b border-secondary-200">
            {[t("account"), t("debit"), t("credit"), ""].map((h, i) => (
              <span
                key={i}
                className={cn(
                  "text-[10px] font-semibold uppercase tracking-[0.4px] text-secondary-400",
                  i > 0 && "text-right",
                )}
              >
                {h}
              </span>
            ))}
          </div>

          {/* Rows */}
          {lines.map((line, i) => (
            <div
              key={i}
              className="grid grid-cols-[1fr_80px_80px_28px] gap-1 px-2 py-2 border-b border-secondary-100 last:border-none items-center"
            >
              <select
                value={line.accountId}
                onChange={(e) => updateLine(i, "accountId", e.target.value)}
                className="text-[11px] bg-secondary-50 border border-secondary-200 rounded-lg px-2 py-1.5 w-full appearance-none text-secondary-700 outline-none focus:border-primary-400"
              >
                <option value="">{t("accountPlaceholder")}</option>
                {POSTING_ACCOUNTS.map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.code}
                  </option>
                ))}
              </select>

              <input
                type="number"
                placeholder="0"
                value={line.debit}
                onChange={(e) => updateLine(i, "debit", e.target.value)}
                className="text-[11px] font-mono bg-secondary-50 border border-secondary-200 rounded-lg px-2 py-1.5 text-right w-full outline-none focus:border-success-400 text-success-700"
              />
              <input
                type="number"
                placeholder="0"
                value={line.credit}
                onChange={(e) => updateLine(i, "credit", e.target.value)}
                className="text-[11px] font-mono bg-secondary-50 border border-secondary-200 rounded-lg px-2 py-1.5 text-right w-full outline-none focus:border-danger-400 text-danger-700"
              />

              <button
                onClick={() => removeLine(i)}
                disabled={lines.length <= 2}
                className="w-7 h-7 rounded-lg bg-danger-50 flex items-center justify-center text-danger-500 disabled:opacity-30"
              >
                <Minus size={13} />
              </button>
            </div>
          ))}
        </div>

        {/* Add row */}
        <button
          onClick={addLine}
          className="w-full mt-2 flex items-center justify-center gap-2 border-2 border-dashed border-secondary-200
                     rounded-xl py-2.5 text-[13px] font-medium text-secondary-400 hover:border-primary-300
                     hover:text-primary-500 transition-colors"
        >
          <Plus size={14} />
          {t("addRow")}
        </button>
      </div>

      {/* Balance strip */}
      <div
        className={cn(
          "rounded-xl px-4 py-3 flex items-center justify-between border",
          isBalanced
            ? "bg-success-50 border-success-200"
            : "bg-secondary-50 border-secondary-200",
        )}
      >
        <span className="text-[12px] font-semibold text-secondary-600">
          {t("balance")}
        </span>
        <div className="flex gap-4 items-center">
          <div className="text-right">
            <p className="text-[9px] uppercase tracking-wide text-secondary-400">
              {t("debit")}
            </p>
            <p className="font-mono text-[12px] font-semibold text-success-700">
              {formatNumber(totalDr)}
            </p>
          </div>
          <div className="text-right">
            <p className="text-[9px] uppercase tracking-wide text-secondary-400">
              {t("credit")}
            </p>
            <p className="font-mono text-[12px] font-semibold text-danger-700">
              {formatNumber(totalCr)}
            </p>
          </div>
          <div
            className={cn(
              "text-[12px] font-bold font-mono",
              isBalanced ? "text-success-600" : "text-secondary-400",
            )}
          >
            {totalDr === 0 && totalCr === 0
              ? "—"
              : isBalanced
                ? t("balanced")
                : `Δ ${formatNumber(Math.abs(totalDr - totalCr))}`}
          </div>
        </div>
      </div>

      {/* Submit */}
      <button
        onClick={handleSubmit}
        disabled={!isBalanced || isSubmitting}
        className={cn(
          "w-full font-semibold text-[15px] py-3.5 rounded-xl transition-all mt-1",
          isBalanced && !isSubmitting
            ? "bg-primary-500 text-white active:scale-[0.98]"
            : "bg-secondary-100 text-secondary-400 cursor-not-allowed",
        )}
      >
        {t("submit")}
      </button>
    </div>
  );
}
