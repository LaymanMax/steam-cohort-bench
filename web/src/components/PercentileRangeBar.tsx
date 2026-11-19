import React from "react";
import { PercentileMetric } from "../types";

interface Props {
  stats: PercentileMetric;
  title?: string;
  unitPrefix?: string;
}

export const PercentileRangeBar: React.FC<Props> = ({
  stats,
  title,
  unitPrefix = "",
}) => {
  if (
    !stats ||
    stats.p50 === null ||
    stats.p50 === undefined ||
    stats.sample_count === 0
  ) {
    return (
      <div
        className="percentile-container"
        style={{
          padding: "1rem",
          textAlign: "center",
          color: "var(--text-muted)",
        }}
      >
        No valid observations for {stats?.metric || "metric"} (Sample: 0)
      </div>
    );
  }

  const p25 = stats.p25 ?? 0;
  const p50 = stats.p50 ?? 0;
  const p75 = stats.p75 ?? 0;
  const min = stats.min ?? Math.max(0, p25 * 0.5);
  const max = stats.max ?? Math.max(p75 * 1.5, p75 + 1);

  // Scale for rendering percentages along track
  const span = Math.max(1e-6, max - min);
  const p25Pos = Math.max(0, Math.min(100, ((p25 - min) / span) * 100));
  const p50Pos = Math.max(0, Math.min(100, ((p50 - min) / span) * 100));
  const p75Pos = Math.max(0, Math.min(100, ((p75 - min) / span) * 100));
  const iqrWidth = Math.max(2, p75Pos - p25Pos);

  const formatVal = (v: number | null | undefined) => {
    if (v === null || v === undefined) return "N/A";
    return `${unitPrefix}${v.toFixed(2)} ${stats.unit}`;
  };

  return (
    <div className="percentile-container">
      {title && (
        <div
          style={{
            fontSize: "0.85rem",
            fontWeight: 600,
            color: "var(--text-main)",
            marginBottom: "0.5rem",
          }}
        >
          {title}
        </div>
      )}

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "0.75rem",
        }}
      >
        <div style={{ display: "flex", gap: "1rem" }}>
          <div className="percentile-stat-pill">
            <span
              style={{
                fontSize: "0.68rem",
                color: "var(--text-muted)",
                textTransform: "uppercase",
              }}
            >
              25th Percentile
            </span>
            <span style={{ fontWeight: 600, color: "var(--cyan)" }}>
              {formatVal(p25)}
            </span>
          </div>
          <div
            className="percentile-stat-pill"
            style={{ borderColor: "rgba(255,255,255,0.2)" }}
          >
            <span
              style={{
                fontSize: "0.68rem",
                color: "var(--text-muted)",
                textTransform: "uppercase",
              }}
            >
              Median (P50)
            </span>
            <span style={{ fontWeight: 700, color: "#fff" }}>
              {formatVal(p50)}
            </span>
          </div>
          <div className="percentile-stat-pill">
            <span
              style={{
                fontSize: "0.68rem",
                color: "var(--text-muted)",
                textTransform: "uppercase",
              }}
            >
              75th Percentile
            </span>
            <span style={{ fontWeight: 600, color: "var(--cyan)" }}>
              {formatVal(p75)}
            </span>
          </div>
        </div>

        <div
          style={{
            textAlign: "right",
            fontSize: "0.75rem",
            color: "var(--text-dim)",
            fontFamily: "var(--font-mono)",
          }}
        >
          <div>Valid N: {stats.sample_count.toLocaleString()}</div>
          {stats.missing_count > 0 && (
            <div style={{ color: "var(--amber)" }}>
              Missing: {stats.missing_count.toLocaleString()}
            </div>
          )}
        </div>
      </div>

      {/* Graphical Range Bar */}
      <div className="percentile-bar-track">
        {/* Interquartile Range (IQR) box */}
        <div
          className="percentile-iqr-box"
          style={{ left: `${p25Pos}%`, width: `${iqrWidth}%` }}
          title={`Interquartile Range (IQR: P25-P75): ${formatVal(p25)} to ${formatVal(p75)}`}
        />
        {/* Median tick marker */}
        <div
          className="percentile-median-mark"
          style={{ left: `${p50Pos}%` }}
          title={`Median (P50): ${formatVal(p50)}`}
        />
      </div>

      <div className="percentile-labels">
        <span>Min: {formatVal(min)}</span>
        <span style={{ color: "var(--text-muted)", fontSize: "0.7rem" }}>
          IQR: P25 to P75
        </span>
        <span>Max: {formatVal(max)}</span>
      </div>
    </div>
  );
};
