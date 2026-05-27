import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './InsufficientBalanceModalMock.module.css';

export default function InsufficientBalanceModalMock() {
  const navigate = useNavigate();
  const close = () => navigate('/main');

  return (
    <>
      <ScreenFrame>
        <BrandHeader />
        <div className={styles.echo} aria-hidden />
      </ScreenFrame>
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="insufficient-balance-title"
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="insufficient-balance-title" className={styles.h2}>
            {COPY.insufficient.h2}
          </h2>
          <p className={styles.sub}>
            {COPY.insufficient.bodyLine1}
            <br />
            {COPY.insufficient.bodyLine2}
          </p>
          <div className={styles.actions}>
            <button onClick={close} className={styles.secondary}>
              {COPY.insufficient.secondary}
            </button>
            <button
              onClick={() => navigate('/budget')}
              className={styles.primary}
            >
              {COPY.insufficient.primary}
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
