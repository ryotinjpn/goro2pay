import { describe, it, expect } from "vitest";
import {
  COPY,
  composeSuggestSubLabel,
  composeIncreaseBudgetLabel,
  composeMonthlyMeta,
} from "@/lib/copy";

describe("COPY constants", () => {
  it("contains landing hero", () => {
    expect(COPY.landing.heroLine1).toBe("考えるな。");
    expect(COPY.landing.heroLine2).toBe("押せ。");
    expect(COPY.landing.sub).toBe("面倒は、こちらで引き受ける。");
  });
  it("contains main button labels", () => {
    expect(COPY.main.idleMain).toBe("めんどくさい");
    expect(COPY.main.idleSub).toBe("— 押せ。考えるな。");
    expect(COPY.main.suggestMain).toBe("押す。");
  });
  it("contains dead state copy", () => {
    expect(COPY.main.deadVerdict).toBe("今月は、終わりだ。");
    expect(COPY.main.deadButtonSub).toBe("— 上出来だ。使い切ったな。");
  });
  it("contains complete screen copy", () => {
    expect(COPY.complete.verdict).toBe("いい判断だ。");
    expect(COPY.complete.body).toBe("面倒は片付いた。");
    expect(COPY.complete.next).toBe("次を待て。");
  });
});

describe("composeSuggestSubLabel", () => {
  it("formats store + amount with リヴァイ調 ending", () => {
    expect(composeSuggestSubLabel("CoCo壱", 1200)).toBe("— CoCo壱 ¥1,200 だ。");
  });
  it("formats large numbers with thousand separator", () => {
    expect(composeSuggestSubLabel("ロイヤルホスト", 12500)).toBe(
      "— ロイヤルホスト ¥12,500 だ。",
    );
  });
});

describe("composeIncreaseBudgetLabel", () => {
  it("formats yen with thousand separator", () => {
    expect(composeIncreaseBudgetLabel(50000)).toBe("¥50,000。来月もこの調子だ。");
  });
});

describe("composeMonthlyMeta", () => {
  it("counts good judgments", () => {
    expect(composeMonthlyMeta(6)).toBe("今月 6 度、いい判断だった。");
  });
  it("works with 1", () => {
    expect(composeMonthlyMeta(1)).toBe("今月 1 度、いい判断だった。");
  });
});
