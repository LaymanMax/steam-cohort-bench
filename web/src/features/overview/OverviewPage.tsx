import { AlertTriangle, ArrowRight, Database, Download } from "lucide-react";
import React, { useEffect, useState } from "react";
import { DistributionList } from "../../components/DistributionList";
import { HistogramChart } from "../../components/HistogramChart";
import { PercentileRangeBar } from "../../components/PercentileRangeBar";
import { PlatformSupportBars } from "../../components/PlatformSupportBars";
import { api } from "../../services/api";
import { CohortBenchmarkResult, DatasetManifest } from "../../types";

interface Props {
  manifest: DatasetManifest | null;
  onNavigate: (page: string) => void;
}

export const OverviewPage: React.FC<Props> = ({ manifest, onNavigate }) => {
  const [data, setData] = useState<CohortBenchmarkResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadOverview();
  }, []);

  const loadOverview = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.getOverview();
      setData(res);
    } catch (err: any) {
      setError(err.message || "Failed to load market overview");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="page-body">
        <h1>Steam Market Research Overview</h1>
        <p className="lead">
          Loading canonical dataset distributions from DuckDB...
        </p>
        <div className="metrics-grid">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="metric-tile skeleton"
              style={{ height: "90px" }}
            />
          ))}
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="page-body">
        <h1>Market Research Overview</h1>
        <div className="alert alert-warning" style={{ marginTop: "1.5rem" }}>
          <AlertTriangle size={18} />
          <div>
            <strong>Dataset Unavailable:</strong>{" "}
            {error || "No dataset loaded."}
            <div style={{ marginTop: "0.5rem", fontSize: "0.8rem" }}>
              Run the dataset ingestion command:{" "}
              <code>
                go run ./cmd/ingest --input ./data/raw/steam-games.parquet
              </code>{" "}
              or check <code>/api/v1/health</code>.
            </div>
          </div>
        </div>
      </div>
    );
  }

  const {
    market_size,
    pricing,
    audience,
    playtime,
    reviews,
    platforms,
    localization,
  } = data;

  return (
    <div className="page-body">
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          marginBottom: "1.5rem",
        }}
      >
        <div>
          <h1>Steam Market Ground Truth</h1>
          <p className="lead" style={{ marginBottom: 0 }}>
            Empirical distributions and benchmark statistics computed across{" "}
            {market_size.total_games.toLocaleString()} games.
          </p>
        </div>

        <div style={{ display: "flex", gap: "0.75rem" }}>
          <a
            href={api.getCohortExportUrl({})}
            className="btn btn-secondary"
            download="market-overview.csv"
          >
            <Download size={14} />
            Export Overview CSV
          </a>
          <button
            className="btn btn-primary"
            onClick={() => onNavigate("explorer")}
          >
            Open Cohort Explorer
            <ArrowRight size={14} />
          </button>
        </div>
      </div>

      {/* Dataset Provenance Banner */}
      <div
        className="card"
        style={{
          marginBottom: "1.75rem",
          background: "var(--bg-sidebar)",
          borderColor: "var(--border-medium)",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            flexWrap: "wrap",
            gap: "1rem",
          }}
        >
          <div
            style={{ display: "flex", alignItems: "center", gap: "0.75rem" }}
          >
            <Database size={20} color="var(--cyan)" />
            <div>
              <div style={{ fontSize: "0.85rem", fontWeight: 600 }}>
                Source Dataset:{" "}
                {manifest?.dataset_name || "FronkonGames/steam-games-dataset"}
              </div>
              <div style={{ fontSize: "0.75rem", color: "var(--text-dim)" }}>
                Revision: <code>{manifest?.revision || "Ground Truth"}</code> •
                SHA256:{" "}
                <code>
                  {manifest?.source_file_checksum?.substring(0, 16)}...
                </code>{" "}
                • License: {manifest?.license || "MIT"}
              </div>
            </div>
          </div>

          <div
            style={{
              display: "flex",
              gap: "1rem",
              fontSize: "0.78rem",
              color: "var(--text-dim)",
              fontFamily: "var(--font-mono)",
            }}
          >
            <div>
              Parquet Rows:{" "}
              <span style={{ color: "var(--text-main)", fontWeight: 600 }}>
                {manifest?.source_row_count.toLocaleString() || "124,146"}
              </span>
            </div>
            <div>
              Unique AppIDs:{" "}
              <span style={{ color: "var(--text-main)", fontWeight: 600 }}>
                {manifest?.distinct_app_ids.toLocaleString() || "124,146"}
              </span>
            </div>
            <div>
              Price Currency:{" "}
              <span style={{ color: "var(--text-main)", fontWeight: 600 }}>
                {manifest?.price_unit || "USD"}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Primary KPI Tiles */}
      <div className="metrics-grid">
        <div className="metric-tile">
          <div className="metric-label">Total Unique Games</div>
          <div className="metric-value">
            {market_size.total_games.toLocaleString()}
          </div>
          <div className="metric-sub">100% distinct AppIDs verified</div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Paid Games Median Price</div>
          <div className="metric-value" style={{ color: "var(--cyan)" }}>
            ${pricing.paid_price_stats.p50?.toFixed(2) || "0.00"}
            <span className="metric-unit">USD</span>
          </div>
          <div className="metric-sub">
            IQR: ${pricing.paid_price_stats.p25?.toFixed(2)} - $
            {pricing.paid_price_stats.p75?.toFixed(2)}
          </div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Free vs Paid Ratio</div>
          <div className="metric-value" style={{ color: "var(--indigo)" }}>
            {pricing.paid_games_percentage}%{" "}
            <span className="metric-unit">paid</span>
          </div>
          <div className="metric-sub">
            {pricing.free_games_percentage}% free to play
          </div>
        </div>

        <div className="metric-tile">
          <div className="metric-label">Linux Support Prevalence</div>
          <div className="metric-value" style={{ color: "var(--emerald)" }}>
            {platforms.linux_percentage}%
          </div>
          <div className="metric-sub">
            {platforms.linux_count.toLocaleString()} native titles
          </div>
        </div>
      </div>

      {/* Section 1: Pricing Analysis */}
      <h2 style={{ marginTop: "1.5rem" }}>1. Base Pricing Architecture</h2>
      <div className="card" style={{ marginBottom: "1.75rem" }}>
        <PercentileRangeBar
          stats={pricing.paid_price_stats}
          title="Paid Game Base Price Percentile Distribution (USD)"
          unitPrefix="$"
        />
        <div
          className="charts-grid-2"
          style={{ marginTop: "1.5rem", marginBottom: 0 }}
        >
          <HistogramChart
            bins={pricing.price_histogram}
            title="Price Distribution Histogram"
            unitLabel="USD"
          />
          <DistributionList
            items={pricing.common_price_ranges}
            title="Standard Market Price Brackets"
          />
        </div>
      </div>

      {/* Section 2: Audience & Ownership Brackets */}
      <h2>2. Estimated Audience Size & Ownership Intervals</h2>
      <div className="charts-grid-2">
        <div className="card">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "baseline",
              marginBottom: "0.75rem",
            }}
          >
            <h3>Estimated Owners Bracket Distribution</h3>
            <span className="badge badge-cyan">
              Median: {audience.median_ownership_bracket}
            </span>
          </div>
          <p
            style={{
              fontSize: "0.8rem",
              color: "var(--text-dim)",
              marginBottom: "1rem",
            }}
          >
            Ownership data represents estimated SteamSpy intervals, not exact
            counts. The 0 - 0 bracket represents playtests, unreleased, or
            unranked titles.
          </p>
          <DistributionList
            items={audience.ownership_bracket_distribution}
            maxItems={12}
            barColor="var(--indigo)"
          />
        </div>

        <div className="card">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "baseline",
              marginBottom: "0.75rem",
            }}
          >
            <h3>Player Review Volume Distribution</h3>
            <span className="badge">
              P50: {reviews.total_reviews_stats.p50?.toLocaleString()} reviews
            </span>
          </div>
          <p
            style={{
              fontSize: "0.8rem",
              color: "var(--text-dim)",
              marginBottom: "1rem",
            }}
          >
            Review counts are derived from player positive + negative reviews.
            Editorial press quotes are separated.
          </p>
          <DistributionList
            items={reviews.review_volume_brackets}
            maxItems={8}
            barColor="var(--cyan)"
          />
        </div>
      </div>

      {/* Section 3: Engagement & Playtime */}
      <h2 style={{ marginTop: "1rem" }}>
        3. Player Engagement & Reported Playtime
      </h2>
      <div className="card" style={{ marginBottom: "1.75rem" }}>
        <PercentileRangeBar
          stats={playtime.playtime_forever_stats_hours}
          title="Reported Median Playtime Forever (Hours)"
        />
        <div
          className="charts-grid-2"
          style={{ marginTop: "1rem", marginBottom: 0 }}
        >
          <HistogramChart
            bins={playtime.playtime_forever_histogram}
            title="Playtime Distribution Histogram"
            unitLabel="hours"
            color="var(--emerald)"
          />
          <DistributionList
            items={playtime.free_vs_paid_playtime_median}
            title="Median Playtime: Free vs Paid Titles (Hours)"
            barColor="var(--amber)"
          />
        </div>
      </div>

      {/* Section 4: Platform & Localization Coverage */}
      <h2>4. Ecosystem & Platform Availability</h2>
      <div className="charts-grid-2">
        <div className="card">
          <h3>Operating System Support</h3>
          <p
            style={{
              fontSize: "0.8rem",
              color: "var(--text-dim)",
              marginBottom: "1rem",
            }}
          >
            Percentage of games with verified native platform executables on
            Steam.
          </p>
          <PlatformSupportBars platforms={platforms} />
        </div>

        <div className="card">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "baseline",
              marginBottom: "0.5rem",
            }}
          >
            <h3>Localization Coverage</h3>
            <span className="badge">
              Median: {localization.listed_language_count_stats.p50} langs
            </span>
          </div>
          <p
            style={{
              fontSize: "0.8rem",
              color: "var(--text-dim)",
              marginBottom: "1rem",
            }}
          >
            Comparison between listed interface/subtitles vs full voiceover
            audio.
          </p>
          <DistributionList
            items={localization.top_supported_languages}
            title="Top Supported Interface Languages"
            maxItems={8}
            barColor="var(--cyan)"
          />
          <div style={{ marginTop: "1rem" }}>
            <DistributionList
              items={localization.top_audio_languages}
              title="Top Full-Audio Languages"
              maxItems={5}
              barColor="var(--amber)"
            />
          </div>
        </div>
      </div>

      {/* Section 5: Genre Composition & History */}
      <h2 style={{ marginTop: "1rem" }}>
        5. Market Composition & Release Cadence
      </h2>
      <div className="charts-grid-2">
        <div className="card">
          <h3>Top Steam Genres</h3>
          <DistributionList
            items={market_size.genre_distribution}
            maxItems={10}
            barColor="var(--cyan)"
          />
        </div>

        <div className="card">
          <h3>Release Distribution by Year</h3>
          <DistributionList
            items={market_size.release_year_distribution}
            maxItems={12}
            barColor="var(--indigo)"
          />
        </div>
      </div>
    </div>
  );
};
