import { REEL_STRIP } from '../lib/fakeApi';
import styles from './SlotReel.module.css';

type Props = {
  result?: { shortName: string; amount: number };
};

export default function SlotReel({ result }: Props) {
  const stopped = result !== undefined;
  const strip = stopped
    ? [result.shortName]
    : [...REEL_STRIP, REEL_STRIP[0]];
  const flashKey = stopped ? `flash-${result.shortName}` : 'spinning';

  return (
    <>
      <div
        key={flashKey}
        className={styles.window}
        data-flash={stopped}
        aria-hidden
      >
        <div className={styles.strip} data-stopped={stopped}>
          {strip.map((s, i) => (
            <div key={i} className={styles.line}>
              {s}
            </div>
          ))}
        </div>
        <div className={styles.scanline} />
      </div>
    </>
  );
}
