import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import CodeBlock from '@theme/CodeBlock';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero', styles.heroBanner)}>
      <div className="container">
        <div className={styles.heroContent}>
          <span className={styles.badge}>Open Source CLI Tool</span>
          <Heading as="h1" className={styles.heroTitle}>
            The All-Seeing<br />Code Analyzer
          </Heading>
          <p className={styles.heroSubtitle}>
            Generate comprehensive context files that help AI assistants
            understand your codebase structure, conventions, and patterns.
          </p>
          <div className={styles.buttons}>
            <Link
              className="button button--primary button--lg"
              to="/docs/getting-started">
              Get Started →
            </Link>
            <Link
              className="button button--secondary button--lg"
              href="https://github.com/Priyans-hu/argus">
              GitHub
            </Link>
          </div>
        </div>

        <div className={styles.codePreview}>
          <div className={styles.codeHeader}>
            <span className={styles.codeDot} style={{background: '#ef4444'}} />
            <span className={styles.codeDot} style={{background: '#eab308'}} />
            <span className={styles.codeDot} style={{background: '#22c55e'}} />
          </div>
          <div className={styles.codeContent}>
            <code>
              <span style={{color: '#22c55e'}}>$</span> argus scan .<br />
              <span style={{color: '#a1a1aa'}}>🔍 Scanning /my-project...</span><br />
              <span style={{color: '#a1a1aa'}}>📊 Analyzing tech stack...</span><br />
              <span style={{color: '#a1a1aa'}}>🏗️  Detecting architecture...</span><br />
              <span style={{color: '#a1a1aa'}}>📝 Finding conventions...</span><br />
              <span style={{color: '#22c55e'}}>✅ Generated CLAUDE.md</span>
            </code>
          </div>
        </div>
      </div>
    </header>
  );
}

const features = [
  {
    title: 'Tech Stack Detection',
    icon: '🔍',
    description: 'Automatically identifies languages, frameworks, libraries, and tools used in your project.',
  },
  {
    title: 'Architecture Analysis',
    icon: '🏗️',
    description: 'Understands your project structure, layer dependencies, and architectural patterns.',
  },
  {
    title: 'Convention Detection',
    icon: '📋',
    description: 'Discovers coding standards, naming conventions, and established practices in your codebase.',
  },
  {
    title: 'Command Discovery',
    icon: '⚡',
    description: 'Finds and prioritizes build, test, lint, and run commands from your project configuration.',
  },
  {
    title: 'Git Integration',
    icon: '🔄',
    description: 'Analyzes commit history, branch patterns, and contributor conventions.',
  },
  {
    title: 'Multi-language Support',
    icon: '🌍',
    description: 'Works with Go, JavaScript, TypeScript, Python, Rust, and many more languages.',
  },
];

function Feature({title, icon, description}: {title: string; icon: string; description: string}) {
  return (
    <div className={styles.featureCard}>
      <div className={styles.featureIcon}>{icon}</div>
      <Heading as="h3" className={styles.featureTitle}>{title}</Heading>
      <p className={styles.featureDesc}>{description}</p>
    </div>
  );
}

function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <Heading as="h2" className={styles.sectionTitle}>Features</Heading>
        <p className={styles.sectionSubtitle}>Everything you need to help AI understand your code</p>
        <div className={styles.featuresGrid}>
          {features.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

function InstallSection() {
  return (
    <section className={styles.install}>
      <div className="container">
        <Heading as="h2" className={styles.sectionTitle}>Installation</Heading>
        <p className={styles.sectionSubtitle}>Choose your preferred installation method</p>

        <div className={styles.installGrid}>
          <div className={styles.installCard}>
            <h4>Homebrew (macOS/Linux)</h4>
            <CodeBlock language="bash">
              brew install Priyans-hu/tap/argus
            </CodeBlock>
          </div>

          <div className={styles.installCard}>
            <h4>Go Install</h4>
            <CodeBlock language="bash">
              go install github.com/Priyans-hu/argus/cmd/argus@latest
            </CodeBlock>
          </div>

          <div className={styles.installCard}>
            <h4>Download Binary</h4>
            <p>
              <Link href="https://github.com/Priyans-hu/argus/releases">
                Download from Releases →
              </Link>
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  return (
    <Layout
      title="The All-Seeing Code Analyzer"
      description="Generate comprehensive context files that help AI assistants understand your codebase structure, conventions, and patterns.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <InstallSection />
      </main>
    </Layout>
  );
}
