/**
 * Group a list by its `type` field, preserving first-seen order.
 *
 * Presentation-only helper for dropdown `<optgroup>` sections. This is not
 * business logic: no aggregation, pricing, or validation happens here — the
 * order and labels come straight from the API payload.
 */
export function groupByType<T extends { type: string }>(items: T[]): Array<[string, T[]]> {
	const map = new Map<string, T[]>();
	for (const item of items) {
		const list = map.get(item.type);
		if (list) {
			list.push(item);
		} else {
			map.set(item.type, [item]);
		}
	}
	return [...map.entries()];
}
