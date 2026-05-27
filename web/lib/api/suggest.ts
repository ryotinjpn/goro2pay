// LC-SUGGEST (frontend) ApiClientSuggest
//
// Suggest 系エンドポイント (/api/suggest) の Browser 側ラッパ。
// 既存 LC-AUTH-09 apiClient.request を経由し、Authorization: Bearer 透過 (BFF)。

import { apiClient } from "../apiClient";

/** SuggestionPlan は提案内容 (凍結契約 §5.2)。 */
export type SuggestionPlan = {
  storeName: string;
  menuName: string;
  amount: number;
  category: string;
};

/** SuggestionResponse は GET /api/suggest のレスポンス (凍結契約 §5.2)。 */
export type SuggestionResponse = {
  hasSuggestion: boolean;
  suggestionId?: string;
  title?: string;
  plan?: SuggestionPlan;
};

/**
 * fetchSuggestion は GET /api/suggest を呼び出す。
 *
 * サジェストはメイン機能 (注文) を阻害しないため、HTTP エラー・JSON 不正時は
 * throw せず `{ hasSuggestion: false }` に丸める (NFRD-D17 / BR-D02)。
 */
export async function fetchSuggestion(): Promise<SuggestionResponse> {
  const res = await apiClient.request({
    path: "/api/suggest",
    method: "GET",
  });
  if (!res.ok) {
    return { hasSuggestion: false };
  }
  try {
    return (await res.json()) as SuggestionResponse;
  } catch {
    return { hasSuggestion: false };
  }
}
