// F-I1 / F-I2 修正検証: OrderHistoryList の error 優先 + aria-busy
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

import { OrderHistoryList } from "@/components/order/OrderHistoryList";
import * as orderHistoryHook from "@/hooks/useOrderHistory";

// 完全な useQuery 戻り値を作る簡易ヘルパ (テストで使うフィールドだけ反映)。
type PartialQuery = Partial<
  ReturnType<typeof orderHistoryHook.useOrderHistory>
>;
function makeQuery(over: PartialQuery): ReturnType<typeof orderHistoryHook.useOrderHistory> {
  return {
    data: undefined,
    isLoading: false,
    isFetching: false,
    isError: false,
    isSuccess: false,
    isPending: false,
    error: null,
    ...over,
  } as ReturnType<typeof orderHistoryHook.useOrderHistory>;
}

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>{ui}</QueryClientProvider>,
  );
}

describe("OrderHistoryList (LC-24 / F-I1 / F-I2)", () => {
  beforeEach(() => {
    vi.spyOn(orderHistoryHook, "useOrderHistory");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("初回ロード中はスケルトンを表示", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({ isLoading: true, isFetching: true }),
    );
    renderWithClient(<OrderHistoryList />);
    expect(screen.getByTestId("order-history-skeleton")).toBeInTheDocument();
  });

  test("F-I1: 初回ロード失敗時 (data なし + isError) はエラーバナー、空状態文言は出さない", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({ isError: true }),
    );
    renderWithClient(<OrderHistoryList />);
    expect(screen.getByTestId("order-history-error")).toBeInTheDocument();
    expect(screen.queryByTestId("order-history-empty")).toBeNull();
  });

  test("0 件 + 正常 (data === [])  → 「履歴がまだありません」", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({ data: [], isSuccess: true }),
    );
    renderWithClient(<OrderHistoryList />);
    expect(screen.getByTestId("order-history-empty")).toBeInTheDocument();
  });

  test("F-I2: refetch 中 (isFetching=true) は aria-busy=true", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({
        data: [
          {
            orderId: "01HZ",
            category: "food",
            storeName: "ゴロゴロ食堂",
            menuName: "おまかせ定食",
            amount: 1000,
            orderedAt: "2026-05-24T00:00:00Z",
          },
        ],
        isFetching: true,
      }),
    );
    renderWithClient(<OrderHistoryList />);
    const list = screen.getByTestId("order-history-list");
    expect(list.getAttribute("aria-busy")).toBe("true");
  });

  test("通常表示 (isFetching=false) は aria-busy=false", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({
        data: [
          {
            orderId: "01HZ",
            category: "food",
            storeName: "ぐうたら亭",
            menuName: "手抜き丼",
            amount: 800,
            orderedAt: "2026-05-23T00:00:00Z",
          },
        ],
        isSuccess: true,
      }),
    );
    renderWithClient(<OrderHistoryList />);
    const list = screen.getByTestId("order-history-list");
    expect(list.getAttribute("aria-busy")).toBe("false");
  });

  test("data 既存 + isError (refetch 失敗) は前回データ + エラーバナー重畳", () => {
    vi.mocked(orderHistoryHook.useOrderHistory).mockReturnValue(
      makeQuery({
        data: [
          {
            orderId: "01HZ",
            category: "food",
            storeName: "ダメ屋",
            menuName: "やる気なしカレー",
            amount: 1200,
            orderedAt: "2026-05-22T00:00:00Z",
          },
        ],
        isError: true,
      }),
    );
    renderWithClient(<OrderHistoryList />);
    expect(screen.getByTestId("order-history-list")).toBeInTheDocument();
    expect(screen.getByText("履歴の取得に失敗しました")).toBeInTheDocument();
  });
});
