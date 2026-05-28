import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenFrame from '../components/ScreenFrame';
import BrandHeader from '../components/BrandHeader';
import BalanceHero from '../components/BalanceHero';
import GoroButton, { GoroButtonState } from '../components/GoroButton';
import SlotReel from '../components/SlotReel';
import DeadVerdict from '../components/DeadVerdict';
import IncreaseBudgetButton from '../components/IncreaseBudgetButton';
import { COPY } from '../lib/copy';
import { OrderResult, placeOrder, suggestNext } from '../lib/fakeApi';
import { playWinChime, setSoundEnabled } from '../lib/sound';
import styles from './MainMock.module.css';
import insufStyles from './InsufficientBalanceModalMock.module.css';
import raiseStyles from './RaiseBudgetModalMock.module.css';

const INITIAL_BUDGET = 30_000;
const RECOMMENDED_RAISE = 50_000;
// 'insufficient' デモ用の低残高。次の注文 (>1,000 円) で確実に不足する値。
const INSUFFICIENT_BALANCE = 500;

type DebugMode = 'auto' | 'idle' | 'suggested' | 'insufficient' | 'dead';
const DEBUG_MODES: DebugMode[] = ['auto', 'idle', 'suggested', 'insufficient', 'dead'];

type ModalView = 'none' | 'insufficient' | 'raise';

export default function MainMock() {
  const navigate = useNavigate();

  const [balance, setBalance] = useState(INITIAL_BUDGET);
  const [monthlyCount, setMonthlyCount] = useState(0);
  const [state, setState] = useState<GoroButtonState>('idle');
  const [slotResult, setSlotResult] = useState<OrderResult | undefined>();
  const [debug, setDebug] = useState<DebugMode>('auto');
  const [sound, setSound] = useState(false);
  const [winFlash, setWinFlash] = useState(0);
  const [modalView, setModalView] = useState<ModalView>('none');

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

    // 予算不足の演出: 次の最小注文額 (1,000 円) を満たさないなら、
    // スロットを回さずに InsufficientBalance モーダルを直接出す。
    if (balance < 1_000) {
      setModalView('insufficient');
      return;
    }

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
    setModalView('none');
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
    } else if (mode === 'insufficient') {
      // 低残高で idle にしておき、ボタンを押すと InsufficientBalance モーダルが出る導線
      setBalance(INSUFFICIENT_BALANCE);
      setMonthlyCount(24);
      setState('idle');
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

  // InsufficientBalance モーダル: primary で Raise モーダルに進む。
  const onInsufficientPrimary = () => setModalView('raise');
  const onInsufficientClose = () => setModalView('none');

  // Raise モーダル: primary で残高を増額して両モーダルを閉じる。
  const onRaiseConfirm = () => {
    setBalance((b) => b + RECOMMENDED_RAISE);
    setModalView('none');
  };
  const onRaiseCancel = () => setModalView('none');

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

      {/* 予算不足: 注文時に balance < 1,000 で出る (InsufficientBalanceModalMock と同じスタイル) */}
      {modalView === 'insufficient' && (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="insufficient-balance-title-inline"
          className={insufStyles.overlay}
        >
          <div className={insufStyles.card}>
            <h2
              id="insufficient-balance-title-inline"
              className={insufStyles.h2}
            >
              {COPY.insufficient.h2}
            </h2>
            <p className={insufStyles.sub}>
              {COPY.insufficient.bodyLine1}
              <br />
              {COPY.insufficient.bodyLine2}
            </p>
            <div className={insufStyles.actions}>
              <button
                onClick={onInsufficientClose}
                className={insufStyles.secondary}
              >
                {COPY.insufficient.secondary}
              </button>
              <button
                onClick={onInsufficientPrimary}
                className={insufStyles.primary}
              >
                {COPY.insufficient.primary}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 増額: Insufficient モーダル primary から遷移 (RaiseBudgetModalMock と同じスタイル) */}
      {modalView === 'raise' && (
        <div className={raiseStyles.overlay}>
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="raise-modal-title-inline"
            className={raiseStyles.card}
          >
            <h2 id="raise-modal-title-inline" className={raiseStyles.h2}>
              {COPY.raiseBudget.h2}
            </h2>

            <p className={raiseStyles.recommend}>
              {COPY.raiseBudget.recommendLabel}
              <span className={raiseStyles.recommendAmount}>
                ¥{RECOMMENDED_RAISE.toLocaleString()}
              </span>
              <span className={raiseStyles.recommendNote}>
                {COPY.raiseBudget.recommendNote}
              </span>
            </p>

            <div className={raiseStyles.actions}>
              <button
                onClick={onRaiseConfirm}
                className={raiseStyles.primary}
              >
                {COPY.raiseBudget.primary}
              </button>
              <button
                onClick={onRaiseCancel}
                className={raiseStyles.secondary}
              >
                {COPY.raiseBudget.secondary}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
