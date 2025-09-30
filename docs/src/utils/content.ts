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

// URL utility function
export function buildUrl(href: string): string {
	return `${import.meta.env.BASE_URL}${href.replace(/^\//, '')}`;
}

// Navigation styling utilities
export function getTocLinkClasses(level: number): string {
	const base = "block transition-colors";
	if (level === 2) return `${base} py-2 px-3 text-gray-700 font-medium hover:text-orange-700 hover:bg-orange-50 rounded-lg`;
	if (level === 3) return `${base} py-1.5 px-2 text-gray-600 text-sm hover:text-orange-600 hover:bg-orange-50 rounded-md`;
	if (level === 4) return `${base} py-1 px-2 text-gray-500 text-xs hover:text-orange-500 hover:bg-orange-25 rounded`;
	return `${base} py-1 px-1 text-gray-400 text-xs hover:text-orange-400 rounded`;
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
