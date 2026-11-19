import React, { useState } from "react";
import { HistogramBin } from "../types";

interface Props {
  bins: HistogramBin[];
  title?: string;
  unitLabel?: string;
  color?: string;
}

export const HistogramChart: React.FC<Props> = ({
  bins,
  title,
  unitLabel = "",
  color = "var(--cyan)",
}) => {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  if (!bins || bins.length === 0) {
    return (
      <div
        style={{
          padding: "1.5rem",
          textAlign: "center",
          color: "var(--text-muted)",
        }}
      >
        No histogram data available
      </div>
    );
  }

  const maxPct = Math.max(...bins.map((b) => b.percentage), 1);

  return (
    <div style={{ marginTop: "0.5rem" }}>
      {title && (
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "baseline",
            marginBottom: "0.25rem",
          }}
        >
          <span
            style={{
              fontSize: "0.85rem",
              fontWeight: 600,
              color: "var(--text-main)",
            }}
          >
            {title}
          </span>
          {hoveredIndex !== null && (
            <span
              style={{
                fontSize: "0.78rem",
                color: "var(--cyan)",
                fontFamily: "var(--font-mono)",
              }}
            >
              {bins[hoveredIndex].label}:{" "}
              {bins[hoveredIndex].count.toLocaleString()} games (
              {bins[hoveredIndex].percentage}%)
            </span>
          )}
        </div>
      )}

      <div className="histogram-chart">
        {bins.map((bin, idx) => {
          const heightPct = Math.max(4, (bin.percentage / maxPct) * 100);
          const isHovered = hoveredIndex === idx;

          return (
            <div
              key={idx}
              className="histogram-col"
              onMouseEnter={() => setHoveredIndex(idx)}
              onMouseLeave={() => setHoveredIndex(null)}
            >
              <div
                className="histogram-bar"
                style={{
                  height: `${heightPct}%`,
                  backgroundColor: isHovered ? "#fff" : color,
                }}
              />
              <div className="histogram-label" title={bin.label}>
                {bin.label}
              </div>
            </div>
          );
        })}
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginTop: "0.4rem",
          fontSize: "0.7rem",
          color: "var(--text-muted)",
        }}
      >
        <span>Lower bracket</span>
        {unitLabel && <span>Units: {unitLabel}</span>}
        <span>Upper bracket</span>
      </div>
    </div>
  );
};
