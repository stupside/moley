import React, { type ReactNode } from 'react';
import styles from './Card.module.css';

export default function CardGrid({ children }: { children: ReactNode }) {
  return <div className={styles.grid}>{children}</div>;
}
