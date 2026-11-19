import {
  AlertTriangle,
  Copy,
  Download,
  Plus,
  RefreshCw,
  Trash2,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { PercentileRangeBar } from "../../components/PercentileRangeBar";
import { api } from "../../services/api";
import {
  CohortBenchmarkResult,
  CohortFilter,
  DatasetManifest,
} from "../../types";

interface Props {
  manifest: DatasetManifest | null;
}

const LOCAL_STORAGE_KEY = "steam_cohort_bench_saved_comparison_v1";

const DEFAULT_COHORTS: CohortFilter[] = [
  {
    id: "c1",
    name: "Action Games",
    color: "#38bdf8",
    genres: ["Action"],
    release_year_min: 2020,
    quality_mode: "exclude_playtests",
  },
  {
    id: "c2",
    name: "Strategy Games",
    color: "#818cf8",
    genres: ["Strategy"],
    release_year_min: 2020,
    quality_mode: "exclude_playtests",
  },
  {
    id: "c3",
    name: "Simulation Games",
    color: "#34d399",
    genres: ["Simulation"],
    release_year_min: 2020,
    quality_mode: "exclude_playtests",
  },
];

export const CohortComparisonPage: React.FC<Props> = ({ manifest }) => {
  const [cohorts, setCohorts] = useState<CohortFilter[]>(() => {
    try {
      const saved = localStorage.getItem(LOCAL_STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        if (Array.isArray(parsed) && parsed.length > 0) return parsed;
      }
    } catch {
      // fallback
    }
    return DEFAULT_COHORTS;
  });

  const [benchmarkResults, setBenchmarkResults] = useState<
    CohortBenchmarkResult[]
  >([]);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(cohorts));
    runComparison();
  }, [cohorts]);

  const runComparison = async () => {
    if (cohorts.length === 0) return;
    setLoading(true);
    setError(null);
    try {
      const res = await api.compareCohorts(cohorts);
      setBenchmarkResults(res.cohorts);
      setWarnings(res.warnings || []);
    } catch (err: any) {
      setError(err.message || "Comparison failed");
    } finally {
      setLoading(false);
    }
  };

  const handleAddCohort = () => {
    if (cohorts.length >= 4) return;
    const colors = ["#38bdf8", "#818cf8", "#34d399", "#fbbf24"];
    const nextIdx = cohorts.length + 1;
    const newCohort: CohortFilter = {
      id: `c${Date.now()}`,
      name: `Cohort ${nextIdx}`,
      color: colors[cohorts.length % colors.length],
      genres: ["Indie"],
      release_year_min: 2020,
      quality_mode: "exclude_playtests",
    };
    setCohorts([...cohorts, newCohort]);
  };

  const handleDuplicateCohort = (idx: number) => {
    if (cohorts.length >= 4) return;
    const src = cohorts[idx];
    const copy: CohortFilter = {
      ...src,
      id: `c${Date.now()}`,
      name: `${src.name} (Copy)`,
    };
    setCohorts([...cohorts, copy]);
  };

  const handleRemoveCohort = (idx: number) => {
    if (cohorts.length <= 1) return;
    setCohorts(cohorts.filter((_, i) => i !== idx));
  };

  const handleUpdateCohort = (idx: number, updated: Partial<CohortFilter>) => {
    const list = [...cohorts];
    list[idx] = { ...list[idx], ...updated };
    setCohorts(list);
  };

  const exportComparisonJSON = () => {
    const dataStr =
      "data:text/json;charset=utf-8," +
      encodeURIComponent(
        JSON.stringify({ cohorts, benchmarkResults }, null, 2),
      );
    const a = document.createElement("a");
    a.href = dataStr;
    a.download = "steam-cohort-comparison.json";
    a.click();
  };

  const genres = manifest?.available_genres || [
    "Action",
    "Adventure",
    "Casual",
    "Indie",
    "RPG",
    "Simulation",
    "Strategy",
  ];

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
          <h1>Compare Cohorts</h1>
          <p className="lead" style={{ marginBottom: 0 }}>
            Side-by-side distribution benchmarking for up to 4 distinct market
            cohorts using identical statistical criteria.
          </p>
        </div>

        <div style={{ display: "flex", gap: "0.75rem" }}>
          {cohorts.length < 4 && (
            <button className="btn btn-secondary" onClick={handleAddCohort}>
              <Plus size={14} />
              Add Cohort ({cohorts.length}/4)
            </button>
          )}
          <button className="btn btn-secondary" onClick={exportComparisonJSON}>
            <Download size={14} />
            Export JSON
          </button>
          <button className="btn btn-primary" onClick={runComparison}>
            <RefreshCw size={14} />
            Recalculate
          </button>
        </div>
      </div>

      {warnings.length > 0 && (
        <div className="alert alert-warning">
          <AlertTriangle size={18} />
          <div>
            <strong>Comparison Caveats:</strong>
            <ul style={{ paddingLeft: "1.25rem", marginTop: "0.25rem" }}>
              {warnings.map((w, idx) => (
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

      {/* Cohort Editor Panels */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: `repeat(${cohorts.length}, 1fr)`,
          gap: "1rem",
          marginBottom: "2rem",
        }}
      >
        {cohorts.map((c, idx) => (
          <div
            key={c.id || idx}
            className="card"
            style={{ borderTop: `3px solid ${c.color || "var(--cyan)"}` }}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: "0.75rem",
              }}
            >
              <input
                className="input"
                style={{ fontWeight: 600, fontSize: "0.9rem", width: "70%" }}
                value={c.name || ""}
                onChange={(e) =>
                  handleUpdateCohort(idx, { name: e.target.value })
                }
              />
              <div style={{ display: "flex", gap: "0.25rem" }}>
                <button
                  className="btn btn-secondary btn-sm"
                  title="Duplicate cohort"
                  onClick={() => handleDuplicateCohort(idx)}
                  disabled={cohorts.length >= 4}
                >
                  <Copy size={12} />
                </button>
                {cohorts.length > 1 && (
                  <button
                    className="btn btn-secondary btn-sm"
                    title="Remove cohort"
                    onClick={() => handleRemoveCohort(idx)}
                  >
                    <Trash2 size={12} color="var(--rose)" />
                  </button>
                )}
              </div>
            </div>

            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "0.5rem",
                fontSize: "0.8rem",
              }}
            >
              <div>
                <label className="filter-label">Genre</label>
                <select
                  className="select"
                  style={{ width: "100%" }}
                  value={c.genres?.[0] || ""}
                  onChange={(e) =>
                    handleUpdateCohort(idx, {
                      genres: e.target.value ? [e.target.value] : [],
                    })
                  }
                >
                  {genres.map((g) => (
                    <option key={g} value={g}>
                      {g}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="filter-label">Price Filter</label>
                <select
                  className="select"
                  style={{ width: "100%" }}
                  value={c.price_type || "all"}
                  onChange={(e) =>
                    handleUpdateCohort(idx, {
                      price_type: e.target.value as any,
                    })
                  }
                >
                  <option value="all">All (Free + Paid)</option>
                  <option value="paid">Paid Only</option>
                  <option value="free">Free Only</option>
                </select>
              </div>

              <div>
                <label className="filter-label">Release Vintage</label>
                <select
                  className="select"
                  style={{ width: "100%" }}
                  value={c.release_year_min || ""}
                  onChange={(e) =>
                    handleUpdateCohort(idx, {
                      release_year_min: e.target.value
                        ? parseInt(e.target.value, 10)
                        : undefined,
                    })
                  }
                >
                  <option value="">All Years</option>
                  <option value="2024">2024 - 2025</option>
                  <option value="2020">2020 - 2025</option>
                  <option value="2015">2015 - 2025</option>
                </select>
              </div>
            </div>
          </div>
        ))}
      </div>

      {loading && (
        <div
          style={{
            padding: "2rem",
            textAlign: "center",
            color: "var(--text-dim)",
          }}
        >
          Computing comparative distributions across DuckDB...
        </div>
      )}

      {/* Comparative Results Tables and Visualizations */}
      {benchmarkResults.length > 0 && !loading && (
        <>
          {/* 1. Market Size & Sample Counts */}
          <div className="card" style={{ marginBottom: "1.75rem" }}>
            <h2>1. Cohort Sample Size & Universe Coverage</h2>
            <div
              className="data-table-container"
              style={{ marginTop: "0.75rem" }}
            >
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Cohort</th>
                    <th>Matching Games</th>
                    <th>Paid Games</th>
                    <th>Free Games</th>
                    <th>Linux Support %</th>
                    <th>Median Owners Bracket</th>
                  </tr>
                </thead>
                <tbody>
                  {benchmarkResults.map((r, i) => (
                    <tr key={i}>
                      <td>
                        <strong style={{ color: r.color || "var(--cyan)" }}>
                          {r.cohort_name}
                        </strong>
                      </td>
                      <td style={{ fontFamily: "var(--font-mono)" }}>
                        {r.market_size.total_games.toLocaleString()}
                      </td>
                      <td style={{ fontFamily: "var(--font-mono)" }}>
                        {r.pricing.paid_games_count.toLocaleString()} (
                        {r.pricing.paid_games_percentage}%)
                      </td>
                      <td style={{ fontFamily: "var(--font-mono)" }}>
                        {r.pricing.free_games_count.toLocaleString()} (
                        {r.pricing.free_games_percentage}%)
                      </td>
                      <td style={{ fontFamily: "var(--font-mono)" }}>
                        {r.platforms.linux_percentage}%
                      </td>
                      <td>
                        <span className="badge badge-cyan">
                          {r.audience.median_ownership_bracket}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* 2. Paid Price Percentile Comparison */}
          <div className="card" style={{ marginBottom: "1.75rem" }}>
            <h2>2. Base Price Percentile Benchmark (P25 - Median - P75)</h2>
            <p
              style={{
                fontSize: "0.8rem",
                color: "var(--text-dim)",
                marginBottom: "1rem",
              }}
            >
              Evaluates base pricing without distorting with free-to-play games
              or expensive software outliers.
            </p>
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "1.25rem",
              }}
            >
              {benchmarkResults.map((r, i) => (
                <div key={i}>
                  <div
                    style={{
                      fontWeight: 600,
                      color: r.color || "var(--cyan)",
                      fontSize: "0.9rem",
                      marginBottom: "0.25rem",
                    }}
                  >
                    {r.cohort_name} (Sample:{" "}
                    {r.pricing.paid_games_count.toLocaleString()} paid games)
                  </div>
                  <PercentileRangeBar
                    stats={r.pricing.paid_price_stats}
                    unitPrefix="$"
                  />
                </div>
              ))}
            </div>
          </div>

          {/* 3. Median Playtime Comparison */}
          <div className="card" style={{ marginBottom: "1.75rem" }}>
            <h2>3. Reported Median Playtime Benchmark (Hours)</h2>
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "1.25rem",
              }}
            >
              {benchmarkResults.map((r, i) => (
                <div key={i}>
                  <div
                    style={{
                      fontWeight: 600,
                      color: r.color || "var(--cyan)",
                      fontSize: "0.9rem",
                      marginBottom: "0.25rem",
                    }}
                  >
                    {r.cohort_name} (Sample:{" "}
                    {r.playtime.playtime_forever_stats_hours?.sample_count.toLocaleString()}{" "}
                    titles)
                  </div>
                  <PercentileRangeBar
                    stats={r.playtime.playtime_forever_stats_hours}
                  />
                </div>
              ))}
            </div>
          </div>

          {/* 4. Platform Comparison */}
          <div className="card" style={{ marginBottom: "1.75rem" }}>
            <h2>4. Platform Support Rates by Cohort</h2>
            <div
              className="data-table-container"
              style={{ marginTop: "0.75rem" }}
            >
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Cohort</th>
                    <th>Windows %</th>
                    <th>macOS %</th>
                    <th>Linux %</th>
                    <th>All 3 OS %</th>
                  </tr>
                </thead>
                <tbody>
                  {benchmarkResults.map((r, i) => {
                    const all3 = r.platforms.platform_combinations.find((c) =>
                      c.platforms.includes("All 3"),
                    );
                    return (
                      <tr key={i}>
                        <td
                          style={{
                            color: r.color || "var(--cyan)",
                            fontWeight: 600,
                          }}
                        >
                          {r.cohort_name}
                        </td>
                        <td style={{ fontFamily: "var(--font-mono)" }}>
                          {r.platforms.windows_percentage}%
                        </td>
                        <td style={{ fontFamily: "var(--font-mono)" }}>
                          {r.platforms.mac_percentage}%
                        </td>
                        <td style={{ fontFamily: "var(--font-mono)" }}>
                          {r.platforms.linux_percentage}%
                        </td>
                        <td style={{ fontFamily: "var(--font-mono)" }}>
                          {all3?.percentage || 0}%
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
