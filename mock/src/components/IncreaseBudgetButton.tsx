import { COPY } from '../lib/copy';
import styles from './IncreaseBudgetButton.module.css';

type Props = {
  recommended: number;
  onClick?: () => void;
};

export default function IncreaseBudgetButton({ recommended, onClick }: Props) {
  return (
    <div className={styles.wrap}>
      <button type="button" className={styles.btn} onClick={onClick}>
        {COPY.main.increaseTemplate(recommended)}
      </button>
    </div>
  );
}
