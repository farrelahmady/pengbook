import CreateJournal from "./create-journal";
import JurnalPageContent from "./jurnal-page-content";

export default async function JurnalPage() {
	return (
		<>
			<JurnalPageContent />
			<CreateJournal />
		</>
	);
}
