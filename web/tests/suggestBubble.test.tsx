import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vitest";

import { SuggestBubble } from "@/components/order/SuggestBubble";

describe("SuggestBubble (LC-SUGGEST-11)", () => {
  test("固定文言「そろそろだろ。」と提案ラベルを表示する (BR-D13 / design spec §3.3)", () => {
    render(
      <SuggestBubble
        plan={{ storeName: "CoCo壱", menuName: "ポークカレー", amount: 1200, category: "food" }}
      />,
    );
    expect(screen.getByTestId("suggest-bubble")).toBeInTheDocument();
    expect(screen.getByText("そろそろだろ。")).toBeInTheDocument();

    const label = screen.getByTestId("suggest-plan-label");
    expect(label).toHaveTextContent("CoCo壱");
    expect(label).toHaveTextContent("1,200"); // toLocaleString のカンマ区切り
  });
});
