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
        {to: 'docs/1_concepts/1_introduction', label: 'Docs', position: 'right'},
        {to: 'blog', label: 'Blog', position: 'right'},
        {
          href: 'https://github.com/erdrix/nifikop',
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
              to: 'docs/1_concepts/1_introduction',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/erdrix/nifikop',
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
              href: 'https://github.com/erdrix/nifikop/issues',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
    },
  },
  themes: ['@docusaurus/theme-classic', '@docusaurus/theme-live-codeblock'],
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


