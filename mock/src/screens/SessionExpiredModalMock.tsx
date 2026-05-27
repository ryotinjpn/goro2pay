import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './SessionExpiredModalMock.module.css';

export default function SessionExpiredModalMock() {
  const navigate = useNavigate();

  return (
    <>
      <ScreenFrame>
        <BrandHeader />
        <div className={styles.echo} aria-hidden />
      </ScreenFrame>
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="session-expired-title"
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="session-expired-title" className={styles.h2}>
            {COPY.sessionExpired.h2}
          </h2>
          <p className={styles.sub}>{COPY.sessionExpired.body}</p>
          <button
            type="button"
            onClick={() => navigate('/login')}
            className={styles.button}
          >
            {COPY.sessionExpired.button}
          </button>
        </div>
      </div>
    </>
  );
}
