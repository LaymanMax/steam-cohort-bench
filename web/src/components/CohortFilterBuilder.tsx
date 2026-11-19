import { Filter, RotateCcw, X } from "lucide-react";
import React from "react";
import {
  CohortFilter,
  DatasetManifest,
  PriceType,
  QualityMode,
} from "../types";

interface Props {
  filter: CohortFilter;
  manifest: DatasetManifest | null;
  onChange: (newFilter: CohortFilter) => void;
  onReset: () => void;
}

export const CohortFilterBuilder: React.FC<Props> = ({
  filter,
  manifest,
  onChange,
  onReset,
}) => {
  const genres = manifest?.available_genres || [
    "Action",
    "Adventure",
    "Casual",
    "Indie",
    "RPG",
    "Simulation",
    "Strategy",
    "Early Access",
  ];
  const categories = manifest?.available_categories || [
    "Single-player",
    "Multi-player",
    "Co-op",
    "Steam Achievements",
    "Full controller support",
  ];
  const brackets = manifest?.available_brackets || [
    "0 - 20K",
    "20K - 50K",
    "50K - 100K",
    "100K - 200K",
    "200K - 500K",
    "500K - 1M",
    "1M - 2M",
  ];

  const handleGenreToggle = (g: string) => {
    const current = filter.genres || [];
    const exists = current.includes(g);
    const updated = exists ? current.filter((x) => x !== g) : [...current, g];
    onChange({ ...filter, genres: updated });
  };

  const handleCategoryToggle = (c: string) => {
    const current = filter.categories || [];
    const exists = current.includes(c);
    const updated = exists ? current.filter((x) => x !== c) : [...current, c];
    onChange({ ...filter, categories: updated });
  };

  const handleBracketToggle = (b: string) => {
    const current = filter.owners_brackets || [];
    const exists = current.includes(b);
    const updated = exists ? current.filter((x) => x !== b) : [...current, b];
    onChange({ ...filter, owners_brackets: updated });
  };

  const hasActiveFilters = Boolean(
    (filter.genres && filter.genres.length > 0) ||
    (filter.categories && filter.categories.length > 0) ||
    filter.release_year_min ||
    filter.release_year_max ||
    filter.price_min ||
    filter.price_max ||
    (filter.price_type && filter.price_type !== "all") ||
    filter.windows !== undefined ||
    filter.mac !== undefined ||
    filter.linux !== undefined ||
    (filter.owners_brackets && filter.owners_brackets.length > 0) ||
    filter.review_count_min ||
    (filter.quality_mode && filter.quality_mode !== "all"),
  );

  return (
    <div className="filter-panel">
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "1rem",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "0.5rem",
            fontWeight: 600,
            fontSize: "0.95rem",
          }}
        >
          <Filter size={16} color="var(--cyan)" />
          <span>Cohort Definition & Query Filters</span>
        </div>
        {hasActiveFilters && (
          <button className="btn btn-secondary btn-sm" onClick={onReset}>
            <RotateCcw size={13} />
            Reset Filters
          </button>
        )}
      </div>

      <div className="filter-grid">
        {/* Genre Multi-select */}
        <div className="filter-group">
          <label className="filter-label">Add Genre Filter</label>
          <select
            className="select"
            value=""
            onChange={(e) => {
              if (e.target.value) handleGenreToggle(e.target.value);
            }}
          >
            <option value="">Select genre to add...</option>
            {genres.map((g) => (
              <option key={g} value={g} disabled={filter.genres?.includes(g)}>
                {g} {filter.genres?.includes(g) ? "✓" : ""}
              </option>
            ))}
          </select>
        </div>

        {/* Category Multi-select */}
        <div className="filter-group">
          <label className="filter-label">Add Category</label>
          <select
            className="select"
            value=""
            onChange={(e) => {
              if (e.target.value) handleCategoryToggle(e.target.value);
            }}
          >
            <option value="">Select category...</option>
            {categories.map((c) => (
              <option
                key={c}
                value={c}
                disabled={filter.categories?.includes(c)}
              >
                {c} {filter.categories?.includes(c) ? "✓" : ""}
              </option>
            ))}
          </select>
        </div>

        {/* Release Year Range */}
        <div className="filter-group">
          <label className="filter-label">Release Year Range</label>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <input
              type="number"
              placeholder="From (e.g. 2020)"
              className="input"
              style={{ width: "50%" }}
              value={filter.release_year_min || ""}
              onChange={(e) =>
                onChange({
                  ...filter,
                  release_year_min: e.target.value
                    ? parseInt(e.target.value, 10)
                    : undefined,
                })
              }
            />
            <input
              type="number"
              placeholder="To (e.g. 2025)"
              className="input"
              style={{ width: "50%" }}
              value={filter.release_year_max || ""}
              onChange={(e) =>
                onChange({
                  ...filter,
                  release_year_max: e.target.value
                    ? parseInt(e.target.value, 10)
                    : undefined,
                })
              }
            />
          </div>
        </div>

        {/* Price Type */}
        <div className="filter-group">
          <label className="filter-label">Pricing Type</label>
          <select
            className="select"
            value={filter.price_type || "all"}
            onChange={(e) =>
              onChange({ ...filter, price_type: e.target.value as PriceType })
            }
          >
            <option value="all">All (Free + Paid)</option>
            <option value="paid">Paid games only</option>
            <option value="free">Free to play only</option>
          </select>
        </div>

        {/* Price Range */}
        <div className="filter-group">
          <label className="filter-label">Price Range ($USD)</label>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <input
              type="number"
              step="0.5"
              placeholder="Min $"
              className="input"
              style={{ width: "50%" }}
              value={filter.price_min ?? ""}
              onChange={(e) =>
                onChange({
                  ...filter,
                  price_min: e.target.value
                    ? parseFloat(e.target.value)
                    : undefined,
                })
              }
            />
            <input
              type="number"
              step="0.5"
              placeholder="Max $"
              className="input"
              style={{ width: "50%" }}
              value={filter.price_max ?? ""}
              onChange={(e) =>
                onChange({
                  ...filter,
                  price_max: e.target.value
                    ? parseFloat(e.target.value)
                    : undefined,
                })
              }
            />
          </div>
        </div>

        {/* Quality Mode */}
        <div className="filter-group">
          <label className="filter-label">Sample Quality Mode</label>
          <select
            className="select"
            value={filter.quality_mode || "all"}
            onChange={(e) =>
              onChange({
                ...filter,
                quality_mode: e.target.value as QualityMode,
              })
            }
          >
            <option value="all">All Records (Preserves All Sparse Data)</option>
            <option value="exclude_playtests">
              Exclude Probable Playtests
            </option>
            <option value="more_complete">
              More Complete Records (Year + Genre/Cat)
            </option>
          </select>
        </div>

        {/* Platform checkboxes */}
        <div className="filter-group">
          <label className="filter-label">Required Platforms</label>
          <div
            style={{
              display: "flex",
              gap: "1rem",
              marginTop: "0.35rem",
              fontSize: "0.82rem",
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
                checked={filter.windows === true}
                onChange={(e) =>
                  onChange({
                    ...filter,
                    windows: e.target.checked ? true : undefined,
                  })
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
                checked={filter.mac === true}
                onChange={(e) =>
                  onChange({
                    ...filter,
                    mac: e.target.checked ? true : undefined,
                  })
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
                checked={filter.linux === true}
                onChange={(e) =>
                  onChange({
                    ...filter,
                    linux: e.target.checked ? true : undefined,
                  })
                }
              />
              Linux
            </label>
          </div>
        </div>

        {/* Ownership Bracket */}
        <div className="filter-group">
          <label className="filter-label">Estimated Owners Bracket</label>
          <select
            className="select"
            value=""
            onChange={(e) => {
              if (e.target.value) handleBracketToggle(e.target.value);
            }}
          >
            <option value="">Select bracket to filter...</option>
            {brackets.map((b) => (
              <option
                key={b}
                value={b}
                disabled={filter.owners_brackets?.includes(b)}
              >
                {b} {filter.owners_brackets?.includes(b) ? "✓" : ""}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Active Filter Chips */}
      {hasActiveFilters && (
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "0.5rem",
            paddingTop: "0.75rem",
            borderTop: "1px solid var(--border-subtle)",
          }}
        >
          {filter.genres?.map((g) => (
            <span key={g} className="chip">
              Genre: {g}
              <X
                size={12}
                className="chip-remove"
                onClick={() => handleGenreToggle(g)}
              />
            </span>
          ))}
          {filter.categories?.map((c) => (
            <span key={c} className="chip">
              Category: {c}
              <X
                size={12}
                className="chip-remove"
                onClick={() => handleCategoryToggle(c)}
              />
            </span>
          ))}
          {filter.owners_brackets?.map((b) => (
            <span key={b} className="chip">
              Owners: {b}
              <X
                size={12}
                className="chip-remove"
                onClick={() => handleBracketToggle(b)}
              />
            </span>
          ))}
          {(filter.release_year_min || filter.release_year_max) && (
            <span className="chip">
              Years: {filter.release_year_min || "Any"} -{" "}
              {filter.release_year_max || "Any"}
              <X
                size={12}
                className="chip-remove"
                onClick={() =>
                  onChange({
                    ...filter,
                    release_year_min: undefined,
                    release_year_max: undefined,
                  })
                }
              />
            </span>
          )}
          {filter.price_type && filter.price_type !== "all" && (
            <span className="chip">
              Price: {filter.price_type}
              <X
                size={12}
                className="chip-remove"
                onClick={() => onChange({ ...filter, price_type: "all" })}
              />
            </span>
          )}
          {(filter.price_min || filter.price_max) && (
            <span className="chip">
              Price: ${filter.price_min || 0} - ${filter.price_max || "Any"}
              <X
                size={12}
                className="chip-remove"
                onClick={() =>
                  onChange({
                    ...filter,
                    price_min: undefined,
                    price_max: undefined,
                  })
                }
              />
            </span>
          )}
          {filter.linux && (
            <span className="chip">
              OS: Linux Required
              <X
                size={12}
                className="chip-remove"
                onClick={() => onChange({ ...filter, linux: undefined })}
              />
            </span>
          )}
          {filter.mac && (
            <span className="chip">
              OS: macOS Required
              <X
                size={12}
                className="chip-remove"
                onClick={() => onChange({ ...filter, mac: undefined })}
              />
            </span>
          )}
          {filter.quality_mode && filter.quality_mode !== "all" && (
            <span className="chip">
              Quality:{" "}
              {filter.quality_mode === "exclude_playtests"
                ? "No Playtests"
                : "Complete Only"}
              <X
                size={12}
                className="chip-remove"
                onClick={() => onChange({ ...filter, quality_mode: "all" })}
              />
            </span>
          )}
        </div>
      )}
    </div>
  );
};
