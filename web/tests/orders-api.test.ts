import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

// apiClient.request を mock 化して fetch / Cognito を bypass する
vi.mock("@/lib/apiClient", () => ({
  apiClient: {
    request: vi.fn(),
  },
}));

// mock 後に import (順序重要)
import { apiClient } from "@/lib/apiClient";
import {
  ApiError,
  fetchOrderHistory,
  placeOrder,
} from "@/lib/api/orders";

const mockedRequest = apiClient.request as unknown as ReturnType<typeof vi.fn>;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("placeOrder (LC-33)", () => {
  beforeEach(() => {
    mockedRequest.mockReset();
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("200 → PlaceOrderResponse を返す", async () => {
    mockedRequest.mockResolvedValue(
      jsonResponse(200, {
        orderId: "01HZ",
        storeName: "ゴロゴロ食堂",
        menuName: "おまかせ定食",
        amount: 1000,
        remainingBalance: 9000,
        idempotent: false,
      }),
    );
    const res = await placeOrder({ category: "food", idempotencyKey: "01HZIDEM" });
    expect(res.orderId).toBe("01HZ");
    expect(res.amount).toBe(1000);
  });

  test("402 → ApiError を throw", async () => {
    mockedRequest.mockResolvedValue(
      jsonResponse(402, {
        error: { code: "INSUFFICIENT_FUNDS", message: "no funds" },
      }),
    );
    await expect(
      placeOrder({ category: "food", idempotencyKey: "01HZ" }),
    ).rejects.toBeInstanceOf(ApiError);
  });

  test("500 → ApiError(status=500)", async () => {
    mockedRequest.mockResolvedValue(jsonResponse(500, {}));
    try {
      await placeOrder({ category: "food", idempotencyKey: "01HZ" });
      throw new Error("must throw");
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      expect((e as ApiError).status).toBe(500);
    }
  });
});

describe("fetchOrderHistory (LC-33)", () => {
  beforeEach(() => {
    mockedRequest.mockReset();
  });

  test("200 → items 配列", async () => {
    mockedRequest.mockResolvedValue(
      jsonResponse(200, {
        items: [
          { orderId: "1", category: "food", storeName: "X", menuName: "Y", amount: 1000, orderedAt: "2026-05-24T00:00:00Z" },
        ],
      }),
    );
    const items = await fetchOrderHistory(10);
    expect(items).toHaveLength(1);
    expect(items[0].orderId).toBe("1");
  });

  test("limit を URL に渡す", async () => {
    mockedRequest.mockResolvedValue(jsonResponse(200, { items: [] }));
    await fetchOrderHistory(50);
    expect(mockedRequest).toHaveBeenCalledWith(
      expect.objectContaining({ path: "/api/orders?limit=50" }),
    );
  });
});
