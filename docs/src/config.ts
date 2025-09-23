/**
 * Application Configuration
 *
 * Modular, environment-aware configuration system.
 * Each section handles a specific domain of configuration.
 */

// Site branding configuration
export const site = {
	name: "Moley",
	logo: "moley.png",
	repo: {
		url: "https://github.com/stupside/moley",
		issues: () => `${site.repo.url}/issues`,
		releases: () => `${site.repo.url}/releases`,
	},
	author: { name: "Kilian Houpeurt", twitter: "@kilianhprt" },
	tagline: "Expose localhost with Cloudflare Tunnels",
} as const;

// Navigation configuration
export const navigation = {
	// Main navigation items
	items: [
		{ label: "Home", href: "/" },
		{ label: "Docs", href: "/docs/" },
	] as const,

	// Generate navigation items
	getItems() {
		return this.items;
	},
} as const;

// Main configuration object
export const config = {
	site,
	navigation,
} as const;

// Type exports
export type Config = typeof config;