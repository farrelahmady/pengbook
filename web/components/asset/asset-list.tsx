"use client";

import { assetService } from "@/services/asset";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { AssetGroupCard } from "./asset-group-card";
import { AssetGroupCardSkeleton } from "./asset-group-card-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { FolderOpen } from "lucide-react";

export default function AssetList() {
	const t = useTranslations("assetPage");
	const { data, isLoading } = useQuery({
		queryKey: ["assetSummary"],
		queryFn: () => assetService.getSummary(),
	});

	return (
		<div className="px-1">
			<p className="section-label">{t("sectionTitle")}</p>

			{isLoading && (
				<div className="flex flex-col gap-2 px-3 pb-4">
					{Array.from({ length: 3 }).map((_, i) => (
						<AssetGroupCardSkeleton key={i} />
					))}
				</div>
			)}

			{!isLoading && (!data || data.groups.length === 0) && (
				<EmptyState
					icon={FolderOpen}
					title={t("emptyTitle")}
					description={t("emptyDescription")}
				/>
			)}

			{!isLoading && data && (
				<div className="flex flex-col gap-2 px-3 pb-4">
					{data.groups.map((group) => (
						<AssetGroupCard key={group.id} group={group} />
					))}
				</div>
			)}
		</div>
	);
}
