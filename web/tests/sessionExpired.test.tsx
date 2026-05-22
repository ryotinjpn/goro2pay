import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { Provider as JotaiProvider, getDefaultStore } from "jotai";

import { SessionExpiredModalHost } from "@/components/auth/SessionExpiredModalHost";
import { sessionExpiredAtom } from "@/state/auth";
import { _resetSessionExpiredHandling } from "@/lib/apiClient";

// next/navigation router のモック
const pushMock = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock, replace: pushMock }),
}));

describe("SessionExpiredModalHost", () => {
  beforeEach(() => {
    pushMock.mockClear();
    _resetSessionExpiredHandling();
    getDefaultStore().set(sessionExpiredAtom, null);
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("atom が null のときは modal を表示しない", () => {
    render(
      <JotaiProvider>
        <SessionExpiredModalHost />
      </JotaiProvider>
    );
    expect(screen.queryByTestId("session-expired-modal")).toBeNull();
  });

  it("atom が起動したら modal を表示し、1.5 秒後に /login へ遷移する", () => {
    render(
      <JotaiProvider>
        <SessionExpiredModalHost />
      </JotaiProvider>
    );

    act(() => {
      getDefaultStore().set(sessionExpiredAtom, { openedAt: Date.now() });
    });
    expect(screen.getByTestId("session-expired-modal")).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(1500);
    });
    expect(pushMock).toHaveBeenCalledWith("/login?from=session_expired");
  });
});
