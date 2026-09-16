"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import AssetTopbar from "@/components/asset/asset-topbar";
import AssetList from "@/components/asset/asset-list";
import { CreateAssetSheet } from "@/components/asset/create-asset-sheet";
import { AdjustBalanceSheet } from "@/components/asset/adjust-balance-sheet";
import { EditAccountSheet } from "@/components/akun/edit-account-sheet";
import type { AssetAccount } from "@/types";

export default function AsetPage() {
	const t = useTranslations("assetPage");

	// Create asset sheet
	const [createSheetOpen, setCreateSheetOpen] = useState(false);

	// Adjust balance sheet
	const [adjustSheetOpen, setAdjustSheetOpen] = useState(false);
	const [selectedAssetForAdjust, setSelectedAssetForAdjust] =
		useState<AssetAccount | null>(null);

	// Rename sheet
	const [renameSheetOpen, setRenameSheetOpen] = useState(false);
	const [selectedAssetForRename, setSelectedAssetForRename] =
		useState<AssetAccount | null>(null);

	function handleAdjustBalance(asset: AssetAccount) {
		setSelectedAssetForAdjust(asset);
		setAdjustSheetOpen(true);
	}

	function handleRename(asset: AssetAccount) {
		setSelectedAssetForRename(asset);
		setRenameSheetOpen(true);
	}

	return (
		<>
			<AssetTopbar />
			<AssetList
				onAdjustBalance={handleAdjustBalance}
				onRename={handleRename}
			/>

			{/* FAB - Tambah Aset */}
			<button
				onClick={() => setCreateSheetOpen(true)}
				className="fixed bottom-22 left-1/2 -translate-x-1/2 z-30
					flex items-center gap-2 px-6 py-3.5 rounded-2xl
					bg-primary-500 text-white font-semibold text-[14px]
					shadow-[0_4px_20px_rgba(59,79,212,0.4)]
					active:scale-95 transition-transform no-tap"
			>
				<Plus size={18} strokeWidth={2.5} />
				{t("create.fab")}
			</button>

			{/* Create Asset Sheet */}
			<CreateAssetSheet
				open={createSheetOpen}
				onOpenChange={setCreateSheetOpen}
			/>

			{/* Adjust Balance Sheet */}
			<AdjustBalanceSheet
				key={selectedAssetForAdjust?.id ?? "closed-adjust-balance"}
				open={adjustSheetOpen}
				onOpenChange={setAdjustSheetOpen}
				asset={selectedAssetForAdjust}
			/>

			{/* Rename Sheet - reused from akun */}
			<EditAccountSheet
				key={selectedAssetForRename?.id ?? "closed-edit-account"}
				open={renameSheetOpen}
				onOpenChange={(v) => {
					setRenameSheetOpen(v);
					if (!v) setSelectedAssetForRename(null);
				}}
				account={
					selectedAssetForRename
						? {
								id: selectedAssetForRename.id,
								code: selectedAssetForRename.code,
								name: selectedAssetForRename.name,
								type: "ASSET" as const,
								level: 3,
								isPosting: true,
								parentId: null,
								createdAt: "",
								updatedAt: "",
								children: [],
							}
						: null
				}
			/>
		</>
	);
}
