import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import BalanceHero from '../components/BalanceHero';
import GoroButton, { GoroButtonState } from '../components/GoroButton';
import SlotReel from '../components/SlotReel';
import DeadVerdict from '../components/DeadVerdict';
import IncreaseBudgetButton from '../components/IncreaseBudgetButton';
import MetricsPanel from '../components/MetricsPanel';
import { COPY } from '../lib/copy';
import { OrderResult, placeOrder, suggestNext } from '../lib/fakeApi';
import { playWinChime, setSoundEnabled } from '../lib/sound';
import styles from './MainMock.module.css';

const INITIAL_BUDGET = 30_000;
const RECOMMENDED_RAISE = 50_000;

type DebugMode = 'auto' | 'idle' | 'suggested' | 'dead';
const DEBUG_MODES: DebugMode[] = ['auto', 'idle', 'suggested', 'dead'];

export default function MainMock() {
  const navigate = useNavigate();

  const [balance, setBalance] = useState(INITIAL_BUDGET);
  const [monthlyCount, setMonthlyCount] = useState(0);
  const [state, setState] = useState<GoroButtonState>('idle');
  const [slotResult, setSlotResult] = useState<OrderResult | undefined>();
  const [debug, setDebug] = useState<DebugMode>('auto');
  const [sound, setSound] = useState(false);
  const [winFlash, setWinFlash] = useState(0);

  const suggestion = suggestNext();

  useEffect(() => {
    if (debug !== 'auto') return;
    if (state !== 'idle') return;
    if (balance === 0) return;
    if (monthlyCount < 1) return;
    const id = setTimeout(() => setState('suggested'), 1400);
    return () => clearTimeout(id);
  }, [debug, state, balance, monthlyCount]);

  const onPress = async () => {
    if (state === 'slot' || state === 'dead') return;

    setSlotResult(undefined);
    setState('slot');

    const result = await placeOrder();
    setSlotResult(result);
    setWinFlash((n) => n + 1);
    if (sound) playWinChime();

    setTimeout(() => {
      const newBalance = Math.max(0, balance - result.amount);
      const newCount = monthlyCount + 1;
      setBalance(newBalance);
      setMonthlyCount(newCount);

      navigate('/main/complete', {
        state: {
          storeName: result.storeName,
          amount: result.amount,
          balance: newBalance,
          monthlyCount: newCount,
        },
      });
    }, 1800);
  };

  const onIncrease = () => {
    setBalance(RECOMMENDED_RAISE);
    setState('idle');
    setMonthlyCount(0);
    setDebug('auto');
  };

  const setDebugMode = (mode: DebugMode) => {
    setDebug(mode);
    if (mode === 'idle') {
      setBalance(INITIAL_BUDGET - 4_300);
      setMonthlyCount(5);
      setState('idle');
      setSlotResult(undefined);
    } else if (mode === 'suggested') {
      setBalance(INITIAL_BUDGET - 4_300);
      setMonthlyCount(5);
      setState('suggested');
      setSlotResult(undefined);
    } else if (mode === 'dead') {
      setBalance(0);
      setMonthlyCount(24);
      setState('dead');
      setSlotResult(undefined);
    }
  };

  const dead = state === 'dead' || balance === 0;
  const effectiveState: GoroButtonState = dead ? 'dead' : state;

  const mainLabel =
    effectiveState === 'suggested'
      ? COPY.main.btnSuggestMain
      : COPY.main.btnIdleMain;
  const subLabel =
    effectiveState === 'dead'
      ? COPY.main.btnDeadSub
      : effectiveState === 'suggested'
        ? `— ${suggestion.shortName} ¥${suggestion.amount.toLocaleString()} だ。`
        : COPY.main.btnIdleSub;

  const onToggleSound = () => {
    const next = !sound;
    setSound(next);
    setSoundEnabled(next);
  };

  return (
    <>
      <div className={styles.debug} aria-hidden>
        {DEBUG_MODES.map((m) => (
          <button
            key={m}
            data-active={debug === m}
            onClick={() => (m === 'auto' ? setDebug('auto') : setDebugMode(m))}
          >
            {m}
          </button>
        ))}
      </div>
      <button
        type="button"
        className={styles.soundToggle}
        data-on={sound}
        onClick={onToggleSound}
        aria-label={sound ? '音オフ' : '音オン'}
      >
        {sound ? '♪ on' : '♪ off'}
      </button>

      {winFlash > 0 && (
        <div
          key={`flash-${winFlash}`}
          className={styles.winFlash}
          aria-hidden
        />
      )}
      {winFlash > 0 && effectiveState === 'slot' && (
        <div
          key={`verdict-${winFlash}`}
          className={styles.winVerdict}
          role="status"
          aria-live="polite"
        >
          {COPY.main.winVerdict}
        </div>
      )}

      <ScreenFrame
        dead={dead}
        overlay={
          dead ? (
            <IncreaseBudgetButton
              recommended={RECOMMENDED_RAISE}
              onClick={onIncrease}
            />
          ) : undefined
        }
      >
        <BrandHeader />
        <BalanceHero
          balance={balance}
          initialBudget={INITIAL_BUDGET}
          monthlyCount={monthlyCount}
          dead={dead}
        />

        <MetricsPanel
          damageCount={monthlyCount}
          consumptionRate={(INITIAL_BUDGET - balance) / INITIAL_BUDGET}
          amountUsed={INITIAL_BUDGET - balance}
        />

        <GoroButton
          state={effectiveState}
          mainLabel={mainLabel}
          subLabel={subLabel}
          onClick={onPress}
          winningKey={
            effectiveState === 'slot' && slotResult ? slotResult.shortName : undefined
          }
        >
          {effectiveState === 'slot' && <SlotReel result={slotResult} />}
        </GoroButton>

        {dead && <DeadVerdict />}
      </ScreenFrame>
    </>
  );
}
