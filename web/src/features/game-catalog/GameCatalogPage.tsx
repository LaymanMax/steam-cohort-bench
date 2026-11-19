import {
  ChevronLeft,
  ChevronRight,
  Download,
  ExternalLink,
  Info,
  Search,
  X,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { api } from "../../services/api";
import { DatasetManifest, GameCatalogResponse, GameRecord } from "../../types";

interface Props {
  manifest: DatasetManifest | null;
}

export const GameCatalogPage: React.FC<Props> = ({ manifest }) => {
  const [gamesData, setGamesData] = useState<GameCatalogResponse | null>(null);
  const [search, setSearch] = useState("");
  const [selectedGenre, setSelectedGenre] = useState("");
  const [selectedQuality, setSelectedQuality] = useState("all");
  const [sortBy, setSortBy] = useState("total_reviews");
  const [sortDir, setSortDir] = useState("desc");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [selectedGame, setSelectedGame] = useState<GameRecord | null>(null);

  useEffect(() => {
    loadGames();
  }, [search, selectedGenre, selectedQuality, sortBy, sortDir, page]);

  const loadGames = async () => {
    setLoading(true);
    try {
      const res = await api.getCatalog({
        search,
        genre: selectedGenre,
        quality_mode: selectedQuality,
        sort_by: sortBy,
        sort_dir: sortDir,
        page,
        page_size: 20,
      });
      setGamesData(res);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  };

  const handleSort = (col: string) => {
    if (sortBy === col) {
      setSortDir(sortDir === "asc" ? "desc" : "asc");
    } else {
      setSortBy(col);
      setSortDir("desc");
    }
    setPage(1);
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
          marginBottom: "1.25rem",
        }}
      >
        <div>
          <h1>Game Catalog & Inspection</h1>
          <p className="lead" style={{ marginBottom: 0 }}>
            Inspect canonical analytical records, data quality tags, and
            benchmark attributes for all{" "}
            {manifest?.source_row_count.toLocaleString() || "124,146"} titles.
          </p>
        </div>

        <a
          href={api.getGamesExportUrl({
            genre: selectedGenre,
            quality_mode: selectedQuality,
            limit: 1000,
          })}
          className="btn btn-secondary"
          download="steam-games-export.csv"
        >
          <Download size={14} />
          Export Matching Games CSV
        </a>
      </div>

      {/* Filter bar */}
      <div
        className="card"
        style={{
          padding: "0.75rem 1.25rem",
          marginBottom: "1.5rem",
          display: "flex",
          flexWrap: "wrap",
          gap: "0.75rem",
          alignItems: "center",
        }}
      >
        <div style={{ position: "relative", flex: "1", minWidth: "220px" }}>
          <Search
            size={15}
            style={{
              position: "absolute",
              left: "10px",
              top: "10px",
              color: "var(--text-muted)",
            }}
          />
          <input
            className="input"
            style={{ width: "100%", paddingLeft: "2rem" }}
            placeholder="Search by title or AppID..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
          />
        </div>

        <select
          className="select"
          value={selectedGenre}
          onChange={(e) => {
            setSelectedGenre(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All Genres</option>
          {genres.map((g) => (
            <option key={g} value={g}>
              {g}
            </option>
          ))}
        </select>

        <select
          className="select"
          value={selectedQuality}
          onChange={(e) => {
            setSelectedQuality(e.target.value);
            setPage(1);
          }}
        >
          <option value="all">Quality: All records</option>
          <option value="exclude_playtests">Exclude playtests</option>
          <option value="more_complete">Complete records only</option>
        </select>
      </div>

      {/* Table */}
      <div className="data-table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th
                onClick={() => handleSort("name")}
                style={{ cursor: "pointer" }}
              >
                Game Title{" "}
                {sortBy === "name" ? (sortDir === "asc" ? "▲" : "▼") : ""}
              </th>
              <th
                onClick={() => handleSort("release_date")}
                style={{ cursor: "pointer" }}
              >
                Release Date{" "}
                {sortBy === "release_date"
                  ? sortDir === "asc"
                    ? "▲"
                    : "▼"
                  : ""}
              </th>
              <th>Genres</th>
              <th
                onClick={() => handleSort("price")}
                style={{ cursor: "pointer" }}
              >
                Price{" "}
                {sortBy === "price" ? (sortDir === "asc" ? "▲" : "▼") : ""}
              </th>
              <th
                onClick={() => handleSort("owners_lower")}
                style={{ cursor: "pointer" }}
              >
                Estimated Owners{" "}
                {sortBy === "owners_lower"
                  ? sortDir === "asc"
                    ? "▲"
                    : "▼"
                  : ""}
              </th>
              <th
                onClick={() => handleSort("total_reviews")}
                style={{ cursor: "pointer" }}
              >
                Reviews{" "}
                {sortBy === "total_reviews"
                  ? sortDir === "asc"
                    ? "▲"
                    : "▼"
                  : ""}
              </th>
              <th
                onClick={() => handleSort("median_playtime_forever")}
                style={{ cursor: "pointer" }}
              >
                Median Playtime{" "}
                {sortBy === "median_playtime_forever"
                  ? sortDir === "asc"
                    ? "▲"
                    : "▼"
                  : ""}
              </th>
              <th>Platforms</th>
              <th>Audit</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td
                  colSpan={9}
                  style={{
                    textAlign: "center",
                    padding: "2rem",
                    color: "var(--text-dim)",
                  }}
                >
                  Loading catalog page...
                </td>
              </tr>
            ) : gamesData?.games?.length === 0 ? (
              <tr>
                <td
                  colSpan={9}
                  style={{
                    textAlign: "center",
                    padding: "2rem",
                    color: "var(--text-muted)",
                  }}
                >
                  No games match the specified filters.
                </td>
              </tr>
            ) : (
              gamesData?.games?.map((g) => (
                <tr
                  key={g.app_id}
                  onClick={() => setSelectedGame(g)}
                  style={{ cursor: "pointer" }}
                >
                  <td>
                    <div style={{ fontWeight: 600, color: "var(--text-main)" }}>
                      {g.name}
                    </div>
                    <div
                      style={{
                        fontSize: "0.7rem",
                        color: "var(--text-muted)",
                        fontFamily: "var(--font-mono)",
                      }}
                    >
                      AppID: {g.app_id}
                    </div>
                  </td>
                  <td style={{ fontSize: "0.8rem", color: "var(--text-dim)" }}>
                    {g.release_date || (
                      <span style={{ color: "var(--text-muted)" }}>
                        Missing
                      </span>
                    )}
                  </td>
                  <td>
                    <div
                      style={{
                        display: "flex",
                        flexWrap: "wrap",
                        gap: "0.25rem",
                        maxWidth: "160px",
                      }}
                    >
                      {g.genres?.slice(0, 2).map((genre) => (
                        <span key={genre} className="badge">
                          {genre}
                        </span>
                      ))}
                      {(g.genres?.length || 0) > 2 && (
                        <span className="badge">+{g.genres.length - 2}</span>
                      )}
                      {(!g.genres || g.genres.length === 0) && (
                        <span
                          style={{ color: "var(--amber)", fontSize: "0.75rem" }}
                        >
                          None
                        </span>
                      )}
                    </div>
                  </td>
                  <td style={{ fontFamily: "var(--font-mono)" }}>
                    {g.is_free ? (
                      <span className="badge badge-emerald">Free</span>
                    ) : (
                      `$${g.price.toFixed(2)}`
                    )}
                  </td>
                  <td>
                    <span className="badge badge-cyan">{g.owners_bracket}</span>
                  </td>
                  <td style={{ fontFamily: "var(--font-mono)" }}>
                    {g.total_reviews > 0 ? (
                      <div>
                        {g.total_reviews.toLocaleString()}{" "}
                        {g.review_score_pct !== null &&
                          g.review_score_pct !== undefined && (
                            <span
                              style={{
                                fontSize: "0.72rem",
                                color: "var(--emerald)",
                              }}
                            >
                              ({g.review_score_pct.toFixed(0)}%)
                            </span>
                          )}
                      </div>
                    ) : (
                      <span style={{ color: "var(--text-muted)" }}>0</span>
                    )}
                  </td>
                  <td style={{ fontFamily: "var(--font-mono)" }}>
                    {g.median_playtime_forever > 0 ? (
                      `${g.median_playtime_hours.toFixed(1)} hrs`
                    ) : (
                      <span style={{ color: "var(--text-muted)" }}>0</span>
                    )}
                  </td>
                  <td>
                    <div
                      style={{
                        display: "flex",
                        gap: "0.2rem",
                        fontSize: "0.7rem",
                        fontFamily: "var(--font-mono)",
                      }}
                    >
                      {g.windows && <span title="Windows">Win</span>}
                      {g.mac && (
                        <span title="macOS" style={{ color: "var(--indigo)" }}>
                          Mac
                        </span>
                      )}
                      {g.linux && (
                        <span title="Linux" style={{ color: "var(--emerald)" }}>
                          Lin
                        </span>
                      )}
                    </div>
                  </td>
                  <td>
                    {g.quality_flags?.length > 0 ? (
                      <span
                        className="badge badge-amber"
                        title={g.quality_flags.join(", ")}
                      >
                        {g.quality_flags.length} flags
                      </span>
                    ) : (
                      <span className="badge badge-emerald">Clean</span>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {gamesData && gamesData.total_pages > 1 && (
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginTop: "1.25rem",
          }}
        >
          <div style={{ fontSize: "0.8rem", color: "var(--text-dim)" }}>
            Showing page {gamesData.page} of {gamesData.total_pages} (
            {gamesData.total_count.toLocaleString()} matching titles)
          </div>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button
              className="btn btn-secondary btn-sm"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft size={14} /> Previous
            </button>
            <button
              className="btn btn-secondary btn-sm"
              disabled={page >= gamesData.total_pages}
              onClick={() => setPage(page + 1)}
            >
              Next <ChevronRight size={14} />
            </button>
          </div>
        </div>
      )}

      {/* Detail Modal */}
      {selectedGame && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            backgroundColor: "rgba(2, 6, 23, 0.75)",
            backdropFilter: "blur(3px)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 100,
            padding: "1rem",
          }}
          onClick={() => setSelectedGame(null)}
        >
          <div
            className="card"
            style={{
              maxWidth: "640px",
              width: "100%",
              maxHeight: "90vh",
              overflowY: "auto",
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "flex-start",
                marginBottom: "1rem",
              }}
            >
              <div>
                <h2>{selectedGame.name}</h2>
                <div
                  style={{
                    fontSize: "0.8rem",
                    color: "var(--text-dim)",
                    fontFamily: "var(--font-mono)",
                  }}
                >
                  AppID: {selectedGame.app_id} • Released:{" "}
                  {selectedGame.release_date || "Unknown"}
                </div>
              </div>
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => setSelectedGame(null)}
              >
                <X size={14} />
              </button>
            </div>

            {selectedGame.quality_flags?.length > 0 && (
              <div
                className="alert alert-warning"
                style={{ marginBottom: "1rem" }}
              >
                <Info size={16} />
                <div>
                  <strong>Data Quality Audit Flags:</strong>
                  <div
                    style={{
                      display: "flex",
                      flexWrap: "wrap",
                      gap: "0.35rem",
                      marginTop: "0.35rem",
                    }}
                  >
                    {selectedGame.quality_flags.map((f) => (
                      <span key={f} className="badge badge-amber">
                        {f}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            )}

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "0.75rem",
                marginBottom: "1rem",
              }}
            >
              <div className="metric-tile">
                <div className="metric-label">Reported Base Price</div>
                <div className="metric-value">
                  {selectedGame.is_free
                    ? "Free"
                    : `$${selectedGame.price.toFixed(2)}`}
                </div>
              </div>
              <div className="metric-tile">
                <div className="metric-label">Estimated Owners Bracket</div>
                <div
                  className="metric-value"
                  style={{ fontSize: "1.25rem", color: "var(--cyan)" }}
                >
                  {selectedGame.owners_bracket}
                </div>
              </div>
            </div>

            <div
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr",
                gap: "0.75rem",
                marginBottom: "1rem",
              }}
            >
              <div className="metric-tile">
                <div className="metric-label">User Reviews</div>
                <div className="metric-value" style={{ fontSize: "1.25rem" }}>
                  {selectedGame.total_reviews.toLocaleString()}
                </div>
                <div className="metric-sub">
                  +{selectedGame.positive} / -{selectedGame.negative}
                </div>
              </div>
              <div className="metric-tile">
                <div className="metric-label">Reported Median Playtime</div>
                <div
                  className="metric-value"
                  style={{ fontSize: "1.25rem", color: "var(--emerald)" }}
                >
                  {selectedGame.median_playtime_hours.toFixed(1)} hrs
                </div>
                <div className="metric-sub">
                  {selectedGame.median_playtime_forever} minutes total
                </div>
              </div>
            </div>

            {selectedGame.short_description && (
              <div
                style={{
                  marginBottom: "1rem",
                  fontSize: "0.85rem",
                  color: "var(--text-dim)",
                }}
              >
                {selectedGame.short_description}
              </div>
            )}

            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginTop: "1rem",
                paddingTop: "1rem",
                borderTop: "1px solid var(--border-subtle)",
              }}
            >
              <div style={{ fontSize: "0.8rem", color: "var(--text-muted)" }}>
                Developers:{" "}
                {selectedGame.developers?.join(", ") || "None listed"}
              </div>
              <a
                href={`https://store.steampowered.com/app/${selectedGame.app_id}`}
                target="_blank"
                rel="noreferrer"
                className="btn btn-secondary btn-sm"
              >
                View on Steam Store <ExternalLink size={12} />
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
