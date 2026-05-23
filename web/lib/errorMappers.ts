// LC-26 OrderErrorMapper / P-FE-ERR-01
//
// HTTP エラー / ネットワークエラーから OrderErrorAction への純関数マッピング。
// Vitest 単体テスト容易性のため、router / toast 等 UI フレームワークには依存しない。
// useOrder hook が Action.type で switch し、router.push / showToast を呼ぶ。
import { ApiError } from "./api/orders";
import { getRandomToast } from "./toasts";

/**
 * OrderErrorAction は mapOrderError の戻り値で、useOrder hook が解釈する。
 *
 *  - navigate: BudgetEmptyScreen 等への遷移 (transitionMs=0 で即時遷移、NFRC-C22)
 *  - toast   : 自虐トースト表示 (durationMs=5000 で 5 秒自動消去)
 *  - silent  : 何もしない (冪等性衝突 409 は実質成功扱い、BR-C39)
 */
export type OrderErrorAction =
  | { type: "navigate"; path: string; transitionMs?: number }
  | { type: "toast"; text: string; durationMs?: number }
  | { type: "silent" };

/**
 * mapOrderError は HTTP エラーを OrderErrorAction にマッピングする純関数。
 *
 * 凍結契約 §4.3 のエラー code を尊重しつつ、エンドユーザ向けの文言生成は
 * Frontend 側で完結させる (Backend message 文字列に依存しない)。
 */
export function mapOrderError(err: unknown): OrderErrorAction {
  if (err instanceof ApiError) {
    switch (err.status) {
      case 402:
        return { type: "navigate", path: "/budget-empty", transitionMs: 0 };
      case 409:
        // 冪等性衝突は実質成功扱い (BR-C39、Idempotent=true で別途 onSuccess)
        return { type: "silent" };
      case 401:
        // session expired は Unit A の hub listener が拾うので silent で良い
        return { type: "silent" };
      default:
        return { type: "toast", text: getRandomToast(), durationMs: 5000 };
    }
  }

  // ネットワークエラー / その他は自虐トースト
  return { type: "toast", text: getRandomToast(), durationMs: 5000 };
}
