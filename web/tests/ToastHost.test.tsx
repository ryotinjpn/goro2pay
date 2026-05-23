import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, test } from "vitest";
import { getDefaultStore } from "jotai";

import { ToastHost } from "@/components/order/ToastHost";
import { toastsAtom } from "@/state/toastAtoms";

describe("ToastHost (LC-30 / P-FE-TOAST-02)", () => {
  beforeEach(() => {
    getDefaultStore().set(toastsAtom, []);
  });
  afterEach(() => {
    getDefaultStore().set(toastsAtom, []);
  });

  test("0 件のとき DOM 上に Toast は存在しない", () => {
    render(<ToastHost />);
    expect(screen.queryAllByRole("status")).toHaveLength(0);
  });

  test("3 件追加で 3 件全て表示される", () => {
    getDefaultStore().set(toastsAtom, [
      { id: "1", text: "a", durationMs: 5000 },
      { id: "2", text: "b", durationMs: 5000 },
      { id: "3", text: "c", durationMs: 5000 },
    ]);
    render(<ToastHost />);
    expect(screen.getAllByRole("status")).toHaveLength(3);
  });

  test("4 件追加でも表示は 3 件のみ (最大 3 件キュー)", () => {
    getDefaultStore().set(toastsAtom, [
      { id: "1", text: "a", durationMs: 5000 },
      { id: "2", text: "b", durationMs: 5000 },
      { id: "3", text: "c", durationMs: 5000 },
      { id: "4", text: "d", durationMs: 5000 },
    ]);
    render(<ToastHost />);
    expect(screen.getAllByRole("status")).toHaveLength(3);
    // 先頭 3 件 (a, b, c) のみ表示、d は表示されない
    expect(screen.queryByText("d")).toBeNull();
  });

  test("コンテナに role=region と aria-live=polite が付く", () => {
    render(<ToastHost />);
    const region = screen.getByRole("region");
    expect(region.getAttribute("aria-live")).toBe("polite");
  });
});
