// LC-33 ApiClientOrders
//
// Order 系エンドポイント (/api/orders) の Browser 側ラッパ。
// 既存 LC-AUTH-09 apiClient.request を経由し、Authorization: Bearer 透過 (BFF)。

import { apiClient } from "../apiClient";

/**
 * ApiError は Order 系 API の HTTP エラー応答を表す class。
 *
 * mapOrderError (LC-26) で `instanceof` 判定して status code 別に分岐する。
 */
export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * PlaceOrderRequest は POST /api/orders のペイロード (凍結契約 §4.1)。
 */
export type PlaceOrderRequest = {
  category: "food";
  idempotencyKey: string;
  suggestionId?: string;
};

export type PlaceOrderResponse = {
  orderId: string;
  storeName: string;
  menuName: string;
  amount: number;
  remainingBalance: number;
  idempotent: boolean;
};

export type OrderRecord = {
  orderId: string;
  category: string;
  storeName: string;
  menuName: string;
  amount: number;
  orderedAt: string;
};

type ErrorResponseBody = {
  error?: { code?: string; message?: string };
};

async function parseError(res: Response): Promise<ApiError> {
  let body: ErrorResponseBody = {};
  try {
    body = (await res.json()) as ErrorResponseBody;
  } catch {
    /* JSON でない応答は無視、status code のみで判定 */
  }
  const code = body.error?.code ?? `HTTP_${res.status}`;
  const message = body.error?.message ?? res.statusText;
  return new ApiError(res.status, code, message);
}

/**
 * placeOrder は POST /api/orders を呼び出す。
 *
 * 成功時は PlaceOrderResponse を返し、HTTP エラーは ApiError を throw する。
 * Network エラーは fetch / apiClient 内部で throw され、本関数からそのまま伝播。
 */
export async function placeOrder(
  req: PlaceOrderRequest,
): Promise<PlaceOrderResponse> {
  const res = await apiClient.request({
    path: "/api/orders",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    throw await parseError(res);
  }
  return (await res.json()) as PlaceOrderResponse;
}

/**
 * fetchOrderHistory は GET /api/orders を呼び出す。
 *
 * limit はデフォルト 20、最大 100 (FD Q-11=A)。
 */
export async function fetchOrderHistory(
  limit = 20,
): Promise<OrderRecord[]> {
  const res = await apiClient.request({
    path: `/api/orders?limit=${limit}`,
    method: "GET",
  });
  if (!res.ok) {
    throw await parseError(res);
  }
  const body = (await res.json()) as { items: OrderRecord[] };
  return body.items;
}
