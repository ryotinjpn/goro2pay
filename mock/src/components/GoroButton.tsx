import { ReactNode } from 'react';
import SuggestBubble from './SuggestBubble';
import { COPY } from '../lib/copy';
import styles from './GoroButton.module.css';

export type GoroButtonState = 'idle' | 'suggested' | 'slot' | 'dead';

type Props = {
  state: GoroButtonState;
  mainLabel: string;
  subLabel: string;
  onClick?: () => void;
  ariaLabel?: string;
  winningKey?: string | number;
  children?: ReactNode;
};

export default function GoroButton({
  state,
  mainLabel,
  subLabel,
  onClick,
  ariaLabel,
  winningKey,
  children,
}: Props) {
  const disabled = state === 'slot' || state === 'dead';
  const pressed = state === 'slot';
  const winning = winningKey !== undefined;

  return (
    <div className={styles.wrap}>
      {state === 'suggested' && <SuggestBubble text={COPY.main.suggestBubble} />}
      <button
        key={winning ? `win-${winningKey}` : 'btn'}
        type="button"
        className={styles.btn}
        data-state={state}
        data-pressed={pressed}
        data-winning={winning}
        onClick={onClick}
        disabled={disabled}
        aria-label={ariaLabel ?? mainLabel}
      >
        {state === 'slot' ? (
          children
        ) : (
          <>
            <span className={styles.main}>{mainLabel}</span>
            <span className={styles.sub}>{subLabel}</span>
          </>
        )}
      </button>
      {winning && <span className={styles.flare} aria-hidden />}
    </div>
  );
}
