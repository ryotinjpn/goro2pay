// Amplify Auth Hub のリスナー (NFR Design LC-AUTH-14)
// tokenRefresh_failure を捕まえて triggerSessionExpired() を呼ぶ。

import { Hub } from "aws-amplify/utils";
import { triggerSessionExpired } from "./apiClient";

let registered = false;
let unsubscribe: (() => void) | null = null;

/**
 * アプリ起動時に 1 度だけ呼ぶ。AppProviders で実行する想定。
 *
 * 戻り値: unsubscribe 関数。AppProviders unmount や HMR reload 時に
 * 呼び出して listener を解除できる。MVP では unmount しないため通常は
 * 呼ばないが、HMR で本ファイル module が再評価された際は registered フラグも
 * reset されるため、`teardownAuthHubListener` を呼んで二重登録を避ける。
 */
export function setupAuthHubListener(): () => void {
  if (registered && unsubscribe) {
    return unsubscribe;
  }
  registered = true;

  unsubscribe = Hub.listen("auth", (capsule) => {
    const event = capsule?.payload?.event;
    if (event === "tokenRefresh_failure") {
      triggerSessionExpired();
    }
  });

  return unsubscribe;
}

/**
 * テスト用 / HMR 用: listener を解除して registered フラグを reset。
 */
export function teardownAuthHubListener(): void {
  if (unsubscribe) {
    unsubscribe();
    unsubscribe = null;
  }
  registered = false;
}
