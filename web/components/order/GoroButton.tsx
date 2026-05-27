// LC-21 GoroButton (+ LC-SUGGEST-11 連携)
//
// メイン画面の「めんどくさい」ボタン。useOrder hook 経由で PlaceOrder mutation を発火。
// Unit D: useSuggestion で先回り提案がある場合は suggestionAtom に流し込み、
// SuggestBubble をボタン上部に出して suggested 状態へ遷移する。
"use client";

import { useRouter } from "next/navigation";
import { useAtomValue, useSetAtom } from "jotai";
import { useEffect, useState } from "react";

import { useAuth } from "@/hooks/useAuth";
import { useOrder } from "@/hooks/useOrder";
import { useSuggestion } from "@/hooks/useSuggestion";
import { generateUlid } from "@/lib/ulid";
import { COPY, composeSuggestSubLabel } from "@/lib/copy";
import {
  balanceAtom,
  monthlyCountAtom,
  screenStateAtom,
  suggestionAtom,
  type ScreenState,
} from "@/state/main";

import { SlotReel } from "./SlotReel";
import { SuggestBubble } from "./SuggestBubble";

import styles from "./GoroButton.module.css";

const SLOT_MIN_DURATION_MS = 1200;

export function GoroButton() {
  const router = useRouter();
  const screenState = useAtomValue(screenStateAtom);
  const setScreenState = useSetAtom(screenStateAtom);
  const setBalance = useSetAtom(balanceAtom);
  const setMonthlyCount = useSetAtom(monthlyCountAtom);
  const suggestion = useAtomValue(suggestionAtom);
  const setSuggestion = useSetAtom(suggestionAtom);
  const { mutate, disabled, isPending } = useOrder();
  void isPending;
  const { user } = useAuth();

  const { suggestion: suggestionResponse } = useSuggestion();
  const plan = suggestionResponse?.hasSuggestion ? suggestionResponse.plan : undefined;

  // Unit D: useSuggestion 結果を suggestionAtom に流し込み、状態を suggested へ遷移
  useEffect(() => {
    if (!suggestionResponse?.hasSuggestion || !suggestionResponse.plan) {
      return;
    }
    setSuggestion({
      storeName: suggestionResponse.plan.storeName,
      amount: suggestionResponse.plan.amount,
      suggestionId: suggestionResponse.suggestionId ?? "",
    });
    setScreenState((current) => (current === "idle" ? "suggested" : current));
  }, [suggestionResponse, setSuggestion, setScreenState]);

  const [winningText, setWinningText] = useState<string | null>(null);
  const [slotStartedAt, setSlotStartedAt] = useState<number>(0);

  useEffect(() => {
    if (screenState !== "slot") {
      setWinningText(null);
    }
  }, [screenState]);

  const handleClick = () => {
    if (screenState !== "idle" && screenState !== "suggested") return;
    if (disabled) return;
    // VR-B-04/05: idempotencyKey は `${userId}:${ulid}` 形式が必須 (cross-user
    // idempotency replay 防止)。useAuth が確定していない瞬間のクリックは no-op。
    if (!user?.userId) return;

    setScreenState("slot");
    setSlotStartedAt(Date.now());
    setWinningText(null);

    mutate(
      {
        category: "food",
        idempotencyKey: `${user.userId}:${generateUlid()}`,
        // BR-D13/BR-C10: suggested 表示中なら suggestionId を送る (Unit C が解決)
        suggestionId: suggestion?.suggestionId || undefined,
      },
      {
        onSuccess: (res) => {
          const elapsed = Date.now() - slotStartedAt;
          const remaining = Math.max(0, SLOT_MIN_DURATION_MS - elapsed);
          setTimeout(() => {
            setWinningText(`${res.storeName} ¥${res.amount.toLocaleString("ja-JP")}`);
            setBalance(res.remainingBalance);
            setMonthlyCount((c) => c + 1);
            setTimeout(() => {
              setScreenState(res.remainingBalance === 0 ? "dead" : "idle");
              router.push(`/order/${res.orderId}/complete`);
            }, 800);
          }, remaining);
        },
        onError: () => {
          setScreenState("idle");
        },
      },
    );
  };

  const labelMain = labelMainFor(screenState);
  const labelSub = labelSubFor(screenState, suggestion);
  const ariaLabel = ariaLabelFor(screenState, suggestion);
  // winning 時に button-jolt + win-flare-burst を 1 度だけ走らせる。
  // key 切り替えで <button> を remount して animation を再発火させる方式 (mock 流)。
  const winning = winningText !== null;

  return (
    <div className={styles.wrap}>
      <button
        key={winning ? `win-${winningText}` : "btn"}
        type="button"
        onClick={handleClick}
        disabled={disabled || screenState === "slot" || screenState === "dead"}
        data-testid="goro-button"
        data-state={screenState}
        data-winning={winning}
        aria-label={ariaLabel}
        className={styles.button}
      >
        {screenState === "suggested" && <SuggestBubble plan={plan} />}
        {screenState === "slot" ? (
          <SlotReel winningText={winningText} />
        ) : (
          <>
            <span className={styles.labelMain}>{labelMain}</span>
            {labelSub && <span className={styles.labelSub}>{labelSub}</span>}
          </>
        )}
      </button>
      {winning && <span className={styles.flare} aria-hidden="true" />}
    </div>
  );
}

function labelMainFor(state: ScreenState): string {
  if (state === "suggested") return COPY.main.suggestMain;
  return COPY.main.idleMain;
}

function labelSubFor(
  state: ScreenState,
  suggestion: { storeName: string; amount: number } | null,
): string | null {
  if (state === "dead") return COPY.main.deadButtonSub;
  if (state === "slot") return null;
  if (state === "suggested" && suggestion) {
    return composeSuggestSubLabel(suggestion.storeName, suggestion.amount);
  }
  return COPY.main.idleSub;
}

function ariaLabelFor(
  state: ScreenState,
  suggestion: { storeName: string; amount: number } | null,
): string {
  if (state === "dead") return "残高不足のため注文できません";
  if (state === "slot") return "注文処理中";
  if (state === "suggested" && suggestion) {
    return `${suggestion.storeName} ¥${suggestion.amount.toLocaleString("ja-JP")} を注文`;
  }
  return "ご飯めんどくさい";
}
