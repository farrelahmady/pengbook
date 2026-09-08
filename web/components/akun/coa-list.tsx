"use client";

import { coaService } from "@/services/coa";
import { useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { CoaTypeGroupCard } from "./coa-type-group";
import { CoaTypeGroupSkeleton } from "./coa-skeleton";
import { EmptyState } from "@/components/shared/empty-state";
import { BookOpen } from "lucide-react";

export default function CoaList() {
	const t = useTranslations("coaPage");
	const { data, isLoading } = useQuery({
		queryKey: ["coaSummary"],
		queryFn: () => coaService.getSummary(),
	});

	return (
		<div className="px-1">
			<p className="section-label">{t("sectionTitle")}</p>

			<div className="flex flex-col gap-2 px-3 pb-4">
				{isLoading && (
					<>
						<CoaTypeGroupSkeleton />
						<CoaTypeGroupSkeleton />
						<CoaTypeGroupSkeleton />
					</>
				)}

				{!isLoading && (!data || data.groups.length === 0) && (
					<EmptyState
						icon={BookOpen}
						title={t("emptyTitle")}
						description={t("emptyDescription")}
					/>
				)}

				{!isLoading && data && (
					<>
						{data.groups.map((group) => (
							<CoaTypeGroupCard key={group.type} group={group} />
						))}
					</>
				)}
			</div>
		</div>
	);
}
