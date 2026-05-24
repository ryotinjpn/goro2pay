// LC-26 OrderErrorMapper / P-FE-ERR-01
//
// HTTP エラー / ネットワークエラーから OrderErrorAction への純関数マッピング。
// Vitest 単体テスト容易性のため、router / toast 等 UI フレームワークには依存しない。
// useOrder hook が Action.type で switch し、router.push / showToast を呼ぶ。
//
// レビュー指摘 (F-C1) 反映: apiClient.request は 401/429/500+ を AuthErrorWithCode
// として throw するため、ApiError だけ判定すると 401-silent / 500-toast 分岐が
// production で到達しない。AuthErrorWithCode の code でも分岐を行う。
//
// レビュー指摘 (F-I4) 反映: 409 (BR-C39 冪等性衝突 = 実質成功) は silent では
// なく "履歴 invalidate + 自虐トースト" に変更。orderId が body に含まれない
// 現状 API では完了画面遷移は不可能だが、履歴更新 (refresh: true) でユーザが
// "もう注文済みだった" 状態を確認できる。
import { ApiError } from "./api/orders";
import { AuthErrorWithCode } from "./authMessages";
import { getRandomToast } from "./toasts";

/**
 * OrderErrorAction は mapOrderError の戻り値で、useOrder hook が解釈する。
 *
 *  - navigate: BudgetEmptyScreen 等への遷移 (transitionMs=0 で即時遷移、NFRC-C22)
 *  - toast   : 自虐トースト表示 (durationMs=5000 で 5 秒自動消去)
 *  - silent  : 何もしない (UI ハンドリング不要、auth hub listener が処理する 401 等)
 *
 * `refresh` フィールド (F-I4 修正): true のとき useOrder.onError 内で
 *   invalidateQueries(['orderHistory']) と (['balance']) を呼ぶ。
 *   主に 409 冪等性衝突時の履歴反映用。
 */
export type OrderErrorAction =
  | { type: "navigate"; path: string; transitionMs?: number; refresh?: boolean }
  | { type: "toast"; text: string; durationMs?: number; refresh?: boolean }
  | { type: "silent"; refresh?: boolean };

/**
 * mapOrderError は HTTP エラーを OrderErrorAction にマッピングする純関数。
 *
 * 凍結契約 §4.3 のエラー code を尊重しつつ、エンドユーザ向けの文言生成は
 * Frontend 側で完結させる (Backend message 文字列に依存しない)。
 */
export function mapOrderError(err: unknown): OrderErrorAction {
  // F-C1 修正: apiClient.request が AuthErrorWithCode を投げた場合の分岐
  if (err instanceof AuthErrorWithCode) {
    switch (err.code) {
      case "SESSION_EXPIRED":
        // SessionExpiredModalHost が atom 経由で modal を出すので silent
        return { type: "silent" };
      case "RATE_LIMIT_EXCEEDED":
      case "NETWORK_ERROR":
      default:
        return { type: "toast", text: getRandomToast(), durationMs: 5000 };
    }
  }

  if (err instanceof ApiError) {
    switch (err.status) {
      case 402:
        return { type: "navigate", path: "/budget-empty", transitionMs: 0 };
      case 409:
        // F-I4 修正: 冪等性衝突 (BR-C39) は実質成功扱い。完了画面に行きたいが
        // orderId が body に含まれないため、履歴 invalidate + 自虐風トーストで代替。
        return {
          type: "toast",
          text: "もう注文済みでした…まあ、ダメ化中ですからね",
          durationMs: 5000,
          refresh: true,
        };
      case 503:
        // B-C2 連動: Unit B 未配線時の SERVICE_UNAVAILABLE
        return { type: "toast", text: getRandomToast(), durationMs: 5000 };
      default:
        return { type: "toast", text: getRandomToast(), durationMs: 5000 };
    }
  }

  // ネットワークエラー / その他は自虐トースト
  return { type: "toast", text: getRandomToast(), durationMs: 5000 };
}
