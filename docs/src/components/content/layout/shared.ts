// Shared utilities for layout components
export const gapClasses = {
	none: "gap-0",
	small: "gap-4",
	medium: "gap-6",
	large: "gap-8",
} as const;

export type GapSize = keyof typeof gapClasses;