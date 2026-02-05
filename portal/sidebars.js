import apiSidebar from './docs/api/sidebar';

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure
  tutorialSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Core Documentation',
      items: [
        'tutorial-basics/create-a-document',
        'tutorial-basics/congratulations',
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
