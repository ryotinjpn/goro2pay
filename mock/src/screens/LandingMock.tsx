import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import SlotReel from '../components/SlotReel';
import { COPY } from '../lib/copy';
import styles from './LandingMock.module.css';

type DemoState = 'idle' | 'slot' | 'done';

export default function LandingMock() {
  const navigate = useNavigate();
  const [demo, setDemo] = useState<DemoState>('idle');

  const onDemo = () => {
    if (demo !== 'idle') return;
    setDemo('slot');
    setTimeout(() => setDemo('done'), 1400);
  };

  const remaining = demo === 'done' ? 0 : 1000;

  return (
    <ScreenFrame>
      <BrandHeader />

      <div className={styles.hero}>
        <p className={styles.eyebrow}>{COPY.landing.sub}</p>
        <h1 className={styles.h1}>{COPY.landing.h1Top}</h1>
        <h1 className={styles.h1}>{COPY.landing.h1Bottom}</h1>
      </div>

      <div className={styles.demoLabel}>{COPY.landing.demoLabel}</div>
      <div className={styles.demoAmount}>¥{remaining.toLocaleString()}</div>

      <div className={styles.demoBtnWrap}>
        {demo === 'slot' ? (
          <div className={styles.demoSlot}>
            <SlotReel />
          </div>
        ) : (
          <button
            type="button"
            className={styles.demoBtn}
            data-done={demo === 'done'}
            onClick={onDemo}
            disabled={demo !== 'idle'}
          >
            <span>{demo === 'done' ? '済' : COPY.landing.demoButtonInitial}</span>
            {demo === 'done' && (
              <span className={styles.demoBtnSub}>
                {COPY.landing.demoButtonDone}
              </span>
            )}
          </button>
        )}
      </div>

      <div className={styles.divider} />

      <div className={styles.ctaWrap} data-hidden={demo !== 'done'}>
        <button
          type="button"
          className={styles.ctaPrimary}
          onClick={() => navigate('/main')}
        >
          {COPY.landing.ctaPrimary}
          <span className={styles.arrow} aria-hidden>
            →
          </span>
        </button>
        <button
          type="button"
          className={styles.ctaSecondary}
          onClick={() => navigate('/main')}
        >
          {COPY.landing.ctaSecondary}
        </button>
      </div>
    </ScreenFrame>
  );
}
