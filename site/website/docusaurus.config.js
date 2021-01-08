module.exports = {
  title: 'NiFiKop',
  tagline: 'Open-Source, Apache NiFi operator for Kubernetes',
  organizationName: 'Orange-OpenSource',
  projectName: 'nifikop',
  url: 'https://orange-opensource.github.io',
  baseUrl: '/nifikop/',
  clientModules: [require.resolve('./snackPlayerInitializer.js')],

  scripts: [
    {
      src:
        'https://cdn.jsdelivr.net/npm/focus-visible@5.2.0/dist/focus-visible.min.js',
      defer: true,
    },
    {src: 'https://snack.expo.io/embed.js', defer: true},
  ],
  favicon: 'img/nifikop.ico',
  titleDelimiter: '·',
  onBrokenLinks: 'throw',
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          showLastUpdateAuthor: true,
          showLastUpdateTime: true,
          editUrl:
            'https://github.com/Orange-OpenSource/nifikop/edit/master/site/website/',
          path: '../docs',
          sidebarPath: require.resolve('./sidebars.json'),
          // remarkPlugins: [require('@react-native-website/remark-snackplayer')],
        },
        blog: {
          path: 'blog',
          blogSidebarCount: 'ALL',
          blogSidebarTitle: 'All Blog Posts',
          feedOptions: {
            type: 'all',
            copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
          },
        },
        theme: {
          customCss: [
            require.resolve('./src/css/customTheme.scss'),
            require.resolve('./src/css/index.scss'),
            // require.resolve('./src/css/showcase.scss'),
            require.resolve('./src/css/versions.scss'),
          ],
        },
      },
    ],
  ],
  plugins: ['@docusaurus/theme-live-codeblock', 'docusaurus-plugin-sass', './sitePlugin'],
  themeConfig: {
    algolia: {
      apiKey: '34dbf55751628f3e3aaf8e06776fba0b',
      indexName: 'nifikop',

      // Optional: see doc section bellow
      contextualSearch: true,

      // Optional: Algolia search parameters
      searchParameters: {},

    },
    prism: {
      defaultLanguage: 'jsx',
      theme: require('./core/PrismTheme'),
      additionalLanguages: ['java', 'kotlin', 'objectivec', 'swift', 'groovy'],
    },
    navbar: {
      title: 'NiFiKop',
      logo: {
        alt: 'NiFiKop Logo',
        src: 'img/nifikop.png',
      },
      style: 'dark',
      items: [
        {to: 'docs/1_concepts/1_introduction', className: 'header-doc-link', 'aria-label': 'Documentation', position: 'right'},
        {to: 'blog', className: 'header-blog-link', 'aria-label': 'Blog', position: 'right'},
        {
          href: 'https://github.com/Orange-OpenSource/nifikop',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
        {
          type: 'docsVersionDropdown',
          position: 'left',
          dropdownActiveClassDisabled: true,
          dropdownItemsAfter: [
            {
              to: '/versions',
              label: 'All versions',
            },
          ],
        },
      ],
    },
    image: 'img/logo-og.png',
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Getting Started',
          items: [
            {
              label: 'Documentation',
              to: 'docs/1_concepts/1_introduction',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/Orange-OpenSource/nifikop',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Slack',
              href: 'https://nifikop.slack.com',
            },
            {
              label: 'Blog',
              to: 'blog',
            },
          ],
        },
        {
          title: 'Contact',
          items: [
            {
              label: 'Feature request',
              href: 'https://github.com/Orange-OpenSource/nifikop/issues',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
    },
  },
};
