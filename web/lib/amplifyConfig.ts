// Amplify Auth v6 の設定オブジェクト + module-level configure ガード。
// NEXT_PUBLIC_* env から Cognito の値を読み出す。
// Browser に露出してよい (Public な値)。
//
// API_ENDPOINT は server-only env なのでここでは使わない (BFF パターン)。

import { Amplify } from "aws-amplify";
import type { ResourcesConfig } from "aws-amplify";

export const amplifyConfig: ResourcesConfig = {
  Auth: {
    Cognito: {
      userPoolId: process.env.NEXT_PUBLIC_USER_POOL_ID ?? "",
      userPoolClientId: process.env.NEXT_PUBLIC_USER_POOL_CLIENT_ID ?? "",
      // Region は Amplify が userPoolId からも判定するが明示しておく
      // signUpVerificationMethod は auto-confirm のため不要
    },
  },
};

// Amplify.configure はモジュールスコープ singleton を書き換えるため、
// プロセスで 1 度だけ呼ぶのが正解。React の render 内で呼ぶと StrictMode の
// 二重 render や複数 mount 経路で複数回呼ばれるリスクがある。
// 本ファイル module-level の configured フラグで完全にガードする。
//
// `ssr: true` は @aws-amplify/adapter-nextjs の cookieStorage 設定とセットで
// はじめて意味を持つ。本 MVP では adapter-nextjs を導入しておらず、
// Token は localStorage に置かれているため `ssr: false` (= 無指定でも同じ)
// に揃えて中途半端な状態を解消する。Phase 2 で adapter-nextjs を導入する
// 際に `ssr: true` + cookieStorage の組合せに切替予定。
let configured = false;

export function ensureAmplifyConfigured(): void {
  if (configured) return;
  Amplify.configure(amplifyConfig, { ssr: false });
  configured = true;
}
