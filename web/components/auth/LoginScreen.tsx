"use client";

import { useState, FormEvent } from "react";
import { useSearchParams, useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { authMessages, AuthErrorWithCode } from "@/lib/authMessages";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function LoginScreen() {
  const { login } = useAuth();
  const router = useRouter();
  const params = useSearchParams();
  const fromSessionExpired = params.get("from") === "session_expired";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && password.length > 0 && !isSubmitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await login(email, password);
      router.push("/");
    } catch (err) {
      const code = err instanceof AuthErrorWithCode ? err.code : "UNKNOWN";
      setErrorMessage(authMessages[code]);
    } finally {
      setPassword(""); // R-Pwd-3-c: 送信完了次第クリア
      setIsSubmitting(false);
    }
  }

  return (
    <main data-testid="login-screen" style={{ padding: 24 }}>
      <h1>ログイン</h1>
      {fromSessionExpired && (
        <p data-testid="login-session-expired-hint" role="status">
          {authMessages.SESSION_EXPIRED}
        </p>
      )}
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="login-email">メールアドレス</label>
          <input
            id="login-email"
            data-testid="login-email-input"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div>
          <label htmlFor="login-password">パスワード</label>
          <input
            id="login-password"
            data-testid="login-password-input"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        {errorMessage && (
          <p data-testid="login-error-message" role="alert" aria-live="polite">
            {errorMessage}
          </p>
        )}
        <button
          type="submit"
          data-testid="login-submit-button"
          disabled={!canSubmit}
        >
          {isSubmitting ? "ログイン中…" : "ログイン"}
        </button>
      </form>
    </main>
  );
}
