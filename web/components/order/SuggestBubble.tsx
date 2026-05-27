"use client";

import type { SuggestionPlan } from "@/lib/api/suggest";
import { COPY, composeSuggestSubLabel } from "@/lib/copy";

import styles from "./SuggestBubble.module.css";

type Props = {
  plan?: SuggestionPlan;
};

export function SuggestBubble({ plan }: Props) {
  return (
    <div className={styles.bubble} data-testid="suggest-bubble" aria-hidden="true">
      <div className={styles.title}>{COPY.main.suggestBubble}</div>
      {plan && (
        <div data-testid="suggest-plan-label" className={styles.plan}>
          {composeSuggestSubLabel(plan.storeName, plan.amount)}
        </div>
      )}
    </div>
  );
}
