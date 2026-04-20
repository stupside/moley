import React, { type ReactNode } from 'react';
import styles from './Step.module.css';

type StepProps = {
  number: number;
  title: string;
  description?: string;
  children?: ReactNode;
};

export default function Step({ number, title, description, children }: StepProps) {
  return (
    <div className={styles.step}>
      <div className={styles.number}>{number}</div>
      <div className={styles.body}>
        <h3 className={styles.title}>{title}</h3>
        {description && <p className={styles.description}>{description}</p>}
        <div className={styles.content}>{children}</div>
      </div>
    </div>
  );
}
