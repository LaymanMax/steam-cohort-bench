# `steam-cohort-bench` Architecture

`steam-cohort-bench` is a high-performance market research and cohort benchmarking application for independent game developers. It is built using **Go**, **DuckDB**, **Apache Parquet**, and **React**.

```
┌────────────────────────────────────────────────────────┐
│                   React Frontend (SPA)                 │
│   (Overview, Cohort Explorer, Comparison, Profile)     │
└───────────────────────────▲────────────────────────────┘
                            │ REST / JSON (HTTP)
┌───────────────────────────▼────────────────────────────┐
│                    Go Backend Service                  │
│   - Router (net/http standard library mux)             │
│   - Middleware (CORS, Recovery, Request Timeout, Slog) │
│   - Handlers & Request Validation                      │
│   - In-Memory Cache (LRU + TTL + Revision invalidation)│
│   - Query Engine (Safe Parameterized Filter Generator) │
└───────────────────────────▲────────────────────────────┘
                            │ SQL & Analytical Functions
┌───────────────────────────▼────────────────────────────┐
│                     DuckDB Engine                      │
│   - Vectorized continuous quantile calculations        │
│   - Columnar filtering and list scans                  │
│   - Sub-millisecond aggregation queries                │
└───────────────────────────▲────────────────────────────┘
                            │ Read & Stream
┌───────────────────────────▼────────────────────────────┐
│            Canonical Parquet Storage (ZSTD)            │
│         `data/processed/canonical_games.parquet`       │
└────────────────────────────────────────────────────────┘
```

---

## Key Design Principles

1. **Analytical Engine without a Heavy Database Server**:
   Instead of requiring PostgreSQL or MySQL, the analytical layer uses DuckDB. DuckDB runs embedded inside the Go process via CGO (`github.com/marcboeker/go-duckdb`), enabling vectorized columnar operations over Parquet files in tens of milliseconds.

2. **Immutable Canonical Parquet**:
   During ingestion, the raw source dataset is normalized into `data/processed/canonical_games.parquet`. Raw values are preserved, while normalized typed columns (e.g. parsed release years, lower/upper ownership bounds, total review counts, playtest flags) are materialized.

3. **Atomic Dataset Publication**:
   Ingestion writes to an isolated temporary staging directory (`os.MkdirTemp`). Only after schema verification, row count audits, and metadata extraction pass is the new dataset published atomically to `data/processed/` and `data/manifests/manifest.json`. If ingestion fails, the previously published version remains intact.

4. **Deterministic Revision-Aware Caching**:
   Cohort benchmarks can be expensive when run repeatedly. An in-memory cache stores benchmark results keyed by `SHA256(dataset_revision + ":" + json(filter))`. When the dataset revision changes upon re-ingestion, the cache is automatically invalidated.

5. **Single-Binary Production Deployment**:
   The Go HTTP router serves both the API endpoints under `/api/v1/` and the compiled static assets in `web/dist/`, falling back to `index.html` for single-page client routing.
