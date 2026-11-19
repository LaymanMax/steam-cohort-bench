# `steam-cohort-bench`

[![Go Test](https://img.shields.io/badge/Go%20Tests-Passing-brightgreen.svg)](#testing)
[![Vitest](https://img.shields.io/badge/Vitest-Passing-brightgreen.svg)](#testing)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![DuckDB](https://img.shields.io/badge/Analytical%20Engine-DuckDB-yellow.svg)](https://duckdb.org/)
[![Parquet](https://img.shields.io/badge/Data%20Format-Parquet-blue.svg)](https://parquet.apache.org/)

**`steam-cohort-bench`** is an open-source market research and cohort benchmarking application designed specifically for independent game developers.

Instead of reducing market research to misleading averages that are easily skewed by expensive outliers or viral anomalies, `steam-cohort-bench` uses **distribution-based benchmarking** (P25, Median P50, P75, and Interquartile Ranges) to deliver statistical clarity on pricing, playtime, audience sizes, localization, and operating system support.

---

## Key Features

- **Empirical Market Overview**: Instant macro-level visibility into Steam’s commercial landscape across 124,146 verified games.
- **Multi-Dimensional Cohort Explorer**: Construct granular peer groups across genre, category, release vintage, price ranges, platform requirements, and sample quality modes.
- **Side-by-Side Cohort Comparison**: Benchmark up to 4 cohorts concurrently with aligned statistical metrics and sample-size caveats.
- **Hypothetical Game Profiler**: Benchmark prospective game concepts (genre, proposed price, platform configuration, localization targets) against real market peers without predictive snake-oil.
- **Full Game Catalog & Audit**: Searchable, paginated catalog with individual game inspection and data quality flags.
- **Data Quality & Provenance Audit**: First-class tracking of source discrepancies, playtests, zero-ownership states, and field completeness.
- **One-Click CSV & JSON Exports**: Export cohort benchmarks and filtered game catalogs directly for spreadsheet analysis.

---

## Technology Stack

| Layer                 | Technology                                                                  |
| --------------------- | --------------------------------------------------------------------------- |
| **Backend**           | Go 1.24+ standard library `net/http` router, structured `slog` logging      |
| **Analytical Engine** | DuckDB embedded columnar engine via CGO (`github.com/marcboeker/go-duckdb`) |
| **Data Format**       | Canonical Apache Parquet with ZSTD compression                              |
| **Frontend**          | React 19, TypeScript, Vite 8, Lucide React icons                            |
| **Testing**           | Go unit, integration, and E2E workflow tests; Vitest component tests        |
| **Deployment**        | Multi-stage Dockerfile, docker-compose, single-binary production mode       |

---

## Architecture Overview

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

For complete architecture details, see [docs/architecture.md](./docs/architecture.md).

---

## Dataset & Provenance

- **Source Dataset**: [`FronkonGames/steam-games-dataset`](https://huggingface.co/datasets/FronkonGames/steam-games-dataset)
- **License**: MIT License
- **Attribution**: Fronkon Games, Steam, SteamSpy
- **Ground Truth Count**: **124,146** verified games in `data/train-00000-of-00001.parquet` (SHA-256: `775b17f2a5a52b96a8441739cd0dd4939916335b6edeebbfc4a1b238388c5a77`).
- **Discrepancy Note**: The source README mentions 148,909 games, whereas the actual downloaded Parquet file contains 124,146 rows. The application uses the actual Parquet data as the source of truth and discloses this in the Data Quality report.

---

## Getting Started

### Prerequisites

- **Go**: 1.22+ (tested on Go 1.24/1.27) with CGO enabled
- **Node.js**: v20+ (tested on Node v22)
- **pnpm**: v9+ (or corepack)
- **C Compiler**: clang or gcc (required for DuckDB CGO)

---

### Installation & Quick Start

1. **Clone the repository**:

   ```bash
   git clone https://github.com/laymanmax/steam-cohort-bench.git
   cd steam-cohort-bench
   ```

2. **Install dependencies**:

   ```bash
   make install
   ```

3. **Ingest the dataset**:
   If you have the dataset locally:

   ```bash
   go run ./cmd/ingest --input ./data/raw/steam-games.parquet
   ```

   Or download directly from Hugging Face:

   ```bash
   go run ./cmd/ingest --download
   ```

4. **Build the application**:

   ```bash
   make build
   ```

5. **Start the server**:
   ```bash
   ./bin/server
   ```
   Open [http://localhost:8080](http://localhost:8080) in your browser.

---

## Development

Run the Go API backend and Vite dev server with hot module reloading:

- **Terminal 1 (Backend API)**:
  ```bash
  go run ./cmd/server
  ```
- **Terminal 2 (Frontend Dev Server)**:
  ```bash
  cd web && pnpm run dev
  ```
  Vite will run on [http://localhost:5173](http://localhost:5173) and proxy all `/api` requests to port 8080.

---

## Testing

Run the complete test suite:

```bash
# Run all Go unit, integration, and end-to-end tests
make test-go

# Run React component tests with Vitest
make test-web

# Run everything
make test
```

### Test Coverage Highlights

- **Normalization Tests** (`internal/normalize`): Ownership range parsing, release date year extraction, review count derivations, playtest classification.
- **Analytics Tests** (`internal/analytics`): Quantile continuous linear interpolation (P25, P50, P75), uniform and custom histogram binning, price tier classifications.
- **Cohort Engine Tests** (`internal/cohort`): Parameterized SQL WHERE clause generation, cache key generation, TTL and dataset revision invalidation.
- **API Handler Tests** (`internal/api/handlers`): HTTP contract tests for Health, Meta, Overview, Cohorts, Comparisons, Profiles, Catalog, and CSV exports.
- **E2E Workflow Tests** (`cmd/server/e2e_test.go`): End-to-end execution of all user workflows against the DuckDB analytical dataset.
- **Component Tests** (`web/src/components`): Vitest testing of Percentile Range Bar, Histogram Charts, and responsive states.

---

## Docker Deployment

Build and run using Docker:

```bash
# Build the production image
docker build -t steam-cohort-bench:latest .

# Run with local dataset mounted
docker run -p 8080:8080 -v $(pwd)/data:/app/data steam-cohort-bench:latest
```

Or using Docker Compose:

```bash
docker compose up -d
```

---

## Documentation

- [Architecture Guide](./docs/architecture.md)
- [Canonical Data Dictionary](./docs/data-dictionary.md)
- [Statistical Methodology](./docs/methodology.md)
- [Data Ingestion Pipeline](./docs/ingestion.md)
- [REST API Reference](./docs/api.md)
- [OpenAPI Specification](./docs/openapi.yaml)

---

## License

This project is licensed under the [MIT License](./LICENSE).
