"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/hooks/useAuth";
import { authMessages, AuthErrorWithCode } from "@/lib/authMessages";
import { COPY } from "@/lib/copy";

import { ScreenFrame } from "@/components/order/ScreenFrame";
import { BrandHeader } from "@/components/order/BrandHeader";

import styles from "./SignupScreen.module.css";

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
    <ScreenFrame testid="signup-screen">
      <BrandHeader />
      <h1 className={styles.h1}>
        <div>{COPY.signup.h1Line1}</div>
        <div>{COPY.signup.h1Line2}</div>
      </h1>
      <div className={styles.sub}>{COPY.signup.sub}</div>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
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
        <div className={styles.field}>
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
          <ul className={styles.hints} data-testid="signup-password-hints" aria-live="polite">
            <li className={hints.lengthOk ? styles.ok : ""}>
              {hints.lengthOk ? "✓" : "・"} 8 文字以上
            </li>
            <li className={hints.upperOk ? styles.ok : ""}>
              {hints.upperOk ? "✓" : "・"} 英大文字を含む
            </li>
            <li className={hints.lowerOk ? styles.ok : ""}>
              {hints.lowerOk ? "✓" : "・"} 英小文字を含む
            </li>
            <li className={hints.digitOk ? styles.ok : ""}>
              {hints.digitOk ? "✓" : "・"} 数字を含む
            </li>
          </ul>
        </div>
        {errorMessage && (
          <p
            className={styles.error}
            data-testid="signup-error-message"
            role="alert"
            aria-live="polite"
          >
            {errorMessage}
          </p>
        )}
        <button
          type="submit"
          className={styles.submit}
          data-testid="signup-submit-button"
          disabled={!canSubmit}
        >
          {isSubmitting ? COPY.signup.submitting : COPY.signup.submit}
        </button>
      </form>
    </ScreenFrame>
  );
}
