/**
 * Documentation pages — all content defined as type-safe objects.
 */

import type { PageDefinition } from "../types/content.js";

export const documentationPages: PageDefinition[] = [
	// Docs landing
	{
		meta: {
			title: "Documentation",
			menuTitle: "Documentation",
			description:
				"Documentation for moley — automate Cloudflare Tunnel creation, DNS routing, and cleanup from the CLI.",
			order: 1,
			category: "Getting Started",
			href: `/docs/`,
		},
		content: {
			type: "page",
			children: [
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Automate ",
						},
						{
							type: "link",
							href: "https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/",
							text: "Cloudflare Tunnels",
							external: true,
							rel: "nofollow noopener noreferrer",
						},
						{
							type: "text",
							text: " from the command line. Create tunnels, configure DNS, expose local services on your own domain, and clean everything up when you're done.",
						},
					],
				},
				{
					type: "heading",
					level: 2,
					text: "Pages",
				},
				{
					type: "grid",
					columns: 2,
					gap: "medium",
					children: [
						{
							type: "card",
							title: "Installation",
							description: "Install via Homebrew or Go.",
							href: `/docs/installation/`,
							icon: "download",
						},
						{
							type: "card",
							title: "Quick Start",
							description: "Set up your first tunnel in a few minutes.",
							href: `/docs/quick-start/`,
							icon: "zap",
						},
						{
							type: "card",
							title: "Configuration",
							description: "moley.yml reference and environment variables.",
							href: `/docs/configuration/`,
							icon: "code-2",
						},
						{
							type: "card",
							title: "Troubleshooting",
							description: "Common issues and how to fix them.",
							href: `/docs/troubleshooting/`,
							icon: "alert-triangle",
						},
					],
				},
			],
		},
	},

	// Installation
	{
		meta: {
			title: "Installation",
			menuTitle: "Installation",
			slug: "installation",
			description:
				"Install moley using Homebrew or Go. Requires cloudflared and a Cloudflare domain.",
			order: 2,
			category: "Getting Started",
			href: "/docs/installation/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "heading",
					level: 2,
					text: "Prerequisites",
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							children: [
								{
									type: "link",
									href: "https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/",
									text: "cloudflared",
									external: true,
									rel: "nofollow noopener noreferrer",
								},
								{
									type: "text",
									text: " — installed and authenticated",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "text",
									text: "A domain on ",
								},
								{
									type: "link",
									href: "https://www.cloudflare.com/",
									text: "Cloudflare",
									external: true,
									rel: "nofollow noopener noreferrer",
								},
								{
									type: "text",
									text: " with DNS managed by Cloudflare",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Install",
				},
				{
					type: "tabs",
					children: [
						{
							type: "tab",
							title: "Homebrew",
							children: [
								{
									type: "codeblock",
									language: "bash",
									code: "brew install --cask stupside/tap/moley",
								},
							],
						},
						{
							type: "tab",
							title: "Go",
							children: [
								{
									type: "paragraph",
									text: "Requires Go 1.23+.",
								},
								{
									type: "codeblock",
									language: "bash",
									code: "go install github.com/stupside/moley@latest",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Verify",
				},
				{
					type: "codeblock",
					language: "bash",
					code: "moley info",
				},

				{
					type: "callout",
					style: "info",
					children: [
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "Next: ",
								},
								{
									type: "link",
									href: `/docs/quick-start/`,
									text: "Quick Start",
								},
								{
									type: "text",
									text: " — set up your first tunnel.",
								},
							],
						},
					],
				},
			],
		},
	},

	// Quick Start
	{
		meta: {
			title: "Quick Start",
			menuTitle: "Quick Start",
			slug: "quick-start",
			description: "Authenticate, configure, and run a tunnel in minutes.",
			order: 3,
			category: "Getting Started",
			href: "/docs/quick-start/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "step",
					number: 1,
					title: "Authenticate",
					description: "Log in to Cloudflare and set your API token.",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: `cloudflared tunnel login

moley config set --cloudflare.token="your-api-token"`,
						},
						{
							type: "infobox",
							style: "tip",
							title: "Getting an API token",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "Go to ",
										},
										{
											type: "link",
											href: "https://dash.cloudflare.com/profile/api-tokens",
											text: "Cloudflare → API Tokens",
											external: true,
											rel: "nofollow noopener noreferrer",
										},
										{
											type: "text",
											text: " and create a token with:",
										},
									],
								},
								{
									type: "list",
									style: "unordered",
									children: [
										{
											type: "listitem",
											text: "Zone:Zone:Read",
										},
										{
											type: "listitem",
											text: "Zone:DNS:Edit",
										},
									],
								},
							],
						},
					],
				},
				{
					type: "step",
					number: 2,
					title: "Initialize",
					description: "Generate a tunnel configuration file.",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: "moley tunnel init",
						},
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "This creates a ",
								},
								{
									type: "inline-code",
									code: "moley.yml",
								},
								{
									type: "text",
									text: " with a tunnel ID and example app config.",
								},
							],
						},
					],
				},
				{
					type: "step",
					number: 3,
					title: "Configure apps",
					description: "Edit moley.yml to match your local services.",
					children: [
						{
							type: "codeblock",
							language: "yaml",
							title: "moley.yml",
							code: `tunnel:
  name: "my-project"
  persistent: false           # delete tunnel on stop

ingress:
  zone: "yourdomain.com"
  mode: subdomain             # or "wildcard"
  apps:
    - target:
        port: 8080
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "api"      # api.yourdomain.com

    - target:
        port: 3000
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "web"      # web.yourdomain.com`,
						},
					],
				},
				{
					type: "step",
					number: 4,
					title: "Run",
					children: [
						{
							type: "tabs",
							children: [
								{
									type: "tab",
									title: "Foreground",
									children: [
										{
											type: "codeblock",
											language: "bash",
											code: "moley tunnel run",
										},
									],
								},
								{
									type: "tab",
									title: "Background",
									children: [
										{
											type: "codeblock",
											language: "bash",
											code: "moley tunnel run --detach",
										},
									],
								},
							],
						},
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "Your apps are now live. Stop with ",
								},
								{
									type: "inline-code",
									code: "moley tunnel stop",
								},
								{
									type: "text",
									text: " — tunnels and DNS records are cleaned up automatically.",
								},
							],
						},
					],
				},
			],
		},
	},

	// Configuration
	{
		meta: {
			title: "Configuration",
			menuTitle: "Configuration",
			slug: "configuration",
			description:
				"moley.yml reference — tunnel settings, ingress zones, app routing, and environment variable overrides.",
			order: 4,
			category: "Guides",
			href: "/docs/configuration/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "heading",
					level: 2,
					text: "Configuration File",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "All tunnel config lives in ",
						},
						{
							type: "inline-code",
							code: "moley.yml",
						},
						{
							type: "text",
							text: " in your project root. Global settings (API token) are in ",
						},
						{
							type: "inline-code",
							code: "~/.moley/config.yml",
						},
						{
							type: "text",
							text: ".",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "moley.yml",
					code: `tunnel:
  name: "my-project"
  persistent: false

ingress:
  zone: "mydomain.com"
  mode: subdomain
  apps:
    - target:
        port: 3000
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "app"       # app.mydomain.com

    - target:
        port: 8080
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "api"       # api.mydomain.com`,
				},

				{
					type: "heading",
					level: 2,
					text: "Tunnel",
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "name",
								},
								{
									type: "text",
									text: " — tunnel name (required). The actual Cloudflare tunnel is created as moley-{name}.",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "persistent",
								},
								{
									type: "text",
									text: " — if true, the tunnel is kept alive when you stop. Defaults to false (tunnel + DNS cleaned up on stop).",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Ingress Mode",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Controls how DNS records are created. Two modes:",
						},
					],
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "subdomain",
								},
								{
									type: "text",
									text: " — creates one DNS record per app (api.domain.com, web.domain.com). Best for production.",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "wildcard",
								},
								{
									type: "text",
									text: " — creates a single *.domain.com record. Cloudflared routes by hostname. Best for dev when apps change frequently.",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Target Protocol",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Each app target requires a ",
						},
						{
							type: "inline-code",
							code: "protocol",
						},
						{
							type: "text",
							text: " field:",
						},
					],
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "http",
								},
								{
									type: "text",
									text: " — standard HTTP (most common)",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "https",
								},
								{
									type: "text",
									text: " — local service uses HTTPS",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "tcp",
								},
								{
									type: "text",
									text: " — raw TCP (databases, custom protocols)",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Global Config",
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "~/.moley/config.yml",
					code: `cloudflare:
  token: "your-api-token"`,
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Or use an environment variable",
					code: `export MOLEY_CLOUDFLARE_TOKEN="your-api-token"`,
				},

				{
					type: "heading",
					level: 2,
					text: "Zone",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "The ",
						},
						{
							type: "inline-code",
							code: "zone",
						},
						{
							type: "text",
							text: " is the Cloudflare-managed domain used for all subdomains.",
						},
					],
				},
				{
					type: "infobox",
					style: "warning",
					title: "Requirements",
					children: [
						{
							type: "list",
							style: "unordered",
							children: [
								{
									type: "listitem",
									text: "Domain must be on Cloudflare",
								},
								{
									type: "listitem",
									text: "DNS managed by Cloudflare (orange cloud enabled)",
								},
								{
									type: "listitem",
									text: "API token needs Zone:Read and DNS:Edit permissions",
								},
							],
						},
					],
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Zone environment variable",
					code: `export MOLEY_TUNNEL_INGRESS_ZONE="yourdomain.com"`,
				},

				{
					type: "heading",
					level: 2,
					text: "Apps",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Each app maps a local ",
						},
						{
							type: "inline-code",
							code: "target",
						},
						{
							type: "text",
							text: " (port + hostname) to a public ",
						},
						{
							type: "inline-code",
							code: "expose",
						},
						{
							type: "text",
							text: " (subdomain).",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "Full example",
					code: `ingress:
  zone: "yourdomain.com"
  mode: subdomain
  apps:
    - target:
        port: 3000
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "app"       # app.yourdomain.com

    - target:
        port: 8080
        hostname: "localhost"
        protocol: http
      expose:
        subdomain: "api"       # api.yourdomain.com

    - target:
        port: 5432
        hostname: "127.0.0.1"
        protocol: tcp
      expose:
        subdomain: "db"        # db.yourdomain.com (TCP)`,
				},
				{
					type: "heading",
					level: 3,
					text: "Environment variables",
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Per-app overrides",
					code: `export MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_PORT="3000"
export MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_HOSTNAME="localhost"
export MOLEY_TUNNEL_INGRESS_APPS_0_EXPOSE_SUBDOMAIN="app"`,
				},
			],
		},
	},

	// Troubleshooting
	{
		meta: {
			title: "Troubleshooting",
			menuTitle: "Troubleshooting",
			slug: "troubleshooting",
			description:
				"Fix DNS issues, auth errors, tunnel failures, and orphaned resources.",
			order: 5,
			category: "Guides",
			href: "/docs/troubleshooting/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "heading",
					level: 2,
					text: "Quick fixes",
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							children: [
								{
									type: "text",
									text: "Tunnel already exists → ",
								},
								{
									type: "inline-code",
									code: "moley tunnel stop",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "text",
									text: "Permission denied → ",
								},
								{
									type: "inline-code",
									code: "cloudflared tunnel login",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "text",
									text: "DNS not resolving → wait 5 minutes, then check the ",
								},
								{
									type: "link",
									href: "https://dash.cloudflare.com/",
									text: "Cloudflare dashboard",
									external: true,
									rel: "nofollow noopener noreferrer",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "DNS issues",
				},
				{
					type: "step",
					number: 1,
					title: "Check DNS resolution",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: "dig your-domain.com\nnslookup your-domain.com 1.1.1.1",
						},
					],
				},
				{
					type: "step",
					number: 2,
					title: "Test access",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: "curl -v https://your-domain.com",
						},
					],
				},
				{
					type: "step",
					number: 3,
					title: "Check Cloudflare dashboard",
					children: [
						{
							type: "list",
							style: "unordered",
							children: [
								{
									type: "listitem",
									text: "DNS record exists and points to your tunnel",
								},
								{
									type: "listitem",
									text: "Proxy status (orange cloud) is enabled",
								},
								{
									type: "listitem",
									text: "Wait up to 5 minutes for propagation",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Tunnel failures",
				},
				{
					type: "step",
					number: 1,
					title: "Check your API token",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: `cat ~/.moley/config.yml
moley config set --cloudflare.token=your_token`,
						},
					],
				},
				{
					type: "step",
					number: 2,
					title: "Check cloudflared auth",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: `cloudflared tunnel login
cloudflared tunnel list`,
						},
					],
				},
				{
					type: "step",
					number: 3,
					title: "Run with debug logs",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: `moley tunnel run --log-level=debug
moley tunnel run --dry-run`,
						},
					],
				},

				{
					type: "heading",
					level: 2,
					text: "Orphaned resources",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "If ",
						},
						{
							type: "inline-code",
							code: "moley tunnel stop",
						},
						{
							type: "text",
							text: " fails or resources remain after a crash:",
						},
					],
				},
				{
					type: "codeblock",
					language: "bash",
					code: `# Force orphan detection
rm -f moley.lock
moley tunnel stop --log-level=debug

# Manual cleanup (last resort)
cloudflared tunnel list
cloudflared tunnel delete <tunnel-id>`,
				},
			],
		},
	},

	// Advanced Debugging (internal)
	{
		meta: {
			title: "Advanced Debugging",
			menuTitle: "Advanced Debugging",
			slug: "advanced-debugging",
			description: "Trace logging, config inspection, and resource debugging.",
			order: 10,
			category: "Internal",
			internal: true,
			href: "/docs/advanced-debugging/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "paragraph",
					text: "Deep debugging for contributors and power users. Use these when basic troubleshooting doesn't cut it.",
				},
				{
					type: "heading",
					level: 2,
					text: "Trace logging",
				},
				{
					type: "codeblock",
					language: "bash",
					code: `moley tunnel run --log-level=trace 2>&1 | tee debug.log`,
				},
				{
					type: "heading",
					level: 2,
					text: "Inspect state",
				},
				{
					type: "codeblock",
					language: "bash",
					code: `cat ~/.moley/config.yml
cat moley.yml
cat moley.lock
cloudflared tunnel list`,
				},
				{
					type: "heading",
					level: 2,
					text: "Resource debugging",
				},
				{
					type: "codeblock",
					language: "bash",
					code: `moley tunnel stop --dry-run --log-level=debug

# Force orphan detection
rm moley.lock
moley tunnel stop --dry-run --log-level=trace`,
				},
			],
		},
	},
];
