import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { HistogramBin } from "../types";
import { HistogramChart } from "./HistogramChart";

describe("HistogramChart Component", () => {
  it("renders histogram bins and labels", () => {
    const bins: HistogramBin[] = [
      { lower: 0, upper: 5, label: "$0 - $5", count: 120, percentage: 40 },
      { lower: 5, upper: 10, label: "$5 - $10", count: 90, percentage: 30 },
      { lower: 10, upper: 20, label: "$10 - $20", count: 60, percentage: 20 },
      { lower: 20, upper: 50, label: "$20+", count: 30, percentage: 10 },
    ];

    render(
      <HistogramChart
        bins={bins}
        title="Test Price Histogram"
        unitLabel="USD"
      />,
    );

    expect(screen.getByText("Test Price Histogram")).toBeDefined();
    expect(screen.getByText("$0 - $5")).toBeDefined();
    expect(screen.getByText("$20+")).toBeDefined();
  });
});
