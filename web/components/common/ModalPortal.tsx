// ModalPortal: モーダルを <body> 直下に portal 出力する共通ラッパ。
//
// なぜ必要か:
// ScreenFrame (`web/components/order/ScreenFrame.module.css`) は
//   1) `isolation: isolate` で stacking context を作る
//   2) `.frame > * { z-index: 1 }` で全直接子に z-index 強制
// この組み合わせにより、ScreenFrame 配下で render されるモーダルは
// `position: fixed; z-index: 1100` を指定しても **frame の stacking context 内**
// に閉じ込められ、frame の sibling である winFlash / winVerdict (z-index 80/90)
// より結果的に下に来て、ボタンが押せなくなる現象が発生していた (2026-05-28 報告)。
//
// 解決策: createPortal で <body> 直下に出力すれば stacking context の縛りから
// 完全に外れ、document 全体の z-index 順序がそのまま効く。
//
// SSR ガード:
// Next.js App Router では本コンポーネントが Client Component として render される
// 前に SSR が走る場合があるため、`document` が無い環境では null を返す ("use client"
// + マウント検出) ことで hydration mismatch を防ぐ。
"use client";

import { useEffect, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

type Props = {
  children: ReactNode;
};

export function ModalPortal({ children }: Props) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;
  return createPortal(children, document.body);
}
