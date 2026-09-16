import type { ReactNode } from "react";
import type { AccountWithChildren } from "@/types";

export function filterAccountTree(
	accounts: AccountWithChildren[],
	query: string,
): { filtered: AccountWithChildren[]; matchCount: number } {
	const q = query.trim().toLowerCase();
	if (q === "") return { filtered: accounts, matchCount: 0 };

	let matchCount = 0;

	function recurse(nodes: AccountWithChildren[]): AccountWithChildren[] {
		const out: AccountWithChildren[] = [];
		for (const node of nodes) {
			const selfMatch =
				node.code.toLowerCase().includes(q) ||
				node.name.toLowerCase().includes(q);
			if (selfMatch) matchCount += 1;
			const filteredChildren = recurse(node.children ?? []);
			if (selfMatch || filteredChildren.length > 0) {
				out.push({ ...node, children: filteredChildren });
			}
		}
		return out;
	}

	return { filtered: recurse(accounts), matchCount };
}

export function highlightMatch(text: string, query: string): ReactNode {
	const q = query.trim();
	if (q === "") return text;

	const lowerText = text.toLowerCase();
	const lowerQ = q.toLowerCase();
	const parts: ReactNode[] = [];
	let cursor = 0;
	let key = 0;

	for (;;) {
		const idx = lowerText.indexOf(lowerQ, cursor);
		if (idx === -1) break;
		if (idx > cursor) parts.push(text.slice(cursor, idx));
		parts.push(
			<mark
				key={key++}
				className="bg-warning-100 text-inherit rounded-sm px-px"
			>
				{text.slice(idx, idx + q.length)}
			</mark>,
		);
		cursor = idx + q.length;
	}
	if (cursor < text.length) parts.push(text.slice(cursor));
	return parts;
}
