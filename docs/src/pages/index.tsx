import React from 'react';
import Layout from '@theme/Layout';
import Hero from '../components/Hero';
import styles from './index.module.css';

const withoutMoley = [
  'Authenticate with Cloudflare',
  'Create tunnel manually',
  'Write YAML config with ingress rules',
  'Route DNS for each subdomain',
  'Run the tunnel',
  'Manually delete tunnel + DNS records when done',
];

export default function Home() {
  return (
    <Layout
      title="Moley — Cloudflare Tunnels from a config file"
      description="Runs Cloudflare Tunnels, DNS, and Access from a config file. No GUI, no vendor lock-in, MIT-licensed."
    >
      <Hero
        title="Use Cloudflared. "
        titleAccent="Without the clicks."
        subtitle="moley runs Cloudflare Tunnels, DNS, and Access from a config file. No GUI, no vendor lock-in, MIT-licensed."
        logoSrc="img/moley.svg"
        logoAlt="Moley"
      />

      <section className={styles.comparison}>
        <div className={styles.container}>
          <h2 className={styles.sectionTitle}>Without Moley</h2>
          <div className={styles.stepList}>
            {withoutMoley.map((step, i) => (
              <div key={step}>
                <span className={styles.stepNumber}>{i + 1}</span>
                {step}
              </div>
            ))}
          </div>

          <h2 className={styles.sectionTitleAccent}>With Moley</h2>
          <div className={styles.command}>
            <span className={styles.prompt}>$</span>
            <span className={styles.commandText}>moley tunnel run</span>
          </div>
          <p className={styles.tagline}>
            Creates the tunnel, generates config, sets up DNS, runs it, and cleans up when
            you stop. That's it.
          </p>
        </div>
      </section>
    </Layout>
  );
}
