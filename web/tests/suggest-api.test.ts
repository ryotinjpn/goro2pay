import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

// apiClient.request を mock 化して fetch / Cognito を bypass する
vi.mock("@/lib/apiClient", () => ({
  apiClient: {
    request: vi.fn(),
  },
}));

import { apiClient } from "@/lib/apiClient";
import { fetchSuggestion } from "@/lib/api/suggest";

const mockedRequest = apiClient.request as unknown as ReturnType<typeof vi.fn>;

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("fetchSuggestion (LC-SUGGEST frontend)", () => {
  beforeEach(() => {
    mockedRequest.mockReset();
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("200 履歴十分 → suggestion を返す", async () => {
    mockedRequest.mockResolvedValue(
      jsonResponse(200, {
        hasSuggestion: true,
        suggestionId: "01HXSUGGEST",
        title: "そろそろご飯めんどくさいですよね？",
        plan: { storeName: "CoCo壱", menuName: "ポークカレー", amount: 1200, category: "food" },
      }),
    );
    const res = await fetchSuggestion();
    expect(res.hasSuggestion).toBe(true);
    expect(res.suggestionId).toBe("01HXSUGGEST");
    expect(res.plan?.storeName).toBe("CoCo壱");
    expect(res.plan?.amount).toBe(1200);
  });

  test("200 履歴不足 → hasSuggestion:false", async () => {
    mockedRequest.mockResolvedValue(jsonResponse(200, { hasSuggestion: false }));
    const res = await fetchSuggestion();
    expect(res.hasSuggestion).toBe(false);
  });

  test("HTTP エラー → hasSuggestion:false に丸める (NFRD-D17、メイン機能を阻害しない)", async () => {
    mockedRequest.mockResolvedValue(jsonResponse(500, {}));
    const res = await fetchSuggestion();
    expect(res.hasSuggestion).toBe(false);
  });

  test("GET /api/suggest を叩く", async () => {
    mockedRequest.mockResolvedValue(jsonResponse(200, { hasSuggestion: false }));
    await fetchSuggestion();
    expect(mockedRequest).toHaveBeenCalledWith(
      expect.objectContaining({ path: "/api/suggest", method: "GET" }),
    );
  });
});
