import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import { themes as prismThemes } from 'prism-react-renderer';

const config: Config = {
  title: 'Moley',
  tagline: 'Expose localhost with Cloudflare Tunnels',
  favicon: 'img/favicon.svg',

  url: 'https://moley.dev',
  baseUrl: '/',

  organizationName: 'stupside',
  projectName: 'moley',

  onBrokenLinks: 'throw',

  trailingSlash: true,

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  stylesheets: [
    {
      href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;900&family=JetBrains+Mono:wght@400;500;600&display=swap',
      rel: 'stylesheet',
    },
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          routeBasePath: '/docs',
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/stupside/moley/edit/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
        sitemap: {
          changefreq: 'weekly',
          priority: 0.8,
          lastmod: 'date',
        },
      } satisfies Preset.Options,
    ],
  ],

  themes: [
    [
      '@easyops-cn/docusaurus-search-local',
      {
        hashed: true,
        indexBlog: false,
        indexPages: false,
        docsRouteBasePath: '/docs',
        highlightSearchTermsOnTargetPage: true,
      },
    ],
  ],

  themeConfig: {
    image: 'img/moley.svg',
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Moley',
      logo: {
        alt: '',
        src: 'img/moley.svg',
      },
      items: [
        { to: '/docs/', label: 'Docs', position: 'left' },
        {
          href: 'https://github.com/stupside/moley',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      links: [
        {
          title: 'Getting started',
          items: [
            { label: 'Installation', to: '/docs/installation/' },
            { label: 'Quick Start', to: '/docs/quick-start/' },
            { label: 'Use cases', to: '/docs/use-cases/' },
          ],
        },
        {
          title: 'Reference',
          items: [
            { label: 'CLI', to: '/docs/cli/' },
            { label: 'Configuration', to: '/docs/configuration/' },
            { label: 'Docker', to: '/docs/docker/' },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/stupside/moley',
            },
            {
              label: 'Issues',
              href: 'https://github.com/stupside/moley/issues',
            },
            {
              label: 'Releases',
              href: 'https://github.com/stupside/moley/releases',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Moley.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.oneDark,
      additionalLanguages: ['bash', 'yaml', 'go'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
