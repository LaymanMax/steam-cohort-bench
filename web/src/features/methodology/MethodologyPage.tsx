import { BookOpen, CheckCircle, Cpu, ShieldAlert } from "lucide-react";
import React from "react";

export const MethodologyPage: React.FC = () => {
  return (
    <div className="page-body" style={{ maxWidth: "960px" }}>
      <div style={{ marginBottom: "2rem" }}>
        <h1>Statistical Methodology & Data Provenance</h1>
        <p className="lead">
          Full transparency into dataset origin, transformation rules,
          percentile math, and analytical caveats.
        </p>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "2rem" }}>
        {/* Provenance */}
        <div className="card">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
              marginBottom: "1rem",
            }}
          >
            <BookOpen size={18} color="var(--cyan)" />
            <h2>1. Source Dataset & Provenance</h2>
          </div>
          <p
            style={{
              color: "var(--text-dim)",
              marginBottom: "0.75rem",
              fontSize: "0.9rem",
            }}
          >
            <code>steam-cohort-bench</code> consumes the open-source dataset{" "}
            <strong>FronkonGames/steam-games-dataset</strong> hosted on Hugging
            Face, licensed under the MIT License. Data is scraped from public
            Steam Web API endpoints and augmented with SteamSpy ownership
            estimations.
          </p>
          <div className="alert alert-info" style={{ marginBottom: 0 }}>
            <div>
              <strong>Count Discrepancy Reconciliation:</strong> The Hugging
              Face README cites 148,909 published games. However, the official
              Parquet export (<code>data/train-00000-of-00001.parquet</code>)
              contains <strong>124,146</strong> records. The application
              enforces the actual downloaded Parquet file as the immutable
              ground truth and records this discrepancy in all data quality
              reports.
            </div>
          </div>
        </div>

        {/* Statistical Percentiles */}
        <div className="card">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
              marginBottom: "1rem",
            }}
          >
            <Cpu size={18} color="var(--indigo)" />
            <h2>2. Percentiles & Quantile Interpolation</h2>
          </div>
          <p
            style={{
              color: "var(--text-dim)",
              marginBottom: "0.75rem",
              fontSize: "0.9rem",
            }}
          >
            We strictly avoid reducing market metrics to misleading averages
            that can be skewed by ultra-expensive software or extreme outliers.
            All numeric benchmarks use continuous percentiles:
          </p>
          <ul
            style={{
              paddingLeft: "1.5rem",
              color: "var(--text-dim)",
              fontSize: "0.9rem",
              display: "flex",
              flexDirection: "column",
              gap: "0.5rem",
            }}
          >
            <li>
              <strong>25th Percentile (P25):</strong> Lower quartile. 25% of
              cohort games fall at or below this value.
            </li>
            <li>
              <strong>50th Percentile (P50 / Median):</strong> The central
              tendency. Exactly half of the observed games fall above and half
              below.
            </li>
            <li>
              <strong>75th Percentile (P75):</strong> Upper quartile. 75% of
              games fall at or below this value.
            </li>
            <li>
              <strong>Interquartile Range (IQR):</strong> The interval between
              P25 and P75, capturing the core 50% of the market segment.
            </li>
            <li>
              <strong>Quantile Method:</strong> Standard continuous linear
              interpolation (R-7 / DuckDB <code>quantile_cont</code>).
            </li>
          </ul>
        </div>

        {/* Ownership and Reviews */}
        <div className="card">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
              marginBottom: "1rem",
            }}
          >
            <CheckCircle size={18} color="var(--emerald)" />
            <h2>3. Metric Normalization & Field Semantics</h2>
          </div>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "1rem",
              fontSize: "0.9rem",
              color: "var(--text-dim)",
            }}
          >
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Estimated Owners:
              </strong>{" "}
              The source dataset contains SteamSpy text intervals (e.g.{" "}
              <code>0 - 20000</code>, <code>20000 - 50000</code>). These are
              never presented as exact counts. The <code>0 - 0</code> interval
              is preserved as a distinct zero-state rather than dropped.
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Review Counts:
              </strong>{" "}
              The <code>reviews</code> column in the raw dataset contains
              editorial critic quotes. Total player reviews are derived directly
              from <code>positive + negative</code> player review counts.
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>Pricing:</strong>{" "}
              Free-to-play games ($0.00) are explicitly separated from paid
              games. Paid price percentiles consider only paid games ($0.01+) so
              free titles do not depress commercial pricing percentiles.
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>Playtime:</strong>{" "}
              Reported Steam median playtime is captured in minutes and
              converted to hours (<code>minutes / 60.0</code>) for user display.
            </div>
            <div>
              <strong style={{ color: "var(--text-main)" }}>
                Platform Denominators:
              </strong>{" "}
              Platform support percentages (Windows, macOS, Linux) are computed
              against games with verified platform metadata, not artificially
              deflated by unknown records.
            </div>
          </div>
        </div>

        {/* Caveats and Limitations */}
        <div className="card">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
              marginBottom: "1rem",
            }}
          >
            <ShieldAlert size={18} color="var(--amber)" />
            <h2>4. Known Limitations & Analytical Caveats</h2>
          </div>
          <ul
            style={{
              paddingLeft: "1.5rem",
              color: "var(--text-dim)",
              fontSize: "0.9rem",
              display: "flex",
              flexDirection: "column",
              gap: "0.75rem",
            }}
          >
            <li>
              <strong>Estimated Ownership Uncertainty:</strong> SteamSpy
              estimates are derived from public profile sampling and carry
              confidence bands, especially for games with under 20,000 owners.
            </li>
            <li>
              <strong>Survivorship & Archival Bias:</strong> Published titles
              remain in the historical catalog even if delisted or abandoned.
              Use release vintage and review volume filters to isolate
              contemporary active cohorts.
            </li>
            <li>
              <strong>Correlation is Not Causation:</strong> Cohort comparisons
              demonstrate historical market distribution, not predictive success
              recipes.
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
};
