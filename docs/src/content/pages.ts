/**
 * Centralized Content Definitions
 *
 * All documentation pages are defined here as type-safe content objects.
 * This replaces individual .astro page files with a declarative system.
 */

import type { PageDefinition } from "../types/content.js";

export const documentationPages: PageDefinition[] = [
	// Landing page
	{
		meta: {
			title: "Moley Documentation - Complete Guide to Localhost Tunneling",
			menuTitle: "Documentation",
			description:
				"Complete guide to using Moley for secure localhost tunneling with Cloudflare. Learn installation, configuration, and troubleshooting for your custom domain setup.",
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
							text: "Moley is a powerful CLI tool that simplifies exposing your localhost applications to the internet using ",
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
							text: ". With Moley, you can share your local development servers on your own custom domain without complex port forwarding or expensive tunnel services.",
						},
					],
				},
				{
					type: "heading",
					level: 2,
					text: "What You'll Learn",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "This comprehensive documentation covers everything from basic installation to advanced troubleshooting. Whether you're a developer wanting to share work-in-progress applications or a DevOps engineer setting up secure tunnels, you'll find step-by-step guides and best practices. Learn more about ",
						},
						{
							type: "link",
							href: "https://developers.cloudflare.com/cloudflare-one/",
							text: "Cloudflare Zero Trust",
							external: true,
							rel: "nofollow noopener noreferrer",
						},
						{
							type: "text",
							text: " for enterprise security features.",
						},
					],
				},
				{
					type: "heading",
					level: 2,
					text: "Getting Started",
				},
				{
					type: "grid",
					columns: 2,
					gap: "medium",
					children: [
						{
							type: "card",
							title: "Installation",
							description:
								"Install Moley using Homebrew, Go, or build from source.",
							href: `/docs/installation/`,
							icon: "download",
						},
						{
							type: "card",
							title: "Quick Start",
							description: "Get up and running with Moley in minutes.",
							href: `/docs/quick-start/`,
							icon: "zap",
						},
						{
							type: "card",
							title: "Configuration",
							description: "Advanced configuration options and examples.",
							href: `/docs/configuration/`,
							icon: "code-2",
						},
						{
							type: "card",
							title: "Troubleshooting",
							description: "Common issues and solutions.",
							href: `/docs/troubleshooting/`,
							icon: "alert-triangle",
						},
					],
				},
			],
		},
	},

	// Installation page
	{
		meta: {
			title: "Install Moley - Cloudflare Tunnel Manager Setup Guide",
			menuTitle: "Installation",
			slug: "installation",
			description:
				"Install Moley using Homebrew, Go, or build from source. Quick setup guide for localhost tunneling with Cloudflare on macOS, Linux, and Windows systems.",
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
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Before installing Moley, make sure you have the required dependencies. If you encounter issues during installation, check our ",
						},
						{
							type: "link",
							href: `/docs/troubleshooting/`,
							text: "troubleshooting guide",
						},
						{
							type: "text",
							text: " for solutions.",
						},
					],
				},
				{
					type: "list",
					style: "unordered",
					children: [
						{
							type: "listitem",
							text: "Cloudflared",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "link",
											href: "https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/",
											text: "Install cloudflared",
											external: true,
											rel: "nofollow noopener noreferrer",
										},
										{
											type: "text",
											text: " and authenticate with your account",
										},
									],
								},
							],
						},
						{
							type: "listitem",
							text: "Cloudflare Domain",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "A ",
										},
										{
											type: "link",
											href: "https://www.cloudflare.com/",
											text: "Cloudflare account",
											external: true,
											rel: "nofollow noopener noreferrer",
										},
										{
											type: "text",
											text: " with a custom domain configured",
										},
									],
								},
							],
						},
						{
							type: "listitem",
							text: "Go 1.23+ (Optional)",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "Only required for building from source. Download from ",
										},
										{
											type: "link",
											href: "https://golang.org/dl/",
											text: "golang.org",
											external: true,
											rel: "nofollow noopener noreferrer",
										},
										{
											type: "text",
											text: ".",
										},
									],
								},
							],
						},
					],
				},

				{
					type: "tabs",
					children: [
						{
							type: "tab",
							title: "Homebrew",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "Recommended for macOS users. ",
										},
										{
											type: "link",
											href: "https://brew.sh/",
											text: "Install Homebrew",
											external: true,
											rel: "nofollow noopener noreferrer",
										},
										{
											type: "text",
											text: " first if you don't have it.",
										},
									],
								},
								{
									type: "codeblock",
									language: "bash",
									code: "brew install --cask stupside/tap/moley",
								},
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "This will install the latest stable version of Moley. After installation, proceed to the ",
										},
										{
											type: "link",
											href: `/docs/quick-start/`,
											text: "Quick Start guide",
										},
										{
											type: "text",
											text: " to set up your first tunnel.",
										},
									],
									className: "text-sm text-gray-500 mt-3",
								},
							],
						},
						{
							type: "tab",
							title: "Go Install",
							children: [
								{
									type: "paragraph",
									children: [
										{
											type: "text",
											text: "If you have Go installed, you can install Moley directly. For advanced configuration options after installation, see our ",
										},
										{
											type: "link",
											href: `/docs/configuration/`,
											text: "Configuration guide",
										},
										{
											type: "text",
											text: ".",
										},
									],
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
					text: "Verify Installation",
				},
				{
					type: "codeblock",
					language: "bash",
					code: "# Check if Moley is installed correctly\nmoley version\n\n# View build information\nmoley info",
				},
				{
					type: "codeblock",
					language: "text",
					title: "Expected Output",
					code: "Moley v1.0.0\nBuild: abc1234\nGo version: go1.23.0",
					className: "mt-6",
				},

				{
					type: "callout",
					style: "info",
					children: [
						{
							type: "heading",
							level: 2,
							text: "Next Steps",
						},
						{
							type: "paragraph",
							text: "Ready to start tunneling? Follow these steps",
						},
						{
							type: "list",
							style: "ordered",
							children: [
								{
									type: "listitem",
									children: [
										{
											type: "link",
											href: `/docs/quick-start/`,
											text: "Follow the Quick Start guide",
										},
									],
								},
								{
									type: "listitem",
									children: [
										{
											type: "link",
											href: `/docs/configuration/`,
											text: "Learn about configuration options",
										},
									],
								},
							],
						},
					],
				},
			],
		},
	},

	// Quick Start page
	{
		meta: {
			title: "Moley Quick Start - Expose Localhost in Minutes",
			menuTitle: "Quick Start",
			slug: "quick-start",
			description:
				"Get up and running with Moley in minutes. Step-by-step guide to authenticate, configure, and run Cloudflare tunnels for your localhost applications instantly.",
			order: 3,
			category: "Getting Started",
			href: "/docs/quick-start/",
		},
		content: {
			type: "page",
			children: [
				{
					type: "infobox",
					style: "info",
					title: "Configuration Options",
					children: [
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "You can configure Moley using either configuration files or environment variables. Environment variables take precedence over file configuration and are ideal for CI/CD, containers, or keeping sensitive data secure. See the ",
								},
								{
									type: "link",
									href: `/docs/configuration/`,
									text: "Configuration guide",
								},
								{
									type: "text",
									text: " for complete environment variable documentation.",
								},
							],
						},
					],
				},
				{
					type: "step",
					number: 1,
					title: "Authentication Setup",
					description: "First, authenticate with Cloudflare",
					children: [
						{
							type: "codeblock",
							language: "bash",
							code: '# Authenticate cloudflared with your Cloudflare account\ncloudflared tunnel login\n\n# Set your Cloudflare API token\nmoley config set --cloudflare.token="your-api-token"',
						},
						{
							type: "infobox",
							style: "tip",
							title: "How to get your API token",
							children: [
								{
									type: "list",
									style: "ordered",
									children: [
										{
											type: "listitem",
											children: [
												{
													type: "text",
													text: "Go to ",
												},
												{
													type: "link",
													href: "https://dash.cloudflare.com/profile/api-tokens",
													text: "Cloudflare Dashboard → My Profile → API Tokens",
													external: true,
													rel: "nofollow noopener noreferrer",
												},
											],
										},
										{
											type: "listitem",
											text: "Create a new token with these permissions",
										},
									],
								},
								{
									type: "list",
									style: "unordered",
									className: "ml-4",
									children: [
										{
											type: "listitem",
											children: [
												{
													type: "text",
													text: "Zone:Zone:Read",
													className: "font-bold",
												},
												{
													type: "text",
													text: " (for the domain you want to use)",
												},
											],
										},
										{
											type: "listitem",
											children: [
												{
													type: "text",
													text: "Zone:DNS:Edit",
													className: "font-bold",
												},
												{
													type: "text",
													text: " (for the domain you want to use)",
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
											text: "This creates your global configuration file at ",
										},
										{
											type: "inline-code",
											code: "~/.moley/config.yml",
										},
										{
											type: "text",
											text: ". Environment variables take precedence over file configuration.",
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
					title: "Initialize Tunnel",
					description: "Create a new tunnel configuration",
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
									text: "This automatically generates a ",
								},
								{
									type: "inline-code",
									code: "moley.yml",
								},
								{
									type: "text",
									text: " file with",
								},
							],
						},
						{
							type: "list",
							style: "unordered",
							children: [
								{ type: "listitem", text: "A unique tunnel ID" },
								{
									type: "listitem",
									text: "Example configuration for common ports",
								},
								{ type: "listitem", text: "Your domain setup" },
							],
						},
					],
				},
				{
					type: "step",
					number: 3,
					title: "Configure Applications",
					description:
						"Edit the configuration file to specify which local services to expose",
					children: [
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "Edit the generated ",
								},
								{
									type: "inline-code",
									code: "moley.yml",
								},
								{
									type: "text",
									text: " file to match your local services. For more advanced options, see the ",
								},
								{
									type: "link",
									href: `/docs/configuration/`,
									text: "Configuration guide",
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
							title: "Basic Configuration (moley.yml)",
							code: `tunnel:
  id: "a-unique-tunnel-id"  # Auto-generated

ingress:
  zone: "yourdomain.com"    # Replace with your domain
  apps:
    - target:
        port: 8080          # Your local service port
        hostname: "localhost"
      expose:
        subdomain: "api"    # Will be available at api.yourdomain.com
    - target:
        port: 3000
        hostname: "localhost"
      expose:
        subdomain: "web"    # Will be available at web.yourdomain.com`,
						},
					],
				},
				{
					type: "step",
					number: 4,
					title: "Run the Tunnel",
					description: "Now you can start exposing your apps",
					children: [
						{
							type: "tabs",
							children: [
								{
									type: "tab",
									title: "Foreground Mode",
									children: [
										{
											type: "codeblock",
											language: "bash",
											code: "moley tunnel run",
										},
										{
											type: "paragraph",
											text: "Keep this terminal open. Your apps will be accessible while this command runs.",
											className: "text-gray-600 mt-3",
										},
									],
								},
								{
									type: "tab",
									title: "Background Mode",
									children: [
										{
											type: "codeblock",
											language: "bash",
											code: "moley tunnel run --detach",
										},
										{
											type: "paragraph",
											children: [
												{
													type: "text",
													text: "The tunnel runs in the background, and you get your terminal back immediately. If you encounter issues, check the ",
												},
												{
													type: "link",
													href: `/docs/troubleshooting/`,
													text: "troubleshooting guide",
												},
												{
													type: "text",
													text: ".",
												},
											],
											className: "text-gray-600 mt-3",
										},
									],
								},
							],
						},
					],
				},
			],
		},
	},

	// Configuration page
	{
		meta: {
			title: "Moley Configuration - Advanced Tunnel Setup Options",
			menuTitle: "Configuration",
			slug: "configuration",
			description:
				"Advanced configuration options for Moley tunnels. Learn YAML configuration, global settings, domain setup, and custom application routing for optimal performance.",
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
							text: "Moley uses a ",
						},
						{
							type: "inline-code",
							code: "moley.yml",
						},
						{
							type: "text",
							text: " file to configure your tunnels and applications. If you're new to Moley, start with the ",
						},
						{
							type: "link",
							href: `/docs/quick-start/`,
							text: "Quick Start guide",
						},
						{
							type: "text",
							text: " to generate your first configuration automatically.",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "moley.yml",
					code: `# Example moley.yml
tunnel:
  id: "unique-tunnel-id"
  name: "my-development-tunnel"

ingress:
  zone: "mydomain.com"
  apps:
    - target:
        port: 3000
        hostname: "localhost"
      expose:
        subdomain: "app"

    - target:
        port: 8080
        hostname: "localhost"
      expose:
        subdomain: "api"`,
				},

				{
					type: "heading",
					level: 2,
					text: "Global Configuration",
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "Global settings are stored in ",
						},
						{
							type: "inline-code",
							code: "~/.moley/config.yml",
						},
						{
							type: "text",
							text: ". This file is created during the ",
						},
						{
							type: "link",
							href: `/docs/installation/`,
							text: "installation process",
						},
						{
							type: "text",
							text: " when you set up your Cloudflare API token.",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "~/.moley/config.yml",
					code: `# Global configuration
cloudflare:
  token: "your-api-token"`,
				},
				{
					type: "paragraph",
					children: [
						{
							type: "text",
							text: "For security best practices when handling API tokens, see ",
						},
						{
							type: "link",
							href: "https://developers.cloudflare.com/fundamentals/api/get-started/create-token/",
							text: "Cloudflare's API token documentation",
							external: true,
							rel: "nofollow noopener noreferrer",
						},
						{
							type: "text",
							text: ".",
						},
					],
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Environment Variable Override",
					code: `# Override with environment variable instead of config file
export MOLEY_CLOUDFLARE_TOKEN="your-api-token"`,
				},

				{
					type: "heading",
					level: 2,
					text: "Ingress Configuration",
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
							code: "ingress",
						},
						{
							type: "text",
							text: " section defines how your local applications are exposed to the internet through your custom domain.",
						},
					],
				},
				{
					type: "heading",
					level: 3,
					text: "Zone Configuration",
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
							text: " field specifies your custom domain that must be configured in Cloudflare:",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "Zone Configuration",
					code: `ingress:
  zone: "yourdomain.com"  # Your Cloudflare-managed domain`,
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Environment Variable Override",
					code: `# Override zone with environment variable
export MOLEY_TUNNEL_INGRESS_ZONE="yourdomain.com"`,
				},
				{
					type: "infobox",
					style: "warning",
					title: "Domain Requirements",
					children: [
						{
							type: "list",
							style: "unordered",
							children: [
								{
									type: "listitem",
									text: "Domain must be registered with Cloudflare",
								},
								{
									type: "listitem",
									text: "DNS must be managed by Cloudflare (orange cloud enabled)",
								},
								{
									type: "listitem",
									text: "Your API token must have DNS edit permissions for this zone",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 3,
					text: "Apps Configuration",
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
							code: "apps",
						},
						{
							type: "text",
							text: " array defines which local services to expose and how to route traffic to them. Each app has two main sections: ",
						},
						{
							type: "inline-code",
							code: "target",
						},
						{
							type: "text",
							text: " (where the service runs locally) and ",
						},
						{
							type: "inline-code",
							code: "expose",
						},
						{
							type: "text",
							text: " (how it's accessible publicly).",
						},
					],
				},
				{
					type: "codeblock",
					language: "yaml",
					title: "Complete Apps Configuration",
					code: `ingress:
  zone: "yourdomain.com"
  apps:
    # Web application
    - target:
        port: 3000                    # Local port where your app runs
        hostname: "localhost"         # Local hostname (usually localhost)
      expose:
        subdomain: "app"             # Public URL: app.yourdomain.com

    # API service
    - target:
        port: 8080
        hostname: "localhost"
      expose:
        subdomain: "api"             # Public URL: api.yourdomain.com

    # Development server
    - target:
        port: 4000
        hostname: "127.0.0.1"        # Alternative to localhost
      expose:
        subdomain: "dev"             # Public URL: dev.yourdomain.com`,
				},

				{
					type: "heading",
					level: 4,
					text: "Target Configuration",
				},
				{
					type: "paragraph",
					text: "The target section specifies where your local service is running:",
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
									code: "port",
								},
								{
									type: "text",
									text: ": The local port number (1-65535) where your application is listening",
								},
							],
						},
						{
							type: "listitem",
							children: [
								{
									type: "inline-code",
									code: "hostname",
								},
								{
									type: "text",
									text: ": The local hostname, typically ",
								},
								{
									type: "inline-code",
									code: "localhost",
								},
								{
									type: "text",
									text: " or ",
								},
								{
									type: "inline-code",
									code: "127.0.0.1",
								},
							],
						},
					],
				},

				{
					type: "heading",
					level: 4,
					text: "Expose Configuration",
				},
				{
					type: "paragraph",
					text: "The expose section defines how the service is accessible publicly:",
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
									text: ": The subdomain prefix that creates the public URL. Must be alphanumeric with hyphens allowed.",
								},
							],
						},
					],
				},
				{
					type: "heading",
					level: 4,
					text: "Environment Variables",
				},
				{
					type: "codeblock",
					language: "bash",
					title: "Apps Environment Variables",
					code: `# Configure first app (frontend)
export MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_PORT="3000"
export MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_HOSTNAME="localhost"
export MOLEY_TUNNEL_INGRESS_APPS_0_EXPOSE_SUBDOMAIN="app"

# Configure second app (backend)
export MOLEY_TUNNEL_INGRESS_APPS_1_TARGET_PORT="8080"
export MOLEY_TUNNEL_INGRESS_APPS_1_TARGET_HOSTNAME="localhost"
export MOLEY_TUNNEL_INGRESS_APPS_1_EXPOSE_SUBDOMAIN="api"`,
				},

			],
		},
	},

	// Troubleshooting page
	{
		meta: {
			title: "Moley Troubleshooting - Fix Common Tunnel Issues",
			menuTitle: "Troubleshooting",
			slug: "troubleshooting",
			description:
				"Common issues and solutions when using Moley. Fix DNS problems, authentication errors, tunnel failures, and orphaned resources with detailed debugging steps.",
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
					text: "Quick Solutions",
				},
				{
					type: "section",
					spacing: "medium",
					children: [
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "Common issues and their immediate solutions. For step-by-step setup instructions, see our ",
								},
								{
									type: "link",
									href: `/docs/quick-start/`,
									text: "Quick Start guide",
								},
								{
									type: "text",
									text: " or ",
								},
								{
									type: "link",
									href: `/docs/configuration/`,
									text: "Configuration guide",
								},
								{
									type: "text",
									text: ":",
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
											type: "text",
											text: "Tunnel already exists: ",
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
											text: "Permission denied: ",
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
											text: "DNS not resolving: Wait 5 minutes or check ",
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
					],
				},
				{
					type: "heading",
					level: 2,
					text: "Detailed Troubleshooting",
				},
				{
					type: "section",
					spacing: "medium",
					children: [
						{
							type: "heading",
							level: 3,
							text: "DNS and connectivity issues",
						},
						{
							type: "paragraph",
							children: [
								{
									type: "text",
									text: "If your domain is not resolving or you can't access your local service through the tunnel, first ensure you've completed the ",
								},
								{
									type: "link",
									href: `/docs/installation/`,
									text: "installation steps",
								},
								{
									type: "text",
									text: " correctly:",
								},
							],
						},
						{
							type: "step",
							number: 1,
							title: "Test DNS resolution with dig",
							description:
								"Verify that your domain resolves correctly using DNS lookup tools",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Check DNS with dig",
									code: "# Check if your domain resolves\ndig your-domain.com\n\n# Check specific record type\ndig CNAME your-domain.com\n\n# Use specific DNS server\ndig @1.1.1.1 your-domain.com",
								},
							],
						},
						{
							type: "step",
							number: 2,
							title: "Test with nslookup as alternative",
							description:
								"Use nslookup as an alternative method to verify DNS resolution",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Check DNS with nslookup",
									code: "# Basic lookup\nnslookup your-domain.com\n\n# Use Cloudflare DNS\nnslookup your-domain.com 1.1.1.1",
								},
							],
						},
						{
							type: "step",
							number: 3,
							title: "Test external access",
							description:
								"Test connectivity to your exposed services from external networks",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Test External Access",
									code: "# Test HTTPS access\ncurl -I https://your-domain.com\n\n# Test with verbose output\ncurl -v https://your-domain.com\n\n# Test from different location\ncurl --connect-timeout 10 https://your-domain.com",
								},
							],
						},
						{
							type: "step",
							number: 4,
							title: "Check Cloudflare dashboard",
							description:
								"Verify DNS records and proxy settings in the Cloudflare dashboard",
							children: [
								{
									type: "list",
									style: "unordered",
									children: [
										{
											type: "listitem",
											text: "Verify DNS record exists in Cloudflare DNS tab",
										},
										{
											type: "listitem",
											text: "Check that the record points to your tunnel ID",
										},
										{
											type: "listitem",
											text: "Ensure proxy status (orange cloud) is enabled",
										},
										{
											type: "listitem",
											text: "Wait up to 5 minutes for DNS propagation",
										},
									],
								},
							],
						},
						{
							type: "heading",
							level: 3,
							text: "Tunnel initialization and startup failures",
						},
						{
							type: "paragraph",
							text: "If `moley tunnel init` or `moley tunnel run` commands fail:",
						},
						{
							type: "step",
							number: 1,
							title: "Check Cloudflare API token configuration",
							description:
								"Verify that your Cloudflare API token is properly configured and has the required permissions",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Verify API Token Setup",
									code: "# Check if global config exists\ncat ~/.moley/config.yml\n\n# Set API token if missing\nmoley config set --cloudflare.token=your_token_here\n\n# Verify token has correct permissions:\n# - Zone:Zone:Read\n# - Zone:DNS:Edit",
								},
							],
						},
						{
							type: "step",
							number: 2,
							title: "Check project configuration files",
							description:
								"Validate your project's moley.yml configuration file for syntax and structure",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Verify Project Setup",
									code: "# Check if moley.yml exists\ncat moley.yml\n\n# Initialize if missing\nmoley tunnel init\n\n# Check for YAML syntax errors\nmoley tunnel run --dry-run",
								},
							],
						},
						{
							type: "step",
							number: 3,
							title: "Verify cloudflared authentication",
							description:
								"Ensure cloudflared is properly authenticated with your Cloudflare account",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Check Cloudflared Setup",
									code: "# Ensure cloudflared is authenticated\ncloudflared tunnel login\n\n# Check existing tunnels\ncloudflared tunnel list\n\n# Verify credentials file\nls -la ~/.cloudflared/",
								},
							],
						},
						{
							type: "step",
							number: 4,
							title: "Use debug mode for detailed logs",
							description:
								"Enable verbose logging to diagnose configuration and connection issues",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Enable Debug Logging",
									code: "# Run with debug logging\nmoley tunnel run --log-level=debug\n\n# Or use trace for maximum verbosity\nmoley tunnel run --log-level=trace\n\n# Test configuration without changes\nmoley tunnel run --dry-run",
								},
							],
						},
						{
							type: "heading",
							level: 3,
							text: "Orphaned resources and cleanup issues",
						},
						{
							type: "paragraph",
							text: "If `moley tunnel stop` fails or resources remain after crashes:",
						},
						{
							type: "step",
							number: 1,
							title: "Check lock file and resource state",
							description:
								"Inspect the current state of tunnel resources and lock files",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Inspect Current State",
									code: "# Check lock file for tracked resources\ncat moley.lock\n\n# Verify project configuration\ncat moley.yml\n\n# List cloudflared tunnels\ncloudflared tunnel list",
								},
							],
						},
						{
							type: "step",
							number: 2,
							title: "Attempt normal stop with debug logging",
							description:
								"Try stopping the tunnel normally while collecting detailed diagnostic information",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Debug Stop Process",
									code: "# Try stopping with debug output\nmoley tunnel stop --log-level=debug\n\n# Use dry-run to see what would be cleaned\nmoley tunnel stop --dry-run\n\n# Check if background processes exist\nps aux | grep cloudflared",
								},
							],
						},
						{
							type: "step",
							number: 3,
							title: "Remove lock file to trigger orphaned resource cleanup",
							description:
								"Force detection and cleanup of orphaned resources by removing the lock file",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Force Orphaned Resource Detection",
									code: "# Remove lock file to trigger automatic cleanup\nrm -f moley.lock\n\n# Moley will now detect orphaned resources from moley.yml\nmoley tunnel stop --log-level=debug\n\n# Verify cleanup completed\ncloudflared tunnel list",
								},
							],
						},
						{
							type: "step",
							number: 4,
							title: "Manual cloudflared cleanup (last resort)",
							description:
								"Manually clean up cloudflared processes and tunnels when automatic cleanup fails",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Manual Tunnel Deletion",
									code: "# Kill any remaining cloudflared processes\npkill -f cloudflared\n\n# Get tunnel ID from moley.yml or cloudflared list\ncloudflared tunnel list\n\n# Delete tunnel manually\ncloudflared tunnel delete <tunnel-id>\n\n# Remove DNS records from Cloudflare dashboard\n# Go to DNS tab and remove CNAME records",
								},
							],
						},
					],
				},
				{
					type: "heading",
					level: 2,
					text: "Advanced Resources",
				},
				{
					type: "section",
					spacing: "medium",
					children: [
						{
							type: "paragraph",
							text: "For developers and advanced debugging scenarios:",
						},
						{
							type: "list",
							style: "unordered",
							children: [
								{
									type: "listitem",
									text: "Advanced Debugging and Logs - Detailed trace analysis and configuration inspection",
								},
								{
									type: "listitem",
									text: "Check the Moley CLI source code at cmd/ and internal/ directories for implementation details",
								},
								{
									type: "listitem",
									text: "Review lock file format and resource management in the codebase",
								},
							],
						},
					],
				},
			],
		},
	},

	// Advanced Debugging (Internal)
	{
		meta: {
			title: "Moley Advanced Debugging - Developer Troubleshooting Guide",
			menuTitle: "Advanced Debugging",
			slug: "advanced-debugging",
			description:
				"Detailed debugging techniques and log analysis for Moley developers. Advanced trace logging, configuration analysis, and resource management troubleshooting.",
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
					text: "This advanced debugging guide is designed for developers and power users who need deep insights into Moley's internal operations. These techniques help diagnose complex issues, analyze performance bottlenecks, and understand the detailed flow of tunnel creation and management.",
				},
				{
					type: "heading",
					level: 2,
					text: "When to Use Advanced Debugging",
				},
				{
					type: "paragraph",
					text: "Advanced debugging is particularly useful when basic troubleshooting doesn't resolve issues, when contributing to the Moley project, or when you need to understand exactly how resources are being managed. This includes scenarios with complex networking setups, custom Cloudflare configurations, or integration with CI/CD pipelines.",
				},
				{
					type: "heading",
					level: 2,
					text: "Debug Logging and Trace Analysis",
				},
				{
					type: "section",
					spacing: "medium",
					children: [
						{
							type: "paragraph",
							text: "For developers and advanced users who need detailed system analysis:",
						},
						{
							type: "step",
							number: 1,
							title: "Enable maximum verbosity logging",
							description:
								"Enable trace-level logging for detailed system analysis and debugging",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Trace Level Debugging",
									code: "# Enable trace logging for maximum detail\nmoley tunnel run --log-level=trace\n\n# Redirect output to file for analysis\nmoley tunnel run --log-level=trace 2>&1 | tee moley-debug.log\n\n# Check specific components\nmoley info --log-level=debug",
								},
							],
						},
						{
							type: "step",
							number: 2,
							title: "Analyze configuration files and state",
							description:
								"Examine all configuration files and system state for comprehensive analysis",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Configuration Analysis",
									code: "# Dump all configuration state\necho '=== Global Config ==='\ncat ~/.moley/config.yml\n\necho '=== Project Config ==='\ncat moley.yml\n\necho '=== Lock File State ==='\ncat moley.lock | jq .\n\necho '=== Cloudflared Tunnels ==='\ncloudflared tunnel list\n\necho '=== Generated Tunnel Configs ==='\nfind ~/.moley/tunnels -name '*.yml' -exec echo '{}:' \\; -exec cat '{}' \\;",
								},
							],
						},
						{
							type: "step",
							number: 3,
							title: "Resource manager debugging",
							description:
								"Debug resource management and API connectivity issues",
							children: [
								{
									type: "codeblock",
									language: "bash",
									title: "Resource State Analysis",
									code: "# Test resource detection without changes\nmoley tunnel stop --dry-run --log-level=debug\n\n# Check for orphaned resources\nrm moley.lock  # Force orphaned detection\nmoley tunnel stop --dry-run --log-level=trace\n\n# Verify Cloudflare API connectivity\ncurl -H 'Authorization: Bearer YOUR_TOKEN' 'https://api.cloudflare.com/client/v4/user/tokens/verify'",
								},
							],
						},
					],
				},
			],
		},
	},
];
