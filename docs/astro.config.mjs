import sitemap from "@astrojs/sitemap";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";

export default defineConfig({
	base: "/moley/",
	site: "https://stupside.github.io/moley",
	server: {
		allowedHosts: [".localhost", ".xonery.dev", ".moley.dev"],
	},
	outDir: "./dist",
	integrations: [
		sitemap({
			lastmod: new Date(),
			priority: 0.8,
			changefreq: "weekly",
		}),
	],
	markdown: {
		shikiConfig: {
			themes: {
				dark: "github-dark",
				light: "github-light",
			},
			wrap: true,
			transformers: [],
		},
	},
	build: {
		format: "directory",
		inlineStylesheets: "auto",
	},
	output: "static",
	compressHTML: true,
	vite: {
		plugins: [tailwindcss()],
		build: {
			cssCodeSplit: true,
			rollupOptions: {
				output: {
					manualChunks: {
						vendor: ["astro"],
					},
				},
			},
		},
	},
});
