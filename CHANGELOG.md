# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-19

### Added

- **Data Ingestion Pipeline**:
  - Direct download and local import for `FronkonGames/steam-games-dataset`.
  - Schema profiling, SHA-256 checksum verification, and duplicate AppID detection.
  - Normalization for release dates, ownership ranges, player review counts, and platform flags.
  - Probable playtest and sparse record classification.
  - Atomic publication to canonical ZSTD-compressed Parquet and DuckDB analytical storage.
- **Analytical Storage & Engine**:
  - Embedded DuckDB engine with continuous linear interpolation quantiles (`QUANTILE_CONT` P25, Median P50, P75).
  - Multi-dimensional cohort filtering engine supporting genres, categories, vintage, price ranges, platforms, and ownership brackets.
  - Deterministic in-memory LRU cache with automatic dataset revision invalidation.
- **HTTP REST API**:
  - Endpoints for health, metadata, global overview, cohort benchmarking, cohort comparisons (up to 4 cohorts), hypothetical game profiling, catalog search/pagination, data quality auditing, and CSV exports.
  - Standard library `net/http` router with request timeout, CORS, recovery, and structured `slog` logging.
- **React Frontend**:
  - Responsive, dark-first research interface built with React 19, TypeScript, and Vite.
  - Feature pages: Overview, Cohort Explorer, Compare Cohorts, Game Profile, Game Catalog, Data Quality, and Methodology.
  - Interactive visualization components: continuous percentile range bar, histogram charts, comparative distribution lists, and platform support meters.
  - Single-binary distribution serving static SPA assets directly from the Go backend.
- **Testing & Documentation**:
  - Go unit, integration, and end-to-end workflow test suite.
  - Vitest React component test suite.
  - Complete documentation: architecture, data dictionary, methodology, ingestion guide, and OpenAPI specification.
  - Multi-stage Dockerfile and Docker Compose configuration.
