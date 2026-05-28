// BrandHeader: mock/src/components/BrandHeader.tsx ベースに移植 (PR ⑤)。
// 左に「ゴロゴロPay」、右に時刻 + ログアウトボタン (web 独自) を並べる。
//
// LogoutButton の出し分けに注意 (PR ⑤ レビュー指摘 Critical 1):
//   - BrandHeader は MainScreen / Complete / Budget だけでなく Landing / Login /
//     Signup でも使われる。後者は未認証画面なので Logout を出してはいけない
//   - 呼び出し側で `showLogout` prop を明示し、さらに useAuth().status を
//     見て二重ガードする (signedOut Hub の反映タイミング差で showLogout=true
//     のまま未認証になる短い瞬間でも誤表示しない)
//
// テーマ:
//   - `light` prop で Light Aurora 表現 (白基調 + ink テキスト) に切り替える。
//     未指定はダーク (legacy)。段階移行中の混在を許容する想定だが、現状の web は
//     全画面で light を渡している。
"use client";

import { useEffect, useState } from "react";

import { LogoutButton } from "@/components/auth/LogoutButton";
import { useAuth } from "@/hooks/useAuth";
import { COPY } from "@/lib/copy";

import styles from "./BrandHeader.module.css";

function formatTime(d: Date): string {
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
}

type Props = {
  /** Light Aurora テーマで描画 (移行済み画面のみ true)。未指定は legacy ダーク。 */
  light?: boolean;
  /**
   * 認証必須画面 (MainScreen / Complete / Budget) からのみ true で渡す。
   * Landing / Login / Signup から渡されないと undefined → false で出ない。
   */
  showLogout?: boolean;
};

export function BrandHeader({ light, showLogout = false }: Props = {}) {
  const [time, setTime] = useState<string>(() => formatTime(new Date()));
  const { status } = useAuth();

  useEffect(() => {
    // 60s 間隔で十分 (1 分単位の表示)。実際のラグは最大 60s だが、ユーザは
    // ヘッダ時計をストップウォッチとして使わないので問題にならない。
    const id = setInterval(() => setTime(formatTime(new Date())), 60_000);
    return () => clearInterval(id);
  }, []);

  const name = COPY.brand.name;
  const splitAt = name.indexOf("Pay");
  const left = splitAt >= 0 ? name.slice(0, splitAt) : name;
  const right = splitAt >= 0 ? name.slice(splitAt) : "";

  // showLogout が true でも、useAuth が authenticated でない場合は出さない。
  // (signOut Hub 反映の遅延中に誤表示しないためのガード)
  const renderLogout = showLogout && status === "authenticated";

  return (
    <header className={styles.header} data-theme={light ? "light" : undefined}>
      <span className={styles.brand}>
        <span className={styles.brandKana}>{left}</span>
        {right && <span className={styles.brandLatin}>{right}</span>}
      </span>
      <div className={styles.right}>
        <span className={styles.time}>{time}</span>
        {renderLogout && <LogoutButton className={styles.logoutBtn} />}
      </div>
    </header>
  );
}
