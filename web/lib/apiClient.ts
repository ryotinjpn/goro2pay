// Browser 側 apiClient (BFF パターン Browser 側、LC-AUTH-09)
//
// すべての /api/* リクエストはこの apiClient.request() 経由とする。
// 1. fetchAuthSession() で AccessToken を取得 (OAuth2 ベストプラクティス、IdToken ではない)
// 2. Authorization: Bearer <accessToken> ヘッダで同一オリジンの /api/* を叩く
//    (Next.js catch-all Route Handler が API Gateway に proxy)
// 3. 401 を検出したら sessionExpiredAtom を起動 (二重発火防止は atom の compare-and-set)

import { fetchAuthSession } from "aws-amplify/auth";
import { getDefaultStore } from "jotai";

import { sessionExpiredAtom } from "@/state/auth";
import { AuthErrorWithCode } from "@/lib/authMessages";

export type ApiClientRequestInit = RequestInit & { path: string };

// triggerSessionExpired は SessionExpired modal の起動を 1 度だけにするための
// gate。jotai atom の compare-and-set のみで二重発火を防ぐ (モジュールフラグは
// 不要、atom が "(null → 値)" 遷移を 1 回しか許さない)。
export function triggerSessionExpired(): void {
  const store = getDefaultStore();
  if (store.get(sessionExpiredAtom) !== null) return;
  store.set(sessionExpiredAtom, { openedAt: Date.now() });
}

// テスト用: atom を null に戻す。本番では SessionExpiredModalHost が
// navigateToLogin() の中で同じことをする。
export function _resetSessionExpiredHandling(): void {
  getDefaultStore().set(sessionExpiredAtom, null);
}

export const apiClient = {
  async request(input: ApiClientRequestInit): Promise<Response> {
    // Authorization の優先順:
    //   (1) input.headers.Authorization が呼出側で明示されていればそれを優先
    //       (テスト時の手動注入や、特殊ケース用)
    //   (2) fetchAuthSession() から AccessToken を取得して付与 (本番経路)
    //   (3) fetchAuthSession 自体が throw した場合は NETWORK_ERROR
    //       (オフラインや Cognito 接続障害を SESSION_EXPIRED と誤分類しない)
    const headers = new Headers(input.headers);

    if (!headers.has("Authorization")) {
      let accessToken: string | undefined;
      try {
        const session = await fetchAuthSession();
        accessToken = session.tokens?.accessToken?.toString();
      } catch {
        // fetchAuthSession の throw はネットワーク全断や Cognito 接続障害が主因。
        // SESSION_EXPIRED modal を出してログイン画面に飛ばすのは UX 的に誤り
        // (ログインしていたのに切れたと誤認させる) なので NETWORK_ERROR に分類。
        throw new AuthErrorWithCode("NETWORK_ERROR");
      }
      if (accessToken) {
        headers.set("Authorization", `Bearer ${accessToken}`);
      }
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
