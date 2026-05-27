"use client";

import { COPY } from "@/lib/copy";

import styles from "./SlotReel.module.css";

const DUMMY_CANDIDATES = [
  "CoCo壱番屋",
  "すき家",
  "ロイヤルホスト",
  "松屋",
  "吉野家",
  "サブウェイ",
];

type Props = {
  winningText: string | null;
};

export function SlotReel({ winningText }: Props) {
  const stopped = winningText !== null;
  const items = stopped ? [winningText] : DUMMY_CANDIDATES;

  return (
    <div className={styles.reelWrap} data-testid="slot-reel">
      <div className={styles.window}>
        <div
          className={styles.strip}
          data-stopped={stopped}
          aria-hidden="true"
        >
          {[...items, ...items].map((item, i) => (
            <div key={i}>{item}</div>
          ))}
        </div>
      </div>
      <div className={styles.caption}>{COPY.main.slotCaption}</div>
      {stopped && winningText && (
        <span className="sr-only">注文確定: {winningText}</span>
      )}
    </div>
  );
}
