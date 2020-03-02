module.exports = {
  title: 'NiFiKop',
  tagline: 'Open-Source, Apache NiFi operator for Kubernetes',
  url: 'https://erdrix.github.io',
  baseUrl: '/nifikop/',
  favicon: 'img/nifikop.ico',
  organizationName: 'erdrix', // Usually your GitHub org/user name.
  projectName: 'nifikop', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'NiFiKop',
      logo: {
        alt: 'NiFiKop Logo',
        src: 'img/nifikop.png',
      },
      links: [
        {to: 'docs/overview', label: 'Docs', position: 'left'},
        {to: 'blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/Orange-OpenSource/nifikop',
          label: 'GitHub',
          position: 'right',
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
              to: 'docs/overview',
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
            {
              label: 'Twitter',
              href: 'https://twitter.com',
            },
          ],
        },
        {
          title: 'Contact',
          items: [
            {
              label: 'Email',
              href: 'mailto:prj.nifikop.support@list.orangeportails.net',
            },
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
            'https://erdrix.github.io/nifikop/edit/master/website/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
