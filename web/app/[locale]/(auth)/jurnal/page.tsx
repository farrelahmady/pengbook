import CreateJournal from "@/components/journal/create-journal";
import JurnalPageContent from "@/components/journal/jurnal-page-content";

export default async function JurnalPage() {
	return (
		<>
			<JurnalPageContent />
			<CreateJournal />
		</>
	);
}
