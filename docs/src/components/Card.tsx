import React, { type ReactNode } from 'react';
import Link from '@docusaurus/Link';
import type { LucideIcon } from 'lucide-react';
import styles from './Card.module.css';

type CardProps = {
  title: string;
  description?: string;
  href: string;
  icon?: LucideIcon;
  children?: ReactNode;
};

export default function Card({ title, description, href, icon: Icon, children }: CardProps) {
  return (
    <Link className={styles.card} to={href}>
      {Icon && (
        <span className={styles.icon}>
          <Icon size={24} strokeWidth={1.75} />
        </span>
      )}
      <span className={styles.body}>
        <span className={styles.title}>{title}</span>
        {description && <span className={styles.description}>{description}</span>}
        {children}
      </span>
    </Link>
  );
}
