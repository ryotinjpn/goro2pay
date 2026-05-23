// LC-27 ToastVariants / P-FE-TOAST-01
//
// 自虐トースト 3 種ローテーション (NFR-DEG-05 体現、NFRC-C22)。
// 純関数で実装することで Vitest からの `vi.spyOn(Math, 'random')` で
// 決定論的にテストできる。

/**
 * トースト文言のラインナップ (3 種)。
 *
 * 最終文言は NFR Design Q-D10 で例示されたものを採用。本番化時に
 * NFR-DEG-05 文言ガイドラインに従って増減可能。
 */
export const TOAST_VARIANTS = [
  "サーバーがやる気を失いました…もう一度お試しください",
  "システムがふぬけてます。少し待ってあげてください",
  "今日はちょっとダメ化に失敗しました。再挑戦しますか？",
] as const;

export type ToastVariant = (typeof TOAST_VARIANTS)[number];

/**
 * Math.random でランダムに 1 文言を選んで返す。
 *
 * 連続避けロジックは採用しない (Q-D10=A 採用根拠: NFRC-C22 の
 * 連打抑制 1 秒で連続表示は構造的に発生しないため)。
 */
export function getRandomToast(): ToastVariant {
  const index = Math.floor(Math.random() * TOAST_VARIANTS.length);
  return TOAST_VARIANTS[index];
}
