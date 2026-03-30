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
	return `${import.meta.env.BASE_URL}${href.replace(/^\//, "")}`;
}

// Navigation styling utilities
export function getTocLinkClasses(level: number): string {
	const base = "block transition-colors duration-150";
	if (level === 2)
		return `${base} py-1.5 text-sm text-neutral-500 hover:text-orange-400`;
	if (level === 3)
		return `${base} py-1 text-sm text-neutral-600 hover:text-orange-400`;
	if (level === 4)
		return `${base} py-1 text-xs text-neutral-600 hover:text-orange-400`;
	return `${base} py-0.5 text-xs text-neutral-700 hover:text-orange-400`;
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
