// Amplify Auth Hub のリスナー (NFR Design LC-AUTH-14)
// tokenRefresh_failure を捕まえて triggerSessionExpired() を呼ぶ。

import { Hub } from "aws-amplify/utils";
import { triggerSessionExpired } from "./apiClient";

let registered = false;

/**
 * アプリ起動時に 1 度だけ呼ぶ。AppProviders で実行する想定。
 */
export function setupAuthHubListener(): void {
  if (registered) return;
  registered = true;

  Hub.listen("auth", (capsule) => {
    const event = capsule?.payload?.event;
    if (event === "tokenRefresh_failure") {
      triggerSessionExpired();
    }
  });
}
