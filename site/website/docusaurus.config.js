module.exports = {
  title: 'NiFiKop',
  tagline: 'Open-Source, Apache NiFi operator for Kubernetes',
  organizationName: 'Konpyūtāika',
  projectName: 'nifikop',
  url: 'https://konpyutaika.github.io',
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
            'https://github.com/konpyutaika/nifikop/edit/master/site/website/',
          path: '../docs',
          numberPrefixParser: false,
          sidebarPath: require.resolve('./sidebars.json'),
          // remarkPlugins: [require('@react-native-website/remark-snackplayer')],
        },
        blog: {
          path: 'blog',
          blogSidebarCount: 'ALL',
          blogSidebarTitle: 'All Blog Posts',
          feedOptions: {
            type: 'all',
            copyright: `Copyright © ${new Date().getFullYear()} Konpyūtāika, Inc. Built with Docusaurus.`,
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
      appId: 'VBPOAOSEIB',
      apiKey: 'd301810165188ea9095145222463ef55',
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
        {to: 'docs/1_concepts/1_start_here', className: 'header-doc-link', 'aria-label': 'Documentation', position: 'right'},
        {to: 'blog', className: 'header-blog-link', 'aria-label': 'Blog', position: 'right'},
        {
          href: 'https://github.com/konpyutaika/nifikop',
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
              to: 'docs/1_concepts/1_start_here',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/konpyutaika/nifikop',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Slack',
              href: 'https://join.slack.com/t/konpytika/shared_invite/zt-14md072lv-Jr8mqYoeUrqzfZF~YGUpXA',
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
              href: 'https://github.com/konpyutaika/nifikop/issues',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
    },
  },
};
