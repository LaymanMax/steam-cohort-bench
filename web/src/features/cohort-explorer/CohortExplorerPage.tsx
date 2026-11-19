import {
  AlertTriangle,
  Clock,
  DollarSign,
  Download,
  Globe,
  Layers,
  Monitor,
  Users,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { CohortFilterBuilder } from "../../components/CohortFilterBuilder";
import { DistributionList } from "../../components/DistributionList";
import { HistogramChart } from "../../components/HistogramChart";
import { PercentileRangeBar } from "../../components/PercentileRangeBar";
import { PlatformSupportBars } from "../../components/PlatformSupportBars";
import { api } from "../../services/api";
import {
  CohortBenchmarkResult,
  CohortFilter,
  DatasetManifest,
} from "../../types";

interface Props {
  manifest: DatasetManifest | null;
}

export const CohortExplorerPage: React.FC<Props> = ({ manifest }) => {
  const [filter, setFilter] = useState<CohortFilter>({
    name: "Custom Research Cohort",
    genres: ["Strategy"],
    release_year_min: 2020,
    release_year_max: 2025,
    quality_mode: "exclude_playtests",
  });

  const [data, setData] = useState<CohortBenchmarkResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    runBenchmark(filter);
  }, [filter]);

  const runBenchmark = async (currentFilter: CohortFilter) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.benchmarkCohort(currentFilter);
      setData(res);
    } catch (err: any) {
      setError(err.message || "Failed to benchmark cohort");
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setFilter({
      name: "All Games",
      quality_mode: "all",
    });
  };

  return (
    <div className="page-body">
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          marginBottom: "1.25rem",
        }}
      >
        <div>
          <h1>Cohort Explorer</h1>
          <p className="lead" style={{ marginBottom: 0 }}>
            Construct fine-grained peer cohorts across genre, pricing,
            platforms, release vintage, and sample quality.
          </p>
        </div>

        {data && (
          <a
            href={api.getCohortExportUrl(filter)}
            className="btn btn-secondary"
            download="cohort-benchmark.csv"
          >
            <Download size={14} />
            Export Cohort CSV
          </a>
        )}
      </div>

      {/* Filter Builder */}
      <CohortFilterBuilder
        filter={filter}
        manifest={manifest}
        onChange={setFilter}
        onReset={handleReset}
      />

      {/* Warnings & Sample size caveats */}
      {data?.warnings && data.warnings.length > 0 && (
        <div className="alert alert-warning">
          <AlertTriangle size={18} />
          <div>
            <strong>Analytical Caveats:</strong>
            <ul style={{ paddingLeft: "1.25rem", marginTop: "0.25rem" }}>
              {data.warnings.map((w, idx) => (
                <li key={idx}>{w}</li>
              ))}
            </ul>
          </div>
        </div>
      )}

      {error && (
        <div className="alert alert-warning">
          <AlertTriangle size={18} />
          <div>{error}</div>
        </div>
      )}

      {loading && (
        <div className="metrics-grid">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="metric-tile skeleton"
              style={{ height: "90px" }}
            />
          ))}
        </div>
      )}

      {data && !loading && (
        <>
          {/* Cohort Summary Metric Tiles */}
          <div className="metrics-grid">
            <div className="metric-tile">
              <div className="metric-label">Matching Games</div>
              <div className="metric-value" style={{ color: "var(--cyan)" }}>
                {data.market_size.total_games.toLocaleString()}
              </div>
              <div className="metric-sub">
                {data.market_size.distinct_app_ids.toLocaleString()} unique
                AppIDs
              </div>
            </div>

            <div className="metric-tile">
              <div className="metric-label">Paid Games Median Price</div>
              <div className="metric-value">
                ${data.pricing.paid_price_stats.p50?.toFixed(2) || "0.00"}
                <span className="metric-unit">USD</span>
              </div>
              <div className="metric-sub">
                IQR: ${data.pricing.paid_price_stats.p25?.toFixed(2)} - $
                {data.pricing.paid_price_stats.p75?.toFixed(2)}
              </div>
            </div>

            <div className="metric-tile">
              <div className="metric-label">Median Playtime (Forever)</div>
              <div className="metric-value" style={{ color: "var(--emerald)" }}>
                {data.playtime.playtime_forever_stats_hours?.p50?.toFixed(1) ||
                  "0.0"}
                <span className="metric-unit">hours</span>
              </div>
              <div className="metric-sub">
                Valid sample:{" "}
                {data.playtime.playtime_forever_stats_hours?.sample_count.toLocaleString()}
              </div>
            </div>

            <div className="metric-tile">
              <div className="metric-label">Median Ownership Bracket</div>
              <div
                className="metric-value"
                style={{ fontSize: "1.25rem", color: "var(--indigo)" }}
              >
                {data.audience.median_ownership_bracket}
              </div>
              <div className="metric-sub">
                {data.platforms.linux_percentage}% support Linux
              </div>
            </div>
          </div>

          {/* Section: Pricing */}
          <div className="card" style={{ marginBottom: "1.75rem" }}>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
                marginBottom: "0.75rem",
              }}
            >
              <DollarSign size={18} color="var(--cyan)" />
              <h2>Cohort Pricing Distribution</h2>
            </div>
            <PercentileRangeBar
              stats={data.pricing.paid_price_stats}
              title="Paid Game Price Percentiles (USD)"
              unitPrefix="$"
            />
            <div
              className="charts-grid-2"
              style={{ marginTop: "1.25rem", marginBottom: 0 }}
            >
              <HistogramChart
                bins={data.pricing.price_histogram}
                title="Cohort Price Histogram"
                unitLabel="USD"
              />
              <DistributionList
                items={data.pricing.common_price_ranges}
                title="Price Tiers in Cohort"
              />
            </div>
          </div>

          {/* Section: Audience & Reviews */}
          <div className="charts-grid-2">
            <div className="card">
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                  marginBottom: "0.75rem",
                }}
              >
                <Users size={18} color="var(--indigo)" />
                <h2>Estimated Audience Brackets</h2>
              </div>
              <p
                style={{
                  fontSize: "0.8rem",
                  color: "var(--text-dim)",
                  marginBottom: "1rem",
                }}
              >
                Reported SteamSpy brackets for cohort games.
              </p>
              <DistributionList
                items={data.audience.ownership_bracket_distribution}
                maxItems={10}
                barColor="var(--indigo)"
              />
            </div>

            <div className="card">
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                  marginBottom: "0.75rem",
                }}
              >
                <Layers size={18} color="var(--cyan)" />
                <h2>Review Volume Breakdown</h2>
              </div>
              <PercentileRangeBar
                stats={data.reviews.total_reviews_stats}
                title="Player Review Counts (Positive + Negative)"
              />
              <div style={{ marginTop: "1rem" }}>
                <DistributionList
                  items={data.reviews.review_volume_brackets}
                  title="Review Volume Tiers"
                  maxItems={8}
                />
              </div>
            </div>
          </div>

          {/* Section: Playtime & Engagement */}
          <div
            className="card"
            style={{ marginTop: "1.5rem", marginBottom: "1.5rem" }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.5rem",
                marginBottom: "0.75rem",
              }}
            >
              <Clock size={18} color="var(--emerald)" />
              <h2>Reported Median Playtime (Hours)</h2>
            </div>
            <PercentileRangeBar
              stats={data.playtime.playtime_forever_stats_hours}
              title="Playtime Forever Percentile Distribution"
            />
            <div
              className="charts-grid-2"
              style={{ marginTop: "1rem", marginBottom: 0 }}
            >
              <HistogramChart
                bins={data.playtime.playtime_forever_histogram}
                title="Playtime Histogram"
                unitLabel="hours"
                color="var(--emerald)"
              />
              <DistributionList
                items={data.playtime.free_vs_paid_playtime_median}
                title="Free vs Paid Median Playtime"
                barColor="var(--amber)"
              />
            </div>
          </div>

          {/* Section: Platforms & Localization */}
          <div className="charts-grid-2">
            <div className="card">
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                  marginBottom: "0.75rem",
                }}
              >
                <Monitor size={18} color="var(--cyan)" />
                <h2>Platform Availability</h2>
              </div>
              <PlatformSupportBars platforms={data.platforms} />
            </div>

            <div className="card">
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                  marginBottom: "0.75rem",
                }}
              >
                <Globe size={18} color="var(--amber)" />
                <h2>Localization Coverage</h2>
              </div>
              <PercentileRangeBar
                stats={data.localization.listed_language_count_stats}
                title="Languages Supported Per Game"
              />
              <div style={{ marginTop: "1rem" }}>
                <DistributionList
                  items={data.localization.top_supported_languages}
                  title="Top Supported Languages"
                  maxItems={8}
                  barColor="var(--amber)"
                />
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
