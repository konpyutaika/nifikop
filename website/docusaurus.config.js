const versions = require('./versions.json');

const allDocHomesPaths = [
  '/docs/',
  '/docs/next/',
  ...versions.slice(1).map((version) => `/docs/${version}/`),
];


module.exports = {
  title: 'NiFiKop',
  tagline: 'Open-Source, Apache NiFi operator for Kubernetes',
  url: 'https://orange-opensource.github.io',
  baseUrl: '/nifikop/',
  favicon: 'img/nifikop.ico',
  organizationName: 'Orange-OpenSource', // Usually your GitHub org/user name.
  projectName: 'nifikop', // Usually your repo name.
  themeConfig: {
    algolia: {
      apiKey: '34dbf55751628f3e3aaf8e06776fba0b',
      indexName: 'nifikop',

      // Optional: see doc section bellow
      contextualSearch: true,

      // Optional: Algolia search parameters
      searchParameters: {},

    },
    navbar: {
      title: 'NiFiKop',
      logo: {
        alt: 'NiFiKop Logo',
        src: 'img/nifikop.png',
      },
      items: [
        {
          type: 'docsVersionDropdown',
          position: 'right',
          dropdownActiveClassDisabled: true,
          dropdownItemsAfter: [
            {
              to: '/versions',
              label: 'All versions',
            },
          ],
        },
        {to: 'docs/1_concepts/1_introduction', className: 'header-doc-link', 'aria-label': 'Documentation', position: 'right'},
        {to: 'blog', className: 'header-blog-link', 'aria-label': 'Blog', position: 'right'},
        {
          href: 'https://github.com/Orange-OpenSource/nifikop',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
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
  themes: ['@docusaurus/theme-live-codeblock'],
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
              'https://github.com/Orange-OpenSource/nifikop/edit/master/website/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        editCurrentVersion: true,
        showLastUpdateAuthor: true,
        showLastUpdateTime: true,
        remarkPlugins: [
          [require('@docusaurus/remark-plugin-npm2yarn'), {sync: true}],
        ],
        disableVersioning: false,
        onlyIncludeVersions: ['current', ...versions.slice(0, 2)],
      },
    ],
  ],
};


