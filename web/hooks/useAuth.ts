// useAuth フック (NFR Design LC-AUTH-08)
// 公開メソッド名は unit-interfaces.md §9 に従い signup / login / logout
// (PR #66 で凍結契約と整合)
//
// 内部実装は Amplify Auth v6 の signUp / signIn / signOut を呼ぶ。
// パスワード state は呼び出し側が管理 (R-Pwd-3-c の都合、本フックでは保持しない)。

"use client";

import { useEffect, useState, useCallback } from "react";
import {
  signUp as amplifySignUp,
  signIn as amplifySignIn,
  signOut as amplifySignOut,
  fetchAuthSession,
  getCurrentUser,
} from "aws-amplify/auth";
import { Hub } from "aws-amplify/utils";

import { AuthErrorWithCode, type AuthErrorCode } from "@/lib/authMessages";
import { apiClient } from "@/lib/apiClient";

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

export interface UseAuthUser {
  userId: string;
  email: string;
}

export interface UseAuthReturn {
  status: AuthStatus;
  user: UseAuthUser | null;
  isAuthenticated: boolean;
  signup: (email: string, password: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

function normalizeEmail(email: string): string {
  return email.trim().toLowerCase();
}

export function mapAmplifyErrorToCode(err: unknown): AuthErrorCode {
  const e = err as { name?: string; message?: string } | null;
  const name = e?.name ?? "";
  const message = e?.message ?? "";
  switch (name) {
    case "NotAuthorizedException":
    case "UserNotFoundException":
      return "INVALID_CREDENTIALS";
    case "UsernameExistsException":
      return "EMAIL_ALREADY_EXISTS";
    case "InvalidPasswordException":
      return "WEAK_PASSWORD";
    case "InvalidParameterException":
      // Cognito は様々な理由で InvalidParameterException を投げる
      // (App Client の auth flow 未許可、属性検証エラー、etc)。
      // クライアント側で R-Email-1 の事前検証を通過した状態で
      // ここに到達するのは設定不整合の蓋然性が高いため UNKNOWN に倒す。
      // 「email」を含むメッセージの時のみメール形式エラーに割り当てる。
      return /email/i.test(message) ? "INVALID_EMAIL_FORMAT" : "UNKNOWN";
    case "TooManyRequestsException":
      return "RATE_LIMIT_EXCEEDED";
    case "NetworkError":
      return "NETWORK_ERROR";
    case "UserAlreadyAuthenticatedException":
      // 既ログイン状態のまま signUp / signIn を再度呼んだケース。
      // 呼出側の signup / login が事前に signOut を試みているので
      // 通常は到達しないが、競合状態の保険として明示的に分類する。
      return "UNKNOWN";
    default:
      return "UNKNOWN";
  }
}

const USE_MOCK = process.env.NEXT_PUBLIC_USE_MOCK === "true";
const MOCK_USER: UseAuthUser = { userId: "mock-user-id", email: "demo@example.com" };

export function useAuth(): UseAuthReturn {
  const [status, setStatus] = useState<AuthStatus>("loading");
  const [user, setUser] = useState<UseAuthUser | null>(null);

  const refreshSession = useCallback(async () => {
    if (USE_MOCK) {
      setUser(MOCK_USER);
      setStatus("authenticated");
      return;
    }
    try {
      const current = await getCurrentUser();
      const session = await fetchAuthSession();
      const accessToken = session.tokens?.accessToken;
      const sub = (accessToken?.payload?.sub as string | undefined) ?? "";
      // email は IdToken にあるが、表示用に取り出すなら ID token claims を見る
      const idToken = session.tokens?.idToken;
      const email = (idToken?.payload?.email as string | undefined) ?? current.username;
      if (sub) {
        setUser({ userId: sub, email });
        setStatus("authenticated");
        return;
      }
    } catch {
      // ignored
    }
    setUser(null);
    setStatus("unauthenticated");
  }, []);

  useEffect(() => {
    refreshSession();
    // Amplify Hub の signedIn / signedOut で同期
    const unsubscribe = Hub.listen("auth", (capsule) => {
      const event = capsule?.payload?.event;
      if (event === "signedIn" || event === "signedOut") {
        refreshSession();
      }
    });
    return () => {
      // Amplify v6 の Hub.listen は cleanup function を返す
      if (typeof unsubscribe === "function") unsubscribe();
    };
  }, [refreshSession]);

  const signup = useCallback(async (email: string, password: string) => {
    const normalized = normalizeEmail(email);
    // 既ログイン状態のまま signUp を呼ぶと UserAlreadyAuthenticatedException が
    // 飛ぶため、念のため事前に signOut を試みる (失敗しても無視)。
    await amplifySignOut().catch(() => undefined);
    try {
      const result = await amplifySignUp({
        username: normalized,
        password,
        options: {
          userAttributes: { email: normalized },
        },
      });
      // auto-confirm されているはずなのでそのまま signIn
      if (!result.isSignUpComplete) {
        // 想定外: Pre Sign-up Trigger が動かなかったか、設定ミス
        throw new AuthErrorWithCode("UNKNOWN");
      }
      await amplifySignIn({ username: normalized, password });
      await refreshSession();
    } catch (err) {
      if (err instanceof AuthErrorWithCode) throw err;
      throw new AuthErrorWithCode(mapAmplifyErrorToCode(err), err);
    }
  }, [refreshSession]);

  const login = useCallback(async (email: string, password: string) => {
    const normalized = normalizeEmail(email);
    // signup と同じ理由で signOut を先に試みる。
    await amplifySignOut().catch(() => undefined);
    try {
      await amplifySignIn({ username: normalized, password });
      await refreshSession();
    } catch (err) {
      if (err instanceof AuthErrorWithCode) throw err;
      throw new AuthErrorWithCode(mapAmplifyErrorToCode(err), err);
    }
  }, [refreshSession]);

  const logout = useCallback(async () => {
    // 監査ログを先に送る (Functional Design F-4 の順序: API → signOut)
    try {
      await apiClient.request({ path: "/api/auth/logout", method: "POST" });
    } catch {
      // 監査ログ送信に失敗しても signOut は実行する
    }
    try {
      await amplifySignOut({ global: true });
    } finally {
      await refreshSession();
    }
  }, [refreshSession]);

  return {
    status,
    user,
    isAuthenticated: status === "authenticated",
    signup,
    login,
    logout,
  };
}
