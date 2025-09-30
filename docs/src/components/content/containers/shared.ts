// Shared utilities for container components (InfoBox, Callout)
// Using Cloudflare brand colors (orange-based theme)
export const containerStyles = {
	info: "bg-orange-50 border-orange-200 ring-1 ring-orange-600/10",
	warning: "bg-amber-50 border-amber-200 ring-1 ring-amber-600/10",
	error: "bg-red-50 border-red-200 ring-1 ring-red-600/10",
	success: "bg-orange-25 border-orange-150 ring-1 ring-orange-500/10",
	tip: "bg-orange-50 border-orange-200 ring-1 ring-orange-600/10",
	note: "bg-gray-50 border-gray-200 ring-1 ring-gray-600/10",
} as const;

export const containerTitleStyles = {
	info: "text-orange-900",
	warning: "text-amber-900",
	error: "text-red-900",
	success: "text-orange-800",
	tip: "text-orange-900",
	note: "text-gray-900",
} as const;

export type ContainerStyle = keyof typeof containerStyles;