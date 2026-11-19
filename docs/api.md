# API Reference

The `steam-cohort-bench` HTTP API is served at `/api/v1/`.

## Endpoints

### 1. `GET /api/v1/health`

Returns service readiness and analytical dataset status.

**Response (200 OK):**

```json
{
  "dataset_ready": true,
  "requires_ingest": false,
  "server": "steam-cohort-bench",
  "status": "ready",
  "version": "1.0.0"
}
```

---

### 2. `GET /api/v1/meta`

Returns dataset provenance, checksum, row counts, available genres, categories, and years.

---

### 3. `GET /api/v1/overview`

Returns global market statistics, price percentiles, ownership bracket distribution, and platform support rates.

---

### 4. `POST /api/v1/cohorts/benchmark`

Calculates comprehensive distributions for a custom multi-dimensional cohort.

**Request Body:**

```json
{
  "name": "Strategy Indie 2020-2025",
  "genres": ["Strategy", "Indie"],
  "release_year_min": 2020,
  "release_year_max": 2025,
  "price_type": "paid",
  "price_min": 4.99,
  "price_max": 29.99,
  "linux": true,
  "quality_mode": "exclude_playtests"
}
```

---

### 5. `POST /api/v1/cohorts/compare`

Calculates side-by-side benchmark statistics for up to 4 cohorts simultaneously.

**Request Body:**

```json
{
  "cohorts": [
    { "name": "Action", "genres": ["Action"] },
    { "name": "Strategy", "genres": ["Strategy"] },
    { "name": "Simulation", "genres": ["Simulation"] }
  ]
}
```

---

### 6. `POST /api/v1/profiles/benchmark`

Benchmarks a hypothetical game specification against comparable market peers.

**Request Body:**

```json
{
  "name": "Project Nova",
  "genre": "Strategy",
  "secondary_genre": "Simulation",
  "release_year": 2026,
  "proposed_price": 19.99,
  "windows": true,
  "linux": true,
  "localization_languages": ["English", "German", "Japanese"]
}
```

---

### 7. `GET /api/v1/games`

Paginated, searchable game catalog with sorting and filtering.

**Query Parameters:**

- `search`: title or AppID substring
- `genre`: comma-separated genres
- `category`: comma-separated categories
- `year_min`, `year_max`: release year bounds
- `price_min`, `price_max`: price bounds
- `price_type`: `all`, `free`, `paid`
- `quality_mode`: `all`, `exclude_playtests`, `more_complete`
- `sort_by`: `total_reviews`, `price`, `release_date`, `median_playtime_forever`, `owners_lower`
- `sort_dir`: `asc`, `desc`
- `page`: 1-based page number
- `page_size`: records per page (default: 25, max: 100)

---

### 8. `GET /api/v1/games/{appID}`

Returns full attributes, descriptions, and audit quality flags for a specific Steam game.

---

### 9. `GET /api/v1/data-quality`

Returns complete audit report on missing values, duplicate IDs, and playtests.

---

### 10. `GET /api/v1/export/cohort`

Exports cohort summary metrics as a CSV file.

---

### 11. `GET /api/v1/export/games`

Streams matching game records as a CSV file.
