// Container style system — dark theme with orange accents
export const containerStyles = {
	info: "bg-neutral-900 border-orange-500/40",
	warning: "bg-neutral-900 border-amber-500/40",
	error: "bg-neutral-900 border-red-500/40",
	success: "bg-neutral-900 border-emerald-500/40",
	tip: "bg-neutral-900 border-orange-500/40",
	note: "bg-neutral-900 border-neutral-600",
} as const;

export const containerTitleStyles = {
	info: "text-orange-400",
	warning: "text-amber-400",
	error: "text-red-400",
	success: "text-emerald-400",
	tip: "text-orange-400",
	note: "text-neutral-400",
} as const;

export type ContainerStyle = keyof typeof containerStyles;
