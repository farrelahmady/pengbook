import JournalList from "./journal-list";
import CreateJournal from "./create-journal";
import JournalTopbar from "./journal-topbar";

export default async function JurnalPage() {
	return (
		<>
			<JournalTopbar />
			<JournalList />
			<CreateJournal />
		</>
	);
}
