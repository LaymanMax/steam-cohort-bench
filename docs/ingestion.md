# Data Ingestion Pipeline

The data ingestion pipeline transforms raw Steam data into an analytical Parquet table and DuckDB database.

## Ingestion Commands

### 1. Ingest from an existing local Parquet file:

```bash
go run ./cmd/ingest --input ./data/raw/steam-games.parquet
```

### 2. Download directly from Hugging Face and ingest:

```bash
go run ./cmd/ingest --download
```

### 3. Verify input Parquet schema and checksum without modifying processed data:

```bash
go run ./cmd/ingest --input ./data/raw/steam-games.parquet --verify-only
```

### 4. Ingest with a record limit (for development / rapid testing):

```bash
go run ./cmd/ingest --input ./data/raw/steam-games.parquet --limit 5000
```

---

## Lifecycle Stages

```
1. Input Verification ──► 2. SHA-256 Checksum ──► 3. Schema Profiling
                                                          │
6. Atomic Publication ◄── 5. Manifest & Metadata ◄── 4. Canonical Transform
```

1. **Input Verification**: Validates file accessibility, readability, and byte size.
2. **Checksum**: Computes SHA-256 hash of the input Parquet file for immutable provenance.
3. **Schema Profiling**: Executes DuckDB profiling to audit row count, nullability, and AppID distinctness.
4. **Canonical Transform**: Executes vectorized SQL normalization in a temporary staging directory (`os.MkdirTemp`), compressing to ZSTD Parquet.
5. **Manifest & Metadata**: Extracts distinct genres, categories, years, and brackets, and writes `manifest.json`.
6. **Atomic Publication**: Atomically moves staged files into `data/processed/canonical_games.parquet`, `data/processed/analytics.duckdb`, and `data/manifests/manifest.json`.
