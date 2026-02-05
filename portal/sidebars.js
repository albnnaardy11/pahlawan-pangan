import apiSidebar from './docs/api/sidebar';

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Architecture and Core',
      items: [
        'architecture/spatial-consistency',
      ],
    },
    {
      type: 'category',
      label: 'Integration Guides',
      items: [
        'integration/pos-automation',
      ],
    },
    {
      type: 'category',
      label: 'Security and Privacy',
      items: [
        'security/standards',
      ],
    },
    {
      type: 'category',
      label: 'Pahlawan Pangan API',
      link: {
        type: 'generated-index',
        title: 'Public API Reference',
        description: 'Explore our national-scale food redistribution APIs.',
        slug: '/api',
      },
      items: apiSidebar,
    },
  ],
};

export default sidebars;
