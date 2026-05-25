// Unit B InsufficientBalanceModal (P-DEG-02) の atom 連携 + 操作検証。
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { Provider, createStore } from "jotai";

import { InsufficientBalanceModal } from "@/components/budget/InsufficientBalanceModal";
import { insufficientBalanceAtom } from "@/state/budget";

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

function renderWithStore(initial: boolean) {
  const store = createStore();
  store.set(insufficientBalanceAtom, initial);
  return {
    store,
    ...render(
      <Provider store={store}>
        <InsufficientBalanceModal />
      </Provider>,
    ),
  };
}

describe("InsufficientBalanceModal (LC-BUDGET-10 / P-DEG-02)", () => {
  beforeEach(() => {
    pushMock.mockReset();
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("atom=false のときは描画されない", () => {
    renderWithStore(false);
    expect(screen.queryByTestId("insufficient-balance-modal")).toBeNull();
  });

  it("atom=true のとき modal が open", () => {
    renderWithStore(true);
    expect(screen.getByTestId("insufficient-balance-modal")).toBeInTheDocument();
  });

  it("「もっとダメになる」ボタンで /budget 遷移 + atom が false", () => {
    const { store } = renderWithStore(true);
    fireEvent.click(screen.getByTestId("raise-budget-button"));
    expect(pushMock).toHaveBeenCalledWith("/budget");
    expect(store.get(insufficientBalanceAtom)).toBe(false);
  });

  it("閉じるボタンで atom が false (遷移なし)", () => {
    const { store } = renderWithStore(true);
    fireEvent.click(screen.getByTestId("close-modal-button"));
    expect(pushMock).not.toHaveBeenCalled();
    expect(store.get(insufficientBalanceAtom)).toBe(false);
  });
});
