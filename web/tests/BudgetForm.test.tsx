// Unit B BudgetForm の入力 / バリデーション / 送信ボタン挙動の検証。
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { BudgetForm } from "@/components/budget/BudgetForm";
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

describe("BudgetForm (LC-BUDGET-12)", () => {
  beforeEach(() => {
    vi.spyOn(setBudgetHook, "useSetBudget");
  });
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("クイックボタンクリックで入力欄が更新される", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<BudgetForm initialBudget={null} />);
    fireEvent.click(screen.getByTestId("budget-quick-50000"));
    const input = screen.getByTestId("budget-input") as HTMLInputElement;
    expect(input.value).toBe("50000");
  });

  it("VR-B-02 違反 (1000 で割り切れない) でバリデーションエラー", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<BudgetForm initialBudget={null} />);
    const input = screen.getByTestId("budget-input") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "1500" } });
    expect(screen.getByTestId("budget-validation-error")).toHaveTextContent("1,000 円単位");
  });

  it("VR-B-01 違反 (上限超過) でバリデーションエラー", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<BudgetForm initialBudget={null} />);
    const input = screen.getByTestId("budget-input") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "200000" } });
    expect(screen.getByTestId("budget-validation-error")).toHaveTextContent("範囲");
  });

  it("送信ボタンは validationError 時 disabled", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({}));
    renderWithClient(<BudgetForm initialBudget={null} />);
    const input = screen.getByTestId("budget-input") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "1500" } });
    const submit = screen.getByTestId("budget-submit") as HTMLButtonElement;
    expect(submit.disabled).toBe(true);
  });

  it("送信時に mutate(monthlyBudget) が呼ばれる", () => {
    const mutate = vi.fn();
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({ mutate }));
    renderWithClient(<BudgetForm initialBudget={null} />);
    fireEvent.click(screen.getByTestId("budget-quick-30000"));
    fireEvent.click(screen.getByTestId("budget-submit"));
    expect(mutate).toHaveBeenCalledWith(30000);
  });

  it("isPending 中は送信ボタン disabled + 「設定中...」", () => {
    vi.mocked(setBudgetHook.useSetBudget).mockReturnValue(makeMutation({ isPending: true }));
    renderWithClient(<BudgetForm initialBudget={30000} />);
    const submit = screen.getByTestId("budget-submit") as HTMLButtonElement;
    expect(submit.disabled).toBe(true);
    expect(submit.textContent).toContain("設定中");
  });
});
