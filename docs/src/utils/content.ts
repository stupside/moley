import type { ContentBlock } from "../types/content.js";

export function createSlug(text: string): string {
	return text
		.toLowerCase()
		.replace(/[^\w\s-]/g, "")
		.replace(/\s+/g, "-");
}

export function traverseContent<T>(
	content: ContentBlock,
	predicate: (block: ContentBlock) => T | null,
): T[] {
	const results: T[] = [];

	function traverse(block: ContentBlock) {
		const result = predicate(block);
		if (result !== null) {
			results.push(result);
		}

		if (block.children) {
			block.children.forEach(traverse);
		}
	}

	traverse(content);
	return results;
}

export interface TocItem {
	id: string;
	title: string;
	level: number;
}

export function extractTocItems(content: ContentBlock): TocItem[] {
	return traverseContent(content, (block): TocItem | null => {
		switch (block.type) {
			case "heading":
				return {
					id: block.id || createSlug(block.text),
					title: block.text,
					level: block.level,
				};
			case "section":
				return {
					id: block.id || createSlug(block.title || ""),
					title: block.title || "",
					level: 2,
				};
			default:
				return null;
		}
	});
}
