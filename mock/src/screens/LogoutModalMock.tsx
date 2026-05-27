import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './LogoutModalMock.module.css';

export default function LogoutModalMock() {
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
        aria-labelledby="logout-confirm-title"
        className={styles.overlay}
      >
        <div className={styles.card}>
          <h2 id="logout-confirm-title" className={styles.h2}>
            {COPY.logout.h2}
          </h2>
          <p className={styles.sub}>{COPY.logout.sub}</p>
          <div className={styles.actions}>
            <button
              type="button"
              onClick={close}
              className={styles.secondary}
            >
              {COPY.logout.secondary}
            </button>
            <button
              type="button"
              onClick={() => navigate('/login')}
              className={styles.primary}
            >
              {COPY.logout.primary}
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
