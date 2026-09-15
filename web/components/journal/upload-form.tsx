"use client";

import { useState, useRef, useCallback } from "react";
import { Upload, FileText, X, Download, Info, FileSpreadsheet, FileDown } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { journalService } from "@/services/journal";
import { queryKeys } from "@/lib/query-keys";
import { createLogger } from "@/lib/logger";

const logger = createLogger("UploadForm");

interface UploadFormProps {
  onSuccess: () => void;
}

async function downloadTemplate() {
  try {
    const blob = await journalService.downloadTemplate();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "template-jurnal.xlsx";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  } catch (err) {
    logger.error("Failed to download template", { error: err });
    toast.error("Failed to download template");
  }
}

export function UploadForm({ onSuccess }: UploadFormProps) {
  const t = useTranslations("journalPage.upload");
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile) {
      validateAndSetFile(droppedFile);
    }
  }, []);

  function validateAndSetFile(f: File) {
    const validExtensions = [".csv", ".xlsx", ".xls"];
    const ext = "." + f.name.split(".").pop()?.toLowerCase();

    if (!validExtensions.includes(ext)) {
      toast.error(t("toastInvalidFormat"));
      return;
    }
    if (f.size > 5 * 1024 * 1024) {
      toast.error(t("toastInvalidFormat"));
      return;
    }
    setFile(f);
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) {
      validateAndSetFile(selectedFile);
    }
  }

  function removeFile() {
    setFile(null);
    if (inputRef.current) {
      inputRef.current.value = "";
    }
  }

  async function handleSubmit() {
    if (!file) {
      toast.error(t("toastNoFile"));
      return;
    }

    const toastId = toast.loading(t("toastLoading"));
    setIsSubmitting(true);

    try {
      // Send file (CSV or Excel) directly to backend - backend will parse it
      const result = await journalService.uploadFile(file);
      toast.success(`${t("toastSuccess")} (${result.count} jurnal)`, { id: toastId });
      queryClient.invalidateQueries({ queryKey: queryKeys.journals.all });
      queryClient.invalidateQueries({ queryKey: queryKeys.journals.summary });
      onSuccess();
    } catch (err) {
      logger.error("Failed to upload journals", { error: err });
      toast.error(t("toastError"), { id: toastId });
    } finally {
      setIsSubmitting(false);
    }
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Description */}
      <p className="text-[13px] text-secondary-500 leading-relaxed">
        {t("description")}
      </p>

      {/* Two Options Section */}
      <div className="grid grid-cols-2 gap-3">
        {/* Option 1: Excel Template */}
        <div className="flex flex-col gap-2 p-4 rounded-xl border border-primary-200 bg-primary-50">
          <div className="flex items-center gap-2">
            <FileSpreadsheet size={16} className="text-primary-600" />
            <span className="text-[13px] font-semibold text-primary-700">{t("optionExcelTitle")}</span>
          </div>
          <p className="text-[11px] text-primary-600 leading-relaxed">
            {t("optionExcelDesc")}
          </p>
          <button
            onClick={downloadTemplate}
            className="flex items-center justify-center gap-2 px-3 py-2 rounded-lg
                       bg-primary-500 text-white text-[12px] font-medium
                       hover:bg-primary-600 transition-colors"
          >
            <FileDown size={14} />
            {t("optionExcelButton")}
          </button>
        </div>

        {/* Option 2: CSV */}
        <div className="flex flex-col gap-2 p-4 rounded-xl border border-secondary-200 bg-secondary-50">
          <div className="flex items-center gap-2">
            <FileText size={16} className="text-secondary-600" />
            <span className="text-[13px] font-semibold text-secondary-700">{t("optionCSVTitle")}</span>
          </div>
          <p className="text-[11px] text-secondary-600 leading-relaxed">
            {t("optionCSVDesc")}
          </p>
          <div className="text-[10px] text-secondary-500 font-mono bg-white p-2 rounded-lg border border-secondary-200">
            <p className="font-semibold mb-1">{t("csvFormatTitle")}</p>
            <p>{t("csvFormatExample1")}</p>
            <p>{t("csvFormatExample2")}</p>
            <p>{t("csvFormatExample3")}</p>
          </div>
        </div>
      </div>

      {/* Format Rules */}
      <div className="flex items-start gap-2 p-3 rounded-lg bg-secondary-50 border border-secondary-200">
        <Info size={14} className="text-secondary-500 mt-0.5 shrink-0" />
        <div className="text-[11px] text-secondary-600 space-y-1">
          <p className="font-medium">{t("rulesTitle")}</p>
          <ul className="list-disc list-inside space-y-0.5">
            <li>{t("ruleDateFormat")}</li>
            <li>{t("ruleAccountCode")}</li>
            <li>{t("ruleMinLines")}</li>
            <li>{t("ruleBalanced")}</li>
          </ul>
        </div>
      </div>

      {/* Drop zone */}
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
        className={`
          border-2 border-dashed rounded-2xl p-8 flex flex-col items-center justify-center gap-3
          cursor-pointer transition-all
          ${isDragging
            ? "border-primary-400 bg-primary-50"
            : file
              ? "border-success-300 bg-success-50"
              : "border-secondary-200 hover:border-primary-300 hover:bg-secondary-50"
          }
        `}
      >
        <input
          ref={inputRef}
          type="file"
          accept=".csv,.xlsx,.xls"
          onChange={handleFileChange}
          className="hidden"
        />

        {file ? (
          <>
            <div className="w-12 h-12 rounded-xl bg-success-100 flex items-center justify-center">
              <FileText size={24} className="text-success-600" />
            </div>
            <div className="text-center">
              <p className="text-[13px] font-medium text-secondary-800">{file.name}</p>
              <p className="text-[11px] text-secondary-400">{formatFileSize(file.size)}</p>
            </div>
            <button
              onClick={(e) => {
                e.stopPropagation();
                removeFile();
              }}
              className="flex items-center gap-1 text-[12px] text-danger-500 hover:text-danger-600"
            >
              <X size={14} />
              {t("remove")}
            </button>
          </>
        ) : (
          <>
            <div className="w-12 h-12 rounded-xl bg-secondary-100 flex items-center justify-center">
              <Upload size={24} className="text-secondary-400" />
            </div>
            <div className="text-center">
              <p className="text-[13px] font-medium text-secondary-600">
                {t("dragDrop")}
              </p>
              <p className="text-[12px] text-secondary-400 mt-1">
                {t("or")}{" "}
                <span className="text-primary-500 font-medium">{t("browse")}</span>
              </p>
            </div>
            <p className="text-[11px] text-secondary-300">{t("acceptedFormats")}</p>
          </>
        )}
      </div>

      {/* Submit */}
      <button
        onClick={handleSubmit}
        disabled={!file || isSubmitting}
        className={`
          w-full font-semibold text-[15px] py-3.5 rounded-xl transition-all mt-1
          ${file && !isSubmitting
            ? "bg-primary-500 text-white active:scale-[0.98]"
            : "bg-secondary-100 text-secondary-400 cursor-not-allowed"
          }
        `}
      >
        {t("submit")}
      </button>
    </div>
  );
}
