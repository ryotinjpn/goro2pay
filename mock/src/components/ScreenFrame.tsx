import { ReactNode } from 'react';
import styles from './ScreenFrame.module.css';

type Props = {
  children: ReactNode;
  dead?: boolean;
  overlay?: ReactNode;
};

export default function ScreenFrame({ children, dead = false, overlay }: Props) {
  return (
    <div className={styles.outer}>
      <div className={styles.column}>
        <div className={styles.frame} data-dead={dead}>
          {children}
        </div>
        {overlay && <div className={styles.overlay}>{overlay}</div>}
      </div>
    </div>
  );
}
