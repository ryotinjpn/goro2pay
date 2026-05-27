import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './RaiseBudgetModalMock.module.css';

const RECOMMENDED = 45_000;

export default function RaiseBudgetModalMock() {
  const navigate = useNavigate();
  const close = () => navigate('/main');

  return (
    <>
      <ScreenFrame>
        <BrandHeader />
        <div className={styles.echo} aria-hidden />
      </ScreenFrame>
      <div className={styles.overlay}>
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="raise-modal-title"
          className={styles.card}
        >
          <h2 id="raise-modal-title" className={styles.h2}>
            {COPY.raiseBudget.h2}
          </h2>

          <p className={styles.recommend}>
            {COPY.raiseBudget.recommendLabel}
            <span className={styles.recommendAmount}>
              ¥{RECOMMENDED.toLocaleString()}
            </span>
            <span className={styles.recommendNote}>
              {COPY.raiseBudget.recommendNote}
            </span>
          </p>

          <div className={styles.actions}>
            <button onClick={close} className={styles.primary}>
              {COPY.raiseBudget.primary}
            </button>
            <button onClick={close} className={styles.secondary}>
              {COPY.raiseBudget.secondary}
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
