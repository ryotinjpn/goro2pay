import { COPY } from '../lib/copy';
import styles from './DeadVerdict.module.css';

export default function DeadVerdict() {
  return (
    <div className={styles.verdict} role="status" aria-live="polite">
      {COPY.main.deadVerdict}
    </div>
  );
}
