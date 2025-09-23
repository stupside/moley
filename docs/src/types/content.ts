/**
 * Recursive Content System Types
 *
 * This system defines a type-safe content block architecture that allows
 * any content block to contain children, enabling complex nested documentation
 * structures while maintaining strict TypeScript typing.
 *
 * Key features:
 * - Recursive content blocks with type safety
 * - Consistent interface patterns extending BaseContentBlock
 * - Union types for compile-time validation
 * - Flexible styling and layout options
 */

import type { CodeLanguage } from "astro";
import type { IconName } from "../components/ui/Icon/constants";

// Code language types

// Spacing and layout types
export type SpacingSize = "none" | "small" | "medium" | "large";
export type GapSize = "none" | "small" | "medium" | "large";
export type InfoBoxStyle = "info" | "warning" | "error" | "tip" | "note";
export type CalloutStyle = "default" | "info" | "highlight" | "box";
export type ListStyle = "ordered" | "unordered";
export type HeadingLevel = 1 | 2 | 3 | 4 | 5 | 6;

/**
 * Base interface for all content blocks
 *
 * All content blocks extend this interface to ensure consistent structure
 * and enable recursive nesting of any content type.
 */
export interface BaseContentBlock<T extends string> {
	/** Optional unique identifier */
	id?: string;
	/** The type identifier for this content block */
	type: T;
	/** Optional array of child content blocks for recursive nesting */
	children?: ContentBlock[];
	/** Optional CSS class names for styling */
	className?: string;
}

// Text and inline content blocks
export interface TextBlock extends BaseContentBlock<"text"> {
	text: string;
}

/**
 * Metadata for documentation pages
 *
 * Contains all the information needed to generate navigation,
 * SEO tags, and page routing.
 */
export interface PageMeta {
	/** Page title for navigation and SEO */
	title: string;
	/** Short title for navigation menu (falls back to title if not provided) */
	menuTitle?: string;
	/** Page description for SEO and previews */
	description: string;
	/** Optional sort order for navigation */
	order?: number;
	/** Optional category grouping for navigation */
	category?: string;
	/** Optional custom slug override (defaults to title-based slug) */
	slug?: string;
	/** Page URL path */
	href: string;
	/** Whether page is internal-only (excluded from public navigation) */
	internal?: boolean;
}

/**
 * Complete page definition
 *
 * Combines page metadata with the actual content structure.
 */
export interface PageDefinition {
	/** Page metadata for navigation and SEO */
	meta: PageMeta;
	/** Root content block (typically a PageBlock) */
	content: ContentBlock;
}

export interface InlineCodeBlock extends BaseContentBlock<"inline-code"> {
	code: string;
}

export interface LinkBlock extends BaseContentBlock<"link"> {
	href: string;
	text: string;
	external?: boolean;
	rel?: string;
}

// Block-level content blocks
export interface HeadingBlock extends BaseContentBlock<"heading"> {
	level: HeadingLevel;
	text: string;
}

export interface ParagraphBlock extends BaseContentBlock<"paragraph"> {
	text?: string;
	children?: ContentBlock[];
}

export interface CodeBlock extends BaseContentBlock<"codeblock"> {
	language: CodeLanguage;
	code: string;
	title?: string;
}

// List structures
export interface ListBlock extends BaseContentBlock<"list"> {
	style: ListStyle;
	children: ListItemBlock[];
}

export interface ListItemBlock extends BaseContentBlock<"listitem"> {
	text?: string;
	children?: ContentBlock[];
}

// Container blocks
export interface SectionBlock extends BaseContentBlock<"section"> {
	title?: string;
	border?: boolean;
	spacing?: SpacingSize;
	children: ContentBlock[];
}

export interface StepBlock extends BaseContentBlock<"step"> {
	number?: number | string;
	title: string;
	description?: string;
	children?: ContentBlock[];
}

export interface InfoBoxBlock extends BaseContentBlock<"infobox"> {
	style?: InfoBoxStyle;
	title?: string;
	children: ContentBlock[];
}

export interface CalloutBlock extends BaseContentBlock<"callout"> {
	style?: CalloutStyle;
	children: ContentBlock[];
}

export interface CardBlock extends BaseContentBlock<"card"> {
	title: string;
	description?: string;
	href: string;
	icon?: IconName;
	external?: boolean;
}

// Layout blocks
export interface GridBlock extends BaseContentBlock<"grid"> {
	columns: number;
	gap?: GapSize;
	children: ContentBlock[];
}

export interface TabsBlock extends BaseContentBlock<"tabs"> {
	children: TabBlock[];
}

export interface TabBlock extends BaseContentBlock<"tab"> {
	title: string;
	children: ContentBlock[];
}

// Page container
export interface PageBlock extends BaseContentBlock<"page"> {
	children: ContentBlock[];
}

/**
 * Union type encompassing all possible content blocks
 *
 * This discriminated union enables TypeScript to provide type safety
 * and autocomplete for all content block types while ensuring only
 * valid combinations are possible.
 */
export type ContentBlock =
	| TextBlock
	| InlineCodeBlock
	| LinkBlock
	| HeadingBlock
	| ParagraphBlock
	| CodeBlock
	| ListBlock
	| ListItemBlock
	| SectionBlock
	| StepBlock
	| InfoBoxBlock
	| CalloutBlock
	| CardBlock
	| GridBlock
	| TabsBlock
	| TabBlock
	| PageBlock;
