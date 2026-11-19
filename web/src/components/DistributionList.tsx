import React from "react";
import { DistributionItem } from "../types";

interface Props {
  items: DistributionItem[];
  title?: string;
  maxItems?: number;
  barColor?: string;
}

export const DistributionList: React.FC<Props> = ({
  items,
  title,
  maxItems = 10,
  barColor = "var(--cyan)",
}) => {
  if (!items || items.length === 0) {
    return (
      <div
        style={{
          padding: "1rem",
          textAlign: "center",
          color: "var(--text-muted)",
        }}
      >
        No distribution records available
      </div>
    );
  }

  const displayed = items.slice(0, maxItems);
  const maxPct = Math.max(...displayed.map((i) => i.percentage), 1);

  return (
    <div>
      {title && (
        <div
          style={{
            fontSize: "0.85rem",
            fontWeight: 600,
            color: "var(--text-dim)",
            marginBottom: "0.5rem",
          }}
        >
          {title}
        </div>
      )}
      <div className="dist-list">
        {displayed.map((item) => {
          const widthPct = Math.max(3, (item.percentage / maxPct) * 100);
          return (
            <div key={item.key} className="dist-item">
              <div className="dist-item-header">
                <span className="dist-item-name">{item.label}</span>
                <span className="dist-item-stats">
                  {item.count.toLocaleString()}{" "}
                  <span style={{ color: "var(--text-muted)" }}>
                    ({item.percentage}%)
                  </span>
                </span>
              </div>
              <div className="dist-track">
                <div
                  className="dist-fill"
                  style={{ width: `${widthPct}%`, backgroundColor: barColor }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
