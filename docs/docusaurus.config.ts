import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'argus',
  tagline: 'The all-seeing code analyzer. Help AI grok your codebase.',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://priyans-hu.github.io',
  baseUrl: '/argus/',

  organizationName: 'Priyans-hu',
  projectName: 'argus',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/Priyans-hu/argus/tree/main/docs/',
        },
        blog: false, // Disable blog
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/argus-social-card.png',
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'argus',
      logo: {
        alt: 'argus Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/Priyans-hu/argus/releases',
          label: 'Releases',
          position: 'left',
        },
        {
          href: 'https://github.com/Priyans-hu/argus',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Getting Started',
              to: '/docs/getting-started',
            },
            {
              label: 'Installation',
              to: '/docs/installation',
            },
            {
              label: 'Configuration',
              to: '/docs/configuration',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub Discussions',
              href: 'https://github.com/Priyans-hu/argus/discussions',
            },
            {
              label: 'Issues',
              href: 'https://github.com/Priyans-hu/argus/issues',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/Priyans-hu/argus',
            },
            {
              label: 'Releases',
              href: 'https://github.com/Priyans-hu/argus/releases',
            },
          ],
        },
      ],
      copyright: `MIT License © ${new Date().getFullYear()} Priyanshu. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'go', 'yaml', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
