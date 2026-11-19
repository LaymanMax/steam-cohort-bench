import {
  AlertCircle,
  DollarSign,
  Globe,
  Monitor,
  Sparkles,
} from "lucide-react";
import React, { useState } from "react";
import { PercentileRangeBar } from "../../components/PercentileRangeBar";
import { api } from "../../services/api";
import {
  DatasetManifest,
  HypotheticalProfile,
  ProfileBenchmarkResult,
} from "../../types";

interface Props {
  manifest: DatasetManifest | null;
}

export const GameProfilePage: React.FC<Props> = ({ manifest }) => {
  const [profile, setProfile] = useState<HypotheticalProfile>({
    name: "Tactical Colony",
    genre: "Strategy",
    secondary_genre: "Simulation",
    release_year: 2026,
    proposed_price: 19.99,
    is_free: false,
    windows: true,
    mac: false,
    linux: true,
    localization_languages: [
      "English",
      "German",
      "French",
      "Japanese",
      "Simplified Chinese",
    ],
    target_owners_bracket: "50K - 100K",
  });

  const [benchmark, setBenchmark] = useState<ProfileBenchmarkResult | null>(
    null,
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const genres = manifest?.available_genres || [
    "Action",
    "Adventure",
    "Casual",
    "Indie",
    "RPG",
    "Simulation",
    "Strategy",
  ];

  const handleBenchmark = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await api.benchmarkProfile(profile);
      setBenchmark(res);
    } catch (err: any) {
      setError(err.message || "Failed to benchmark hypothetical profile");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page-body">
      <div style={{ marginBottom: "1.5rem" }}>
        <h1>Hypothetical Game Profile Benchmark</h1>
        <p className="lead" style={{ marginBottom: 0 }}>
          Define prospective game specifications and evaluate your positioning
          against historical market cohorts.
        </p>
      </div>

      <div className="alert alert-info" style={{ marginBottom: "1.5rem" }}>
        <AlertCircle size={18} />
        <div>
          <strong>Descriptive Context, Not Predictive Guidance:</strong> This
          feature benchmarks your proposed attributes against actual historical
          Steam distributions. It does not calculate success probability or
          endorse specific decisions.
        </div>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "minmax(320px, 380px) 1fr",
          gap: "1.75rem",
          alignItems: "flex-start",
        }}
      >
        {/* Profile Specification Form */}
        <div className="card">
          <div
            style={{
              display: "flex",
              alignItems: "center",
              gap: "0.5rem",
              marginBottom: "1rem",
            }}
          >
            <Sparkles size={18} color="var(--cyan)" />
            <h2>Game Specifications</h2>
          </div>

          <form
            onSubmit={handleBenchmark}
            style={{ display: "flex", flexDirection: "column", gap: "1rem" }}
          >
            <div>
              <label className="filter-label">Project Title</label>
              <input
                className="input"
                style={{ width: "100%" }}
                value={profile.name || ""}
                onChange={(e) =>
                  setProfile({ ...profile, name: e.target.value })
                }
                placeholder="e.g. Tactical Colony"
              />
            </div>

            <div>
              <label className="filter-label">Primary Genre *</label>
              <select
                className="select"
                style={{ width: "100%" }}
                value={profile.genre}
                onChange={(e) =>
                  setProfile({ ...profile, genre: e.target.value })
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
              <label className="filter-label">Secondary Genre (Optional)</label>
              <select
                className="select"
                style={{ width: "100%" }}
                value={profile.secondary_genre || ""}
                onChange={(e) =>
                  setProfile({ ...profile, secondary_genre: e.target.value })
                }
              >
                <option value="">None</option>
                {genres
                  .filter((g) => g !== profile.genre)
                  .map((g) => (
                    <option key={g} value={g}>
                      {g}
                    </option>
                  ))}
              </select>
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "0.75rem",
              }}
            >
              <div>
                <label className="filter-label">Target Vintage</label>
                <input
                  type="number"
                  className="input"
                  style={{ width: "100%" }}
                  value={profile.release_year}
                  onChange={(e) =>
                    setProfile({
                      ...profile,
                      release_year: parseInt(e.target.value, 10) || 2026,
                    })
                  }
                />
              </div>

              <div>
                <label className="filter-label">Proposed Price ($)</label>
                <input
                  type="number"
                  step="0.5"
                  className="input"
                  style={{ width: "100%" }}
                  value={profile.proposed_price}
                  onChange={(e) =>
                    setProfile({
                      ...profile,
                      proposed_price: parseFloat(e.target.value) || 0,
                    })
                  }
                />
              </div>
            </div>

            <div>
              <label className="filter-label">Platform Support</label>
              <div
                style={{
                  display: "flex",
                  gap: "1rem",
                  marginTop: "0.35rem",
                  fontSize: "0.85rem",
                }}
              >
                <label
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.3rem",
                    cursor: "pointer",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={profile.windows}
                    onChange={(e) =>
                      setProfile({ ...profile, windows: e.target.checked })
                    }
                  />
                  Windows
                </label>
                <label
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.3rem",
                    cursor: "pointer",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={profile.mac}
                    onChange={(e) =>
                      setProfile({ ...profile, mac: e.target.checked })
                    }
                  />
                  macOS
                </label>
                <label
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.3rem",
                    cursor: "pointer",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={profile.linux}
                    onChange={(e) =>
                      setProfile({ ...profile, linux: e.target.checked })
                    }
                  />
                  Linux
                </label>
              </div>
            </div>

            <div>
              <label className="filter-label">
                Localization Languages (Comma-separated)
              </label>
              <input
                className="input"
                style={{ width: "100%" }}
                value={profile.localization_languages.join(", ")}
                onChange={(e) =>
                  setProfile({
                    ...profile,
                    localization_languages: e.target.value
                      .split(",")
                      .map((s) => s.trim())
                      .filter(Boolean),
                  })
                }
              />
            </div>

            <button
              type="submit"
              className="btn btn-primary"
              style={{ marginTop: "0.5rem" }}
            >
              Benchmark Against Market Peers
            </button>
          </form>
        </div>

        {/* Benchmark Results Display */}
        <div>
          {loading && (
            <div
              className="card"
              style={{ padding: "3rem", textAlign: "center" }}
            >
              <div style={{ color: "var(--cyan)", fontWeight: 600 }}>
                Calculating cohort distributions for comparable titles...
              </div>
            </div>
          )}

          {error && (
            <div className="alert alert-warning">
              <AlertCircle size={18} />
              <div>{error}</div>
            </div>
          )}

          {!benchmark && !loading && (
            <div
              className="card"
              style={{
                padding: "3rem",
                textAlign: "center",
                color: "var(--text-dim)",
              }}
            >
              Click <strong>"Benchmark Against Market Peers"</strong> to
              generate comparative pricing, audience, platform, and localization
              distributions.
            </div>
          )}

          {benchmark && !loading && (
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "1.5rem",
              }}
            >
              {/* Peer Group Summary */}
              <div className="card">
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    marginBottom: "0.75rem",
                  }}
                >
                  <h2>Comparable Peer Group: {profile.genre}</h2>
                  <span className="badge badge-cyan">
                    {benchmark.cohort_summary.total_games.toLocaleString()}{" "}
                    games evaluated
                  </span>
                </div>
                <p style={{ fontSize: "0.85rem", color: "var(--text-dim)" }}>
                  Constructed from existing Steam titles matching{" "}
                  <strong>{profile.genre}</strong>{" "}
                  {profile.secondary_genre
                    ? `and ${profile.secondary_genre}`
                    : ""}{" "}
                  across adjacent release vintages.
                </p>
              </div>

              {/* 1. Pricing Position Context */}
              <div className="card">
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.5rem",
                    marginBottom: "0.75rem",
                  }}
                >
                  <DollarSign size={18} color="var(--cyan)" />
                  <h2>Pricing Position</h2>
                </div>

                <div className="metrics-grid" style={{ marginBottom: "1rem" }}>
                  <div className="metric-tile">
                    <div className="metric-label">Your Proposed Price</div>
                    <div
                      className="metric-value"
                      style={{ color: "var(--cyan)" }}
                    >
                      ${benchmark.pricing_context.proposed_price.toFixed(2)}
                    </div>
                  </div>
                  <div className="metric-tile">
                    <div className="metric-label">Peer Group Median Price</div>
                    <div className="metric-value">
                      $
                      {benchmark.pricing_context.cohort_paid_price_stats.p50?.toFixed(
                        2,
                      ) || "0.00"}
                    </div>
                  </div>
                  <div className="metric-tile">
                    <div className="metric-label">Position In Distribution</div>
                    <div
                      style={{
                        fontSize: "0.9rem",
                        fontWeight: 600,
                        color: "#fff",
                        marginTop: "0.4rem",
                      }}
                    >
                      {benchmark.pricing_context.position_description}
                    </div>
                  </div>
                </div>

                <PercentileRangeBar
                  stats={benchmark.pricing_context.cohort_paid_price_stats}
                  title="Peer Group Base Price Percentiles (USD)"
                  unitPrefix="$"
                />
              </div>

              {/* 2. Platform Context */}
              <div className="card">
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: "0.5rem",
                    marginBottom: "0.75rem",
                  }}
                >
                  <Monitor size={18} color="var(--indigo)" />
                  <h2>Platform Support Context</h2>
                </div>
                <div style={{ marginBottom: "1rem", fontSize: "0.85rem" }}>
                  Your proposed platform configuration:{" "}
                  <strong style={{ color: "var(--cyan)" }}>
                    {benchmark.platform_context.proposed_combination}
                  </strong>
                  <span
                    className={`badge ${benchmark.platform_context.is_common_combination ? "badge-emerald" : "badge-amber"}`}
                    style={{ marginLeft: "0.75rem" }}
                  >
                    {benchmark.platform_context.is_common_combination
                      ? "Common combination"
                      : "Differentiating / Selective"}
                  </span>
                </div>
                <div
                  style={{
                    display: "grid",
                    gridTemplateColumns: "repeat(3, 1fr)",
                    gap: "0.75rem",
                  }}
                >
                  <div className="metric-tile" style={{ textAlign: "center" }}>
                    <div className="metric-label">Peer Windows Rate</div>
                    <div
                      className="metric-value"
                      style={{ justifyContent: "center" }}
                    >
                      {benchmark.platform_context.windows_rate}%
                    </div>
                  </div>
                  <div className="metric-tile" style={{ textAlign: "center" }}>
                    <div className="metric-label">Peer macOS Rate</div>
                    <div
                      className="metric-value"
                      style={{ justifyContent: "center" }}
                    >
                      {benchmark.platform_context.mac_rate}%
                    </div>
                  </div>
                  <div className="metric-tile" style={{ textAlign: "center" }}>
                    <div className="metric-label">Peer Linux Rate</div>
                    <div
                      className="metric-value"
                      style={{
                        justifyContent: "center",
                        color: "var(--emerald)",
                      }}
                    >
                      {benchmark.platform_context.linux_rate}%
                    </div>
                  </div>
                </div>
              </div>

              {/* 3. Localization Coverage */}
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
                  <h2>Target Localization Analysis</h2>
                </div>
                <p
                  style={{
                    fontSize: "0.85rem",
                    color: "var(--text-dim)",
                    marginBottom: "1rem",
                  }}
                >
                  You specified{" "}
                  {benchmark.localization_context.proposed_language_count}{" "}
                  target languages. In this peer group, the median game lists{" "}
                  {
                    benchmark.localization_context.cohort_language_count_stats
                      .p50
                  }{" "}
                  supported languages.
                </p>
                <div className="dist-list">
                  {benchmark.localization_context.languages_analysis.map(
                    (l) => (
                      <div key={l.key} className="dist-item">
                        <div className="dist-item-header">
                          <span className="dist-item-name">{l.label}</span>
                          <span className="dist-item-stats">
                            {l.percentage}% of peer cohort
                          </span>
                        </div>
                        <div className="dist-track">
                          <div
                            className="dist-fill"
                            style={{
                              width: `${Math.max(2, l.percentage)}%`,
                              backgroundColor: "var(--amber)",
                            }}
                          />
                        </div>
                      </div>
                    ),
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
