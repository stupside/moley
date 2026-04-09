/**
 * Icon component constants and configuration
 */

import {
	ChevronRight,
	CloudDownload,
	CodeXml,
	Download,
	Github,
	Link,
	Lock,
	Shield,
	TriangleAlert,
	Zap,
} from "@lucide/astro";

import type { BaseComponentProps } from "../../../types/shared";

export const iconMap = {
	download: Download,
	zap: Zap,
	"code-2": CodeXml,
	"alert-triangle": TriangleAlert,
	"chevron-right": ChevronRight,
	link: Link,
	github: Github,
	"cloud-download": CloudDownload,
	lock: Lock,
	shield: Shield,
} as const;

export type IconName = keyof typeof iconMap;

export type IconSize = "xs" | "sm" | "md" | "lg" | "xl";

export const sizeStyles = {
	xs: "w-3 h-3",
	sm: "w-4 h-4",
	md: "w-5 h-5",
	lg: "w-6 h-6",
	xl: "w-8 h-8",
} as const;

export interface IconProps extends BaseComponentProps {
	name: IconName;
	size?: IconSize;
}
