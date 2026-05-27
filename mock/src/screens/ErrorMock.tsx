import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './ErrorMock.module.css';

export default function ErrorMock() {
  const navigate = useNavigate();

  return (
    <ScreenFrame>
      <BrandHeader />
      <div className={styles.center} role="alert" aria-live="assertive">
        <h1 className={styles.h1}>{COPY.error.h1}</h1>
        <p className={styles.body}>{COPY.error.body}</p>
        <button
          type="button"
          className={styles.button}
          onClick={() => navigate('/')}
        >
          {COPY.error.button}
        </button>
      </div>
    </ScreenFrame>
  );
}
