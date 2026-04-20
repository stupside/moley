import React from 'react';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './Hero.module.css';

type HeroProps = {
  title: string;
  titleAccent: string;
  subtitle: string;
  logoSrc: string;
  logoAlt: string;
};

export default function Hero({ title, titleAccent, subtitle, logoSrc, logoAlt }: HeroProps) {
  return (
    <section className={styles.hero}>
      <img className={styles.logo} src={useBaseUrl(logoSrc)} alt={logoAlt} />
      <h1 className={styles.title}>
        {title}
        <span className={styles.titleAccent}>{titleAccent}</span>
      </h1>
      <p className={styles.subtitle}>{subtitle}</p>
      <div className={styles.ctaRow}>
        <Link className="button button--primary button--lg" to="/docs/">
          Docs
        </Link>
        <Link
          className="button button--secondary button--lg"
          to="https://github.com/stupside/moley"
        >
          GitHub
        </Link>
      </div>
    </section>
  );
}
