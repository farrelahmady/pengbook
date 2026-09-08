import { dummyJournals } from "@/lib/dummy-data";
import { JournalEntry, JournalSummary, CreateJournalDto } from "@/types";

export const journalService = {
	getAllScrollView: async (request: {
		page: number;
		limit: number;
		startDate?: Date;
		endDate?: Date;
		accountIds?: string[];
	}): Promise<JournalEntry[]> => {
		// Delay to simulate network latency
		await new Promise((resolve) => setTimeout(resolve, 1500));

		let filtered = [...dummyJournals];

		// Filter by date range
		if (request.startDate) {
			const start = new Date(request.startDate);
			start.setHours(0, 0, 0, 0);
			filtered = filtered.filter((j) => new Date(j.date) >= start);
		}
		if (request.endDate) {
			const end = new Date(request.endDate);
			end.setHours(23, 59, 59, 999);
			filtered = filtered.filter((j) => new Date(j.date) <= end);
		}

		// Filter by account IDs (journal must contain at least one of the selected accounts)
		if (request.accountIds && request.accountIds.length > 0) {
			filtered = filtered.filter((j) =>
				j.lines.some((line) => request.accountIds!.includes(line.accountId)),
			);
		}

		const start = request.page * request.limit - request.limit;
		return Promise.resolve(filtered.slice(start, start + request.limit));
	},

	getTotalSummary: async (): Promise<JournalSummary> => {
		// Dummy data, replace with actual API call
		await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate network delay
		return {
			totalDebit: dummyJournals.reduce(
				(sum, j) => sum + j.lines.reduce((s, l) => s + parseFloat(l.debit), 0),
				0,
			),
			totalCredit: dummyJournals.reduce(
				(sum, j) => sum + j.lines.reduce((s, l) => s + parseFloat(l.credit), 0),
				0,
			),
			transactionCount: dummyJournals.length,
		};
	},

	create: async (dto: CreateJournalDto): Promise<JournalEntry> => {
		// Simulate API call
		await new Promise((resolve) => setTimeout(resolve, 1500));

		// Simulate random failure (10% chance)
		if (Math.random() < 0.1) {
			throw new Error("Gagal menyimpan jurnal. Silakan coba lagi.");
		}

		const newEntry: JournalEntry = {
			id: `j${Date.now()}`,
			date: dto.date,
			description: dto.description ?? null,
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString(),
			lines: dto.lines.map((line, idx) => ({
				id: `l${Date.now()}_${idx}`,
				journalEntryId: `j${Date.now()}`,
				accountId: line.accountId,
				debit: String(line.debit),
				credit: String(line.credit),
			})),
		};

		// Add to dummy data (simulating persistence)
		dummyJournals.unshift(newEntry);

		return Promise.resolve(newEntry);
	},

	update: async (
		id: string,
		dto: CreateJournalDto,
	): Promise<JournalEntry> => {
		// Simulate API call
		await new Promise((resolve) => setTimeout(resolve, 1500));

		// Simulate random failure (10% chance)
		if (Math.random() < 0.1) {
			throw new Error("Gagal memperbarui jurnal. Silakan coba lagi.");
		}

		const index = dummyJournals.findIndex((j) => j.id === id);
		if (index === -1) {
			throw new Error("Jurnal tidak ditemukan.");
		}

		const updatedEntry: JournalEntry = {
			...dummyJournals[index],
			date: dto.date,
			description: dto.description ?? null,
			updatedAt: new Date().toISOString(),
			lines: dto.lines.map((line, idx) => ({
				id: `l${Date.now()}_${idx}`,
				journalEntryId: id,
				accountId: line.accountId,
				debit: String(line.debit),
				credit: String(line.credit),
			})),
		};

		// Update dummy data (simulating persistence)
		dummyJournals[index] = updatedEntry;

		return Promise.resolve(updatedEntry);
	},

	createBulk: async (file: File): Promise<{ count: number }> => {
		// Simulate file upload and processing
		await new Promise((resolve) => setTimeout(resolve, 2000));

		// Simulate random failure (10% chance)
		if (Math.random() < 0.1) {
			throw new Error("Gagal memproses file. Periksa format file Anda.");
		}

		// Simulate processing 5 entries from file
		return Promise.resolve({ count: 5 });
	},

	downloadTransactions: async (): Promise<void> => {
		// Simulate API call
		await new Promise((resolve) => setTimeout(resolve, 1500));

		// Generate CSV from dummy data
		const headers = ["Tanggal", "Deskripsi", "Akun", "Debit", "Kredit"];
		const rows: string[][] = [];

		for (const journal of dummyJournals) {
			for (const line of journal.lines) {
				rows.push([
					journal.date.slice(0, 10),
					journal.description ?? "",
					`${line.account?.code ?? ""} · ${line.account?.name ?? ""}`,
					line.debit !== "0" ? line.debit : "",
					line.credit !== "0" ? line.credit : "",
				]);
			}
		}

		const csvContent = [headers, ...rows]
			.map((row) => row.map((cell) => `"${cell}"`).join(","))
			.join("\n");

		const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `transaksi-${new Date().toISOString().slice(0, 10)}.csv`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	},
};
