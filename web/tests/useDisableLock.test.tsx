import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";

import { useDisableLock } from "@/hooks/useDisableLock";

describe("useDisableLock (LC-25 / P-FE-LOCK-01)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  test("triggerLock 直後は isLocked=true", () => {
    const { result } = renderHook(() => useDisableLock(1000));
    expect(result.current.isLocked).toBe(false);

    act(() => result.current.triggerLock());
    expect(result.current.isLocked).toBe(true);
  });

  test("999ms 経過時はまだ lock 中", () => {
    const { result } = renderHook(() => useDisableLock(1000));
    act(() => result.current.triggerLock());

    act(() => {
      vi.advanceTimersByTime(999);
    });
    expect(result.current.isLocked).toBe(true);
  });

  test("1000ms 経過で lock 解除 (境界)", () => {
    const { result } = renderHook(() => useDisableLock(1000));
    act(() => result.current.triggerLock());

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.isLocked).toBe(false);
  });

  test("triggerLock を再度呼ぶと延長される", () => {
    const { result } = renderHook(() => useDisableLock(1000));
    act(() => result.current.triggerLock());
    act(() => {
      vi.advanceTimersByTime(800);
    });
    act(() => result.current.triggerLock()); // 再 trigger で +1000ms

    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(result.current.isLocked).toBe(true);

    act(() => {
      vi.advanceTimersByTime(600);
    });
    expect(result.current.isLocked).toBe(false);
  });
});
