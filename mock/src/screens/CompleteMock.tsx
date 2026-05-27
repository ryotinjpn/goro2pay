import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import { COPY } from '../lib/copy';
import styles from './CompleteMock.module.css';

type State = {
  storeName?: string;
  amount?: number;
  balance?: number;
  monthlyCount?: number;
};

export default function CompleteMock() {
  const navigate = useNavigate();
  const { state } = useLocation() as { state: State | null };

  const storeName = state?.storeName ?? 'CoCo壱番屋 新宿店';
  const amount = state?.amount ?? 1200;
  const balance = state?.balance ?? 27_600;
  const monthlyCount = state?.monthlyCount ?? 6;

  useEffect(() => {
    const id = setTimeout(() => navigate('/main'), 8000);
    return () => clearTimeout(id);
  }, [navigate]);

  return (
    <ScreenFrame>
      <BrandHeader />
      <div className={styles.wrap}>
        <div className={styles.glow} />
        <div className={styles.headline}>{COPY.complete.headline}</div>
        <div className={styles.body}>{COPY.complete.body}</div>
        <div className={styles.store}>{storeName}</div>
        <div className={styles.amount}>¥{amount.toLocaleString()}</div>
        <div className={styles.divider} />
        <div className={styles.metrics}>
          {COPY.complete.metricsTemplate(monthlyCount)}
        </div>
        <div className={styles.balance}>
          残りダメ予算 ¥{balance.toLocaleString()}
        </div>
        <button
          type="button"
          className={styles.back}
          onClick={() => navigate('/main')}
        >
          {COPY.complete.backLink}
        </button>
      </div>
    </ScreenFrame>
  );
}
