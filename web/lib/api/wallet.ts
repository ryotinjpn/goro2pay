// Unit B Wallet API ラッパ (LC-AUTH-09 apiClient.request を経由)。
//
// 凍結 IF (unit-interfaces.md §3.3):
//   - GET /api/wallet → { balance, monthlyBudget, updatedAt }
//   - POST /api/wallet/budget → { monthlyBudget, appliedFrom }
//
// エラーは ApiError (orders.ts と統一) を throw する。

import { apiClient } from "../apiClient";
import { ApiError } from "./orders";

export type WalletResponse = {
  balance: number;
  monthlyBudget: number;
  updatedAt: string;
};

export type SetBudgetResponse = {
  monthlyBudget: number;
  appliedFrom: string;
};

type ErrorResponseBody = {
  code?: string;
  message?: string;
};

async function parseError(res: Response): Promise<ApiError> {
  let body: ErrorResponseBody = {};
  try {
    body = (await res.json()) as ErrorResponseBody;
  } catch {
    /* JSON でない応答は status code のみで判定 */
  }
  const code = body.code ?? `HTTP_${res.status}`;
  const message = body.message ?? res.statusText;
  return new ApiError(res.status, code, message);
}

/**
 * getWallet は GET /api/wallet を呼び出す。
 *
 * 成功時は WalletResponse を返し、HTTP エラーは ApiError を throw する。
 * 404 (WALLET_NOT_FOUND) は ApiError として伝播 — 呼び出し側が `/budget` への
 * リダイレクトを判定する。
 */
export async function getWallet(): Promise<WalletResponse> {
  const res = await apiClient.request({
    path: "/api/wallet",
    method: "GET",
  });
  if (!res.ok) {
    throw await parseError(res);
  }
  return (await res.json()) as WalletResponse;
}

/**
 * postWalletBudget は POST /api/wallet/budget を呼び出す。
 *
 * 成功時は SetBudgetResponse を返し、HTTP エラーは ApiError を throw する。
 * VR-B-01 / VR-B-02 違反 (400 VALIDATION_FAILED) は ApiError として伝播。
 */
export async function postWalletBudget(
  monthlyBudget: number,
): Promise<SetBudgetResponse> {
  const res = await apiClient.request({
    path: "/api/wallet/budget",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ monthlyBudget }),
  });
  if (!res.ok) {
    throw await parseError(res);
  }
  return (await res.json()) as SetBudgetResponse;
}
