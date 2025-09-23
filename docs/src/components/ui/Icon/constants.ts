/**
 * Icon component constants and configuration
 */

import {
	Check,
	ChevronRight,
	CloudDownload,
	CodeXml,
	Download,
	Github,
	Link,
	SquarePen,
	TriangleAlert,
	X,
	Zap,
} from "@lucide/astro";

import type { BaseComponentProps } from "../../../types/shared";

export const iconMap = {
	check: Check,
	x: X,
	download: Download,
	zap: Zap,
	"code-2": CodeXml,
	"alert-triangle": TriangleAlert,
	"chevron-right": ChevronRight,
	link: Link,
	github: Github,
	"cloud-download": CloudDownload,
	edit: SquarePen,
} as const;

export type IconName = keyof typeof iconMap;

export type IconSize = "xs" | "sm" | "md" | "lg" | "xl";

export interface IconProps extends BaseComponentProps {
	name: IconName;
	size?: IconSize;
}
