import { useEffect, useState } from 'react';
import styles from './BrandHeader.module.css';

function formatTime(d: Date): string {
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  return `${hh}:${mm}`;
}

export default function BrandHeader() {
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 30_000);
    return () => clearInterval(id);
  }, []);

  return (
    <header className={styles.header}>
      <span className={styles.brand}>
        <span className={styles.brandKana}>ゴロゴロ</span>
        <span className={styles.brandLatin}>Pay</span>
      </span>
      <span className={styles.time}>{formatTime(now)}</span>
    </header>
  );
}
