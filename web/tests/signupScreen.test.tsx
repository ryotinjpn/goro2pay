import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";

const signupMock = vi.fn().mockResolvedValue(undefined);
vi.mock("@/hooks/useAuth", () => ({
  useAuth: () => ({
    signup: (email: string, pw: string) => signupMock(email, pw),
    status: "unauthenticated",
    user: null,
    isAuthenticated: false,
    login: vi.fn(),
    logout: vi.fn(),
  }),
}));

const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, replace: pushMock }),
}));

import { SignupScreen } from "@/components/auth/SignupScreen";

describe("SignupScreen", () => {
  beforeEach(() => {
    signupMock.mockClear();
    pushMock.mockClear();
    signupMock.mockResolvedValue(undefined);
  });

  it("メール無効 / パスワード弱なら submit ボタンが disabled", () => {
    render(<SignupScreen />);
    const submit = screen.getByRole("button", { name: /登録|送信|はじめる|サインアップ/i });
    // 初期 (空入力) は disabled
    expect(submit).toBeDisabled();

    // メールだけ正しくしてもパスワード強度を満たさないと disabled
    const email = screen.getByTestId("signup-email-input");
    fireEvent.change(email, { target: { value: "user@example.com" } });
    expect(submit).toBeDisabled();
  });

  it("メール有効 + 強いパスワードなら submit ボタンが enabled、signup を呼ぶ", async () => {
    render(<SignupScreen />);
    const email = screen.getByTestId("signup-email-input");
    const password = screen.getByTestId("signup-password-input");

    fireEvent.change(email, { target: { value: "user@example.com" } });
    fireEvent.change(password, { target: { value: "Abcdef12" } }); // 8 chars + upper + lower + digit

    const submit = screen.getByRole("button", { name: /登録|送信|はじめる|サインアップ/i });
    expect(submit).not.toBeDisabled();

    await act(async () => {
      fireEvent.click(submit);
    });

    expect(signupMock).toHaveBeenCalledWith("user@example.com", "Abcdef12");
    expect(pushMock).toHaveBeenCalledWith("/");
  });

  it("送信完了後にパスワード入力がクリアされる (R-Pwd-3-c)", async () => {
    render(<SignupScreen />);
    const email = screen.getByTestId("signup-email-input");
    const password = screen.getByTestId("signup-password-input") as HTMLInputElement;

    fireEvent.change(email, { target: { value: "user@example.com" } });
    fireEvent.change(password, { target: { value: "Abcdef12" } });

    const submit = screen.getByRole("button", { name: /登録|送信|はじめる|サインアップ/i });
    await act(async () => {
      fireEvent.click(submit);
    });

    expect(password.value).toBe("");
  });
});
