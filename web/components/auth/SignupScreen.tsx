"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { authMessages, AuthErrorWithCode } from "@/lib/authMessages";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

type PwdHints = {
  lengthOk: boolean;
  upperOk: boolean;
  lowerOk: boolean;
  digitOk: boolean;
};

function checkPasswordHints(pw: string): PwdHints {
  return {
    lengthOk: pw.length >= 8,
    upperOk: /[A-Z]/.test(pw),
    lowerOk: /[a-z]/.test(pw),
    digitOk: /\d/.test(pw),
  };
}

function isPasswordValid(h: PwdHints): boolean {
  return h.lengthOk && h.upperOk && h.lowerOk && h.digitOk;
}

export function SignupScreen() {
  const { signup } = useAuth();
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const hints = checkPasswordHints(password);
  const isEmailValid = EMAIL_REGEX.test(email);
  const canSubmit = isEmailValid && isPasswordValid(hints) && !isSubmitting;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setIsSubmitting(true);
    setErrorMessage(null);
    try {
      await signup(email, password);
      router.push("/");
    } catch (err) {
      const code = err instanceof AuthErrorWithCode ? err.code : "UNKNOWN";
      setErrorMessage(authMessages[code]);
    } finally {
      setPassword(""); // R-Pwd-3-c
      setIsSubmitting(false);
    }
  }

  return (
    <main data-testid="signup-screen" style={{ padding: 24 }}>
      <h1>はじめる</h1>
      <p>30 秒でダメ化体験スタート</p>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="signup-email">メールアドレス</label>
          <input
            id="signup-email"
            data-testid="signup-email-input"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
        <div>
          <label htmlFor="signup-password">パスワード</label>
          <input
            id="signup-password"
            data-testid="signup-password-input"
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <ul data-testid="signup-password-hints" aria-live="polite">
            <li>{hints.lengthOk ? "✓" : "・"} 8 文字以上</li>
            <li>{hints.upperOk ? "✓" : "・"} 英大文字を含む</li>
            <li>{hints.lowerOk ? "✓" : "・"} 英小文字を含む</li>
            <li>{hints.digitOk ? "✓" : "・"} 数字を含む</li>
          </ul>
        </div>
        {errorMessage && (
          <p data-testid="signup-error-message" role="alert" aria-live="polite">
            {errorMessage}
          </p>
        )}
        <button
          type="submit"
          data-testid="signup-submit-button"
          disabled={!canSubmit}
        >
          {isSubmitting ? "登録中…" : "登録"}
        </button>
      </form>
    </main>
  );
}
