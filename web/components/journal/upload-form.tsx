"use client";

import { useState, useRef, useCallback } from "react";
import { Upload, FileText, X, Download } from "lucide-react";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import { useQueryClient } from "@tanstack/react-query";
import { journalService } from "@/services/journal";

interface UploadFormProps {
  onSuccess: () => void;
}

const CSV_TEMPLATE = `Tanggal,Deskripsi,Kode Akun,Debit,Kredit
2026-04-21,Pembelian perlengkapan,5.01.01.04,500000,0
2026-04-21,Pembelian perlengkapan,1.01.02.01,0,500000
2026-04-20,Pendapatan jasa,1.01.01.02,2000000,0
2026-04-20,Pendapatan jasa,4.01.01.01,0,2000000`;

function downloadTemplate() {
  const blob = new Blob([CSV_TEMPLATE], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "template-jurnal.csv";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
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
    const validTypes = [
      "text/csv",
      "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ];
    const validExtensions = [".csv", ".xlsx", ".xls"];
    const ext = "." + f.name.split(".").pop()?.toLowerCase();

    if (!validTypes.includes(f.type) && !validExtensions.includes(ext)) {
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
      const result = await journalService.createBulk(file);
      toast.success(`${t("toastSuccess")} (${result.count} jurnal)`, { id: toastId });
      queryClient.invalidateQueries({ queryKey: ["journals"] });
      queryClient.invalidateQueries({ queryKey: ["journalSummary"] });
      onSuccess();
    } catch (err) {
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

      {/* Download Template */}
      <button
        onClick={downloadTemplate}
        className="flex items-center gap-2 px-4 py-2.5 rounded-xl border border-secondary-200
                   text-[13px] font-medium text-secondary-600 hover:bg-secondary-50 transition-colors"
      >
        <Download size={16} />
        {t("downloadTemplate")}
      </button>

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
