import { describe, it, expect } from "vitest";
import { authMessages, AuthErrorWithCode, type AuthErrorCode } from "@/lib/authMessages";

describe("authMessages", () => {
  it("9 種別すべての ErrorCode に日本語メッセージが定義されている", () => {
    const codes: AuthErrorCode[] = [
      "INVALID_CREDENTIALS",
      "EMAIL_ALREADY_EXISTS",
      "WEAK_PASSWORD",
      "INVALID_EMAIL_FORMAT",
      "RATE_LIMIT_EXCEEDED",
      "SESSION_EXPIRED",
      "NETWORK_ERROR",
      "UNKNOWN",
    ];
    for (const code of codes) {
      const msg = authMessages[code];
      expect(msg).toBeTruthy();
      expect(typeof msg).toBe("string");
      // 日本語文字 (ひらがな/カタカナ/漢字) を含むか
      expect(msg).toMatch(/[぀-ヿ一-鿿]/);
    }
  });
});

describe("AuthErrorWithCode", () => {
  it("code を持って throw できる", () => {
    const err = new AuthErrorWithCode("INVALID_CREDENTIALS");
    expect(err).toBeInstanceOf(Error);
    expect(err.code).toBe("INVALID_CREDENTIALS");
  });

  it("originalError を内部参照しない (パスワード漏洩防止)", () => {
    const sensitiveError = { name: "NotAuthorizedException", body: "password=secret123" };
    const err = new AuthErrorWithCode("INVALID_CREDENTIALS", sensitiveError);
    // err.message に sensitive 情報が含まれてはならない
    expect(err.message).not.toContain("secret123");
    expect(err.message).not.toContain("password=");
  });
});
