// AddBudgetForm の入力 / バリデーション / 送信ボタン挙動の検証。
//
// 仕様: currentBudget + addition を mutate に渡す (backend は「予算総額」を受ける)。
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { AddBudgetForm } from "@/components/budget/AddBudgetForm";
import * as setBudgetHook from "@/hooks/useSetBudget";

function renderWithClient(ui: React.ReactElement) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>{ui}</QueryClientProvider>,
  );
}

type PartialMutation = Partial<ReturnType<typeof setBudgetHook.useSetBudget>>;
function makeMutation(over: PartialMutation): ReturnType<typeof setBudgetHook.useSetBudget> {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
    error: null,
    ...over,
  } as ReturnType<typeof setBudgetHook.useSetBudget>;
}

describe("AddBudgetForm", () => {
  beforeEach(() => {
    vi.spyOn(setBudgetHook, "useSetBudget");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("初期は最初のクイック額 (¥10,000) が選択され合計プレビューが現在 + 10,000 を出す", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<AddBudgetForm currentBudget={30_000} />);
    expect(screen.getByTestId("add-budget-preview")).toHaveTextContent("¥40,000");
    expect(screen.getByTestId("add-budget-preview")).toHaveTextContent("現在 ¥30,000");
  });

  it("クイックボタン (+¥20,000) を選択すると合計が currentBudget + 20,000 になる", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<AddBudgetForm currentBudget={30_000} />);
    fireEvent.click(screen.getByTestId("add-budget-quick-20000"));
    expect(screen.getByTestId("add-budget-preview")).toHaveTextContent("¥50,000");
  });

  it("送信時に mutate(currentBudget + addition) が呼ばれる (= 予算総額を送る)", () => {
    const mutate = vi.fn();
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({ mutate }));
    renderWithClient(<AddBudgetForm currentBudget={30_000} />);
    fireEvent.click(screen.getByTestId("add-budget-quick-20000"));
    fireEvent.click(screen.getByTestId("add-budget-submit"));
    expect(mutate).toHaveBeenCalledWith(50_000);
  });

  it("合計が 100,000 を超えるクイック額は disabled", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<AddBudgetForm currentBudget={90_000} />);
    const btn50k = screen.getByTestId("add-budget-quick-50000") as HTMLButtonElement;
    expect(btn50k.disabled).toBe(true);
  });

  it("カスタム入力で合計上限超過時にバリデーションエラーが出る", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<AddBudgetForm currentBudget={80_000} />);
    const input = screen.getByTestId("add-budget-input") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "30000" } });
    expect(screen.getByTestId("add-budget-validation-error")).toHaveTextContent(
      "合計が 100,000 円を超える",
    );
  });

  it("isPending 中は submit disabled + 「追加中...」表示", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(
      makeMutation({ isPending: true }),
    );
    renderWithClient(<AddBudgetForm currentBudget={30_000} />);
    const submit = screen.getByTestId("add-budget-submit") as HTMLButtonElement;
    expect(submit.disabled).toBe(true);
    expect(submit.textContent).toContain("追加中");
  });
});
