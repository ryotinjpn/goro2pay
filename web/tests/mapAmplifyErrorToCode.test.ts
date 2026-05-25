import { describe, it, expect } from "vitest";
import { mapAmplifyErrorToCode } from "@/hooks/useAuth";

describe("mapAmplifyErrorToCode", () => {
  it("NotAuthorizedException → INVALID_CREDENTIALS", () => {
    expect(mapAmplifyErrorToCode({ name: "NotAuthorizedException" })).toBe("INVALID_CREDENTIALS");
  });

  it("UserNotFoundException → INVALID_CREDENTIALS (R-Err-1)", () => {
    expect(mapAmplifyErrorToCode({ name: "UserNotFoundException" })).toBe("INVALID_CREDENTIALS");
  });

  it("UsernameExistsException → EMAIL_ALREADY_EXISTS", () => {
    expect(mapAmplifyErrorToCode({ name: "UsernameExistsException" })).toBe("EMAIL_ALREADY_EXISTS");
  });

  it("InvalidPasswordException → WEAK_PASSWORD", () => {
    expect(mapAmplifyErrorToCode({ name: "InvalidPasswordException" })).toBe("WEAK_PASSWORD");
  });

  it("TooManyRequestsException → RATE_LIMIT_EXCEEDED", () => {
    expect(mapAmplifyErrorToCode({ name: "TooManyRequestsException" })).toBe("RATE_LIMIT_EXCEEDED");
  });

  it("NetworkError → NETWORK_ERROR", () => {
    expect(mapAmplifyErrorToCode({ name: "NetworkError" })).toBe("NETWORK_ERROR");
  });

  it("UserAlreadyAuthenticatedException → UNKNOWN (競合状態の保険分類)", () => {
    expect(mapAmplifyErrorToCode({ name: "UserAlreadyAuthenticatedException" })).toBe("UNKNOWN");
  });

  it("未知の name → UNKNOWN", () => {
    expect(mapAmplifyErrorToCode({ name: "SomeNewException" })).toBe("UNKNOWN");
  });

  it("null / undefined / 空オブジェクト → UNKNOWN", () => {
    expect(mapAmplifyErrorToCode(null)).toBe("UNKNOWN");
    expect(mapAmplifyErrorToCode(undefined)).toBe("UNKNOWN");
    expect(mapAmplifyErrorToCode({})).toBe("UNKNOWN");
  });

  describe("InvalidParameterException — メッセージで分岐", () => {
    it("message に email を含む → INVALID_EMAIL_FORMAT", () => {
      expect(
        mapAmplifyErrorToCode({
          name: "InvalidParameterException",
          message: "Invalid email address format.",
        }),
      ).toBe("INVALID_EMAIL_FORMAT");
    });

    it("message に email を含まない (auth flow 不許可) → UNKNOWN", () => {
      // App Client の auth flow 設定ミスなど、メアド形式とは無関係に
      // Cognito が InvalidParameterException を返すケースで「メアド形式
      // が正しくない」と誤誘導しないことを保証する。
      expect(
        mapAmplifyErrorToCode({
          name: "InvalidParameterException",
          message: "Auth flow not enabled for this client",
        }),
      ).toBe("UNKNOWN");
    });

    it("message が無い → UNKNOWN", () => {
      expect(mapAmplifyErrorToCode({ name: "InvalidParameterException" })).toBe("UNKNOWN");
    });
  });
});
