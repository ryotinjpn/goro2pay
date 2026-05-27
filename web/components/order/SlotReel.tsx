// SlotReel: mock/src/components/SlotReel.tsx ベースに移植 (PR ②)。
// web の API 契約 (winningText: string | null) は維持しつつ、内部実装を mock の
// 縦 flex リール + scanline + flash 演出に揃える。
//
// SLOT_LINES の長さは SlotReel.module.css の `@keyframes slot-spin` 終端値
// (-240px) と整合させること: 各 line の height 48px × 5 行 = 240px で 1 周。
// 行を増減したら CSS 側の終端値も同時に更新する。
//
// 当選金額バブル (mock の `.amount`) は本 PR では未実装。PR ⑧ で `lastOrderAtom`
// 経由で MainScreen 側に表示する予定 (本コンポーネントの API は変えない)。
"use client";

import { COPY } from "@/lib/copy";

import styles from "./SlotReel.module.css";

const SLOT_LINES = [
  "ダメ食堂",
  "ゴロネ亭",
  "よこなり弁当",
  "もちかえり食堂",
  "ぐうたらカレー",
];

type Props = {
  winningText: string | null;
};

export function SlotReel({ winningText }: Props) {
  const stopped = winningText !== null;
  // 停止時は当選店名 1 行、回転中は 5 候補 + 先頭再掲で見た目をループさせる。
  const lines = stopped ? [winningText] : [...SLOT_LINES, SLOT_LINES[0]];
  // window に `key` で remount をかけて、停止時の slot-frame-flash を再走させる。
  const flashKey = stopped ? `flash-${winningText}` : "spinning";

  return (
    <div className={styles.reelWrap} data-testid="slot-reel">
      <div
        key={flashKey}
        className={styles.window}
        data-flash={stopped}
        aria-hidden="true"
      >
        <div className={styles.strip} data-stopped={stopped}>
          {lines.map((s, i) => (
            <div key={i} className={styles.line}>
              {s}
            </div>
          ))}
        </div>
        <div className={styles.scanline} />
      </div>
      <div className={styles.caption}>{COPY.main.slotCaption}</div>
      {stopped && winningText && (
        <span className="sr-only">注文確定: {winningText}</span>
      )}
    </div>
  );
}
