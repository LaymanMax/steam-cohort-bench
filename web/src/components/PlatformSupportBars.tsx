import React from "react";
import { PlatformSection } from "../types";

interface Props {
  platforms: PlatformSection;
}

export const PlatformSupportBars: React.FC<Props> = ({ platforms }) => {
  if (!platforms) return null;

  return (
    <div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(3, 1fr)",
          gap: "0.75rem",
          marginBottom: "1.25rem",
        }}
      >
        <div className="metric-tile" style={{ textAlign: "center" }}>
          <div className="metric-label">Windows Support</div>
          <div
            className="metric-value"
            style={{ justifyContent: "center", color: "var(--cyan)" }}
          >
            {platforms.windows_percentage}%
          </div>
          <div className="metric-sub">
            {platforms.windows_count.toLocaleString()} games
          </div>
        </div>

        <div className="metric-tile" style={{ textAlign: "center" }}>
          <div className="metric-label">macOS Support</div>
          <div
            className="metric-value"
            style={{ justifyContent: "center", color: "var(--indigo)" }}
          >
            {platforms.mac_percentage}%
          </div>
          <div className="metric-sub">
            {platforms.mac_count.toLocaleString()} games
          </div>
        </div>

        <div className="metric-tile" style={{ textAlign: "center" }}>
          <div className="metric-label">Linux Support</div>
          <div
            className="metric-value"
            style={{ justifyContent: "center", color: "var(--emerald)" }}
          >
            {platforms.linux_percentage}%
          </div>
          <div className="metric-sub">
            {platforms.linux_count.toLocaleString()} games
          </div>
        </div>
      </div>

      <div
        style={{
          fontSize: "0.82rem",
          fontWeight: 600,
          color: "var(--text-dim)",
          marginBottom: "0.5rem",
        }}
      >
        Common OS Combinations
      </div>
      <div className="dist-list">
        {platforms.platform_combinations?.map((c) => (
          <div key={c.platforms} className="dist-item">
            <div className="dist-item-header">
              <span className="dist-item-name">{c.platforms}</span>
              <span className="dist-item-stats">
                {c.count.toLocaleString()}{" "}
                <span style={{ color: "var(--text-muted)" }}>
                  ({c.percentage}%)
                </span>
              </span>
            </div>
            <div className="dist-track">
              <div
                className="dist-fill"
                style={{
                  width: `${Math.max(2, c.percentage)}%`,
                  backgroundColor: c.platforms.includes("All 3")
                    ? "var(--emerald)"
                    : "var(--cyan)",
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
