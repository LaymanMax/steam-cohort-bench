import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PercentileMetric } from "../types";
import { PercentileRangeBar } from "./PercentileRangeBar";

describe("PercentileRangeBar Component", () => {
  it("renders percentile values correctly", () => {
    const stats: PercentileMetric = {
      metric: "paid_price",
      unit: "USD",
      sample_count: 500,
      missing_count: 10,
      p25: 4.99,
      p50: 9.99,
      p75: 19.99,
      min: 0.99,
      max: 59.99,
    };

    render(
      <PercentileRangeBar
        stats={stats}
        title="Base Pricing Benchmark"
        unitPrefix="$"
      />,
    );

    expect(screen.getByText("Base Pricing Benchmark")).toBeDefined();
    expect(screen.getByText("$9.99 USD")).toBeDefined();
    expect(screen.getByText("$4.99 USD")).toBeDefined();
    expect(screen.getByText("$19.99 USD")).toBeDefined();
    expect(screen.getByText("Valid N: 500")).toBeDefined();
  });

  it("renders empty fallback when sample count is 0", () => {
    const emptyStats: PercentileMetric = {
      metric: "empty_metric",
      unit: "USD",
      sample_count: 0,
      missing_count: 50,
      p25: null,
      p50: null,
      p75: null,
    };

    render(<PercentileRangeBar stats={emptyStats} />);
    expect(screen.getByText(/No valid observations/i)).toBeDefined();
  });
});
