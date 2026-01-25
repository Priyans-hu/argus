import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'getting-started',
    'installation',
    {
      type: 'category',
      label: 'Usage',
      items: [
        'usage/scan',
        'usage/watch',
        'usage/output',
      ],
    },
    'configuration',
    {
      type: 'category',
      label: 'Features',
      items: [
        'features/tech-stack',
        'features/architecture',
        'features/conventions',
        'features/patterns',
      ],
    },
    'contributing',
  ],
};

export default sidebars;
