// Browser 側 apiClient (BFF パターン Browser 側、LC-AUTH-09)
//
// すべての /api/* リクエストはこの apiClient.request() 経由とする。
// 1. fetchAuthSession() で AccessToken を取得 (OAuth2 ベストプラクティス、IdToken ではない)
// 2. Authorization: Bearer <accessToken> ヘッダで同一オリジンの /api/* を叩く
//    (Next.js catch-all Route Handler が API Gateway に proxy)
// 3. 401 を検出したら sessionExpiredAtom を起動 (二重発火防止は atom の atomic write)

import { fetchAuthSession } from "aws-amplify/auth";
import { getDefaultStore } from "jotai";

import { sessionExpiredAtom } from "@/state/auth";
import { AuthErrorWithCode } from "@/lib/authMessages";

export type ApiClientRequestInit = RequestInit & { path: string };

let isHandlingSessionExpired = false;

export function triggerSessionExpired(): void {
  if (isHandlingSessionExpired) return;
  isHandlingSessionExpired = true;
  const store = getDefaultStore();
  // atom が null の場合のみ書き込む (二重発火防止)
  const current = store.get(sessionExpiredAtom);
  if (current === null) {
    store.set(sessionExpiredAtom, { openedAt: Date.now() });
  }
}

// テスト用: 内部フラグをリセット
export function _resetSessionExpiredHandling(): void {
  isHandlingSessionExpired = false;
}

export const apiClient = {
  async request(input: ApiClientRequestInit): Promise<Response> {
    let accessToken: string | undefined;
    try {
      const session = await fetchAuthSession();
      accessToken = session.tokens?.accessToken?.toString();
    } catch {
      // session 取得失敗は 401 と同等扱い
    }

    const headers = new Headers(input.headers);
    if (accessToken) {
      headers.set("Authorization", `Bearer ${accessToken}`);
    }

    // path は RequestInit に存在しないキー。fetch 第 2 引数に紛れ込ませると
    // 仕様外プロパティの混入になり、strict な TypeScript 設定で破綻するため、
    // destructure で path を切り出してから残りを渡す。headers は新しい
    // Headers インスタンスを優先するので元の headers も除外しておく。
    const { path, headers: _omit, ...init } = input;
    void _omit;
    const res = await fetch(path, {
      ...init,
      headers,
    });

    if (res.status === 401) {
      triggerSessionExpired();
      throw new AuthErrorWithCode("SESSION_EXPIRED");
    }
    if (res.status === 429) {
      throw new AuthErrorWithCode("RATE_LIMIT_EXCEEDED");
    }
    if (res.status >= 500) {
      throw new AuthErrorWithCode("NETWORK_ERROR");
    }
    return res;
  },
};
