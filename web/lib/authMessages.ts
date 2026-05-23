// AuthErrorCode → 日本語ダメ化トーン軽メッセージ
// Functional Design domain-entities.md §6.2 / frontend-components.md §11
// Q-A8=B (ダメ化トーン軽) を採用

export type AuthErrorCode =
  | "INVALID_CREDENTIALS"
  | "EMAIL_ALREADY_EXISTS"
  | "WEAK_PASSWORD"
  | "INVALID_EMAIL_FORMAT"
  | "RATE_LIMIT_EXCEEDED"
  | "SESSION_EXPIRED"
  | "NETWORK_ERROR"
  | "UNKNOWN";

export const authMessages: Record<AuthErrorCode, string> = {
  INVALID_CREDENTIALS: "ログインすらめんどくさいですよね…もう一度お試しください",
  EMAIL_ALREADY_EXISTS: "このメールはもう使われています。ログインしますか？",
  WEAK_PASSWORD: "パスワードは8文字以上、英大文字・小文字・数字を含めてください",
  INVALID_EMAIL_FORMAT: "メールアドレスの形式が正しくないようです",
  RATE_LIMIT_EXCEEDED: "ちょっと頑張りすぎです。少し待ってから試してください",
  SESSION_EXPIRED: "お疲れ様でした。もう一度ログインしてください",
  NETWORK_ERROR: "通信が…ちょっと待ってもう一度",
  UNKNOWN: "うまくいきませんでした。もう一度お試しください",
};

export class AuthErrorWithCode extends Error {
  code: AuthErrorCode;

  constructor(code: AuthErrorCode, originalError?: unknown) {
    super(`AuthError: ${code}`);
    this.code = code;
    // パスワードを example 含む originalError は出力しない (R-Pwd-3-b)
    void originalError;
  }
}
