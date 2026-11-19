package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/laymanmax/steam-cohort-bench/internal/api"
	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/domain"
	"github.com/laymanmax/steam-cohort-bench/internal/logging"
	"github.com/laymanmax/steam-cohort-bench/internal/storage"
)

func TestEndToEndWorkflows(t *testing.T) {
	cfg := config.Load()
	cwd, _ := os.Getwd()
	if !strings.HasSuffix(cwd, "steam-cohort-bench") {
		cfg.DataDir = "../../data"
		cfg.RawParquetPath = "../../data/raw/steam-games.parquet"
		cfg.ProcessedParquetPath = "../../data/processed/canonical_games.parquet"
		cfg.AnalyticsDBPath = "../../data/processed/analytics.duckdb"
		cfg.ManifestPath = "../../data/manifests/manifest.json"
		cfg.StaticDir = "../../web/dist"
	}
	logger := logging.InitLogger("error")

	store, err := storage.NewStorage(cfg)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}
	if !store.IsReady() {
		t.Fatalf("expected storage to be ready with ingested dataset")
	}

	handler := api.NewRouter(cfg, store, logger)

	// Workflow 1: Open overview & check dataset readiness
	t.Run("Workflow 1: Overview", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/overview", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 overview, got status %v, body %s", rec.Code, rec.Body.String())
		}
		var overview domain.CohortBenchmarkResult
		json.NewDecoder(rec.Body).Decode(&overview)
		if overview.MarketSize.TotalGames != 124146 {
			t.Errorf("expected 124146 games in overview, got %d", overview.MarketSize.TotalGames)
		}
	})

	// Workflow 2: Build Strategy cohort with release year filter
	t.Run("Workflow 2: Strategy Cohort with Vintage", func(t *testing.T) {
		yMin := 2020
		yMax := 2024
		filter := domain.CohortFilter{
			Name:           "Strategy (2020-2024)",
			Genres:         []string{"Strategy"},
			ReleaseYearMin: &yMin,
			ReleaseYearMax: &yMax,
			QualityMode:    domain.QualityModeExcludePlaytests,
		}
		body, _ := json.Marshal(filter)
		req := httptest.NewRequest("POST", "/api/v1/cohorts/benchmark", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 benchmark, got status %v, body %s", rec.Code, rec.Body.String())
		}
		var bench domain.CohortBenchmarkResult
		json.NewDecoder(rec.Body).Decode(&bench)

		if bench.MarketSize.TotalGames == 0 {
			t.Fatalf("expected matching strategy games > 0")
		}

		// Workflow 3: Inspect price and playtime percentiles
		if bench.Pricing.PaidPriceStats.P50 == nil {
			t.Errorf("expected price median p50 to be non-nil")
		}
		if bench.Pricing.PaidPriceStats.P25 == nil || bench.Pricing.PaidPriceStats.P75 == nil {
			t.Errorf("expected price p25 and p75 to be non-nil")
		}
		if *bench.Pricing.PaidPriceStats.P25 > *bench.Pricing.PaidPriceStats.P75 {
			t.Errorf("p25 (%v) should be <= p75 (%v)", *bench.Pricing.PaidPriceStats.P25, *bench.Pricing.PaidPriceStats.P75)
		}
	})

	// Workflow 4: Compare Action, Strategy, Simulation cohorts
	t.Run("Workflow 4: Compare Action, Strategy, Simulation", func(t *testing.T) {
		reqPayload := map[string]interface{}{
			"cohorts": []domain.CohortFilter{
				{Name: "Action", Genres: []string{"Action"}},
				{Name: "Strategy", Genres: []string{"Strategy"}},
				{Name: "Simulation", Genres: []string{"Simulation"}},
			},
		}
		body, _ := json.Marshal(reqPayload)
		req := httptest.NewRequest("POST", "/api/v1/cohorts/compare", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 compare, got %v, body %s", rec.Code, rec.Body.String())
		}
		var comp domain.CohortComparisonResult
		json.NewDecoder(rec.Body).Decode(&comp)
		if len(comp.Cohorts) != 3 {
			t.Errorf("expected 3 cohorts, got %d", len(comp.Cohorts))
		}
	})

	// Workflow 5: Create hypothetical game profile
	t.Run("Workflow 5: Hypothetical Game Profile", func(t *testing.T) {
		prof := domain.HypotheticalProfile{
			Name:                  "Test Indie Tactical",
			Genre:                 "Strategy",
			SecondaryGenre:        "Indie",
			ReleaseYear:           2025,
			ProposedPrice:         19.99,
			Windows:               true,
			Linux:                 true,
			LocalizationLanguages: []string{"English", "German", "Japanese"},
		}
		body, _ := json.Marshal(prof)
		req := httptest.NewRequest("POST", "/api/v1/profiles/benchmark", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 profile benchmark, got %v, body %s", rec.Code, rec.Body.String())
		}
		var pBench domain.ProfileBenchmarkResult
		json.NewDecoder(rec.Body).Decode(&pBench)
		if pBench.Pricing.PositionDescription == "" {
			t.Errorf("expected position description")
		}
		if pBench.Platforms.ProposedCombination == "" {
			t.Errorf("expected platform proposed combination")
		}
	})

	// Workflow 6: Inspect a catalog record
	t.Run("Workflow 6: Catalog Record Inspection", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/games?search=Portal&page=1&page_size=5", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 catalog, got %v", rec.Code)
		}
		var cat domain.GameCatalogResponse
		json.NewDecoder(rec.Body).Decode(&cat)
		if len(cat.Games) == 0 {
			t.Fatalf("expected games found for 'Portal'")
		}
		appID := cat.Games[0].AppID

		// Fetch individual detail
		dReq := httptest.NewRequest("GET", "/api/v1/games/"+appID, nil)
		dRec := httptest.NewRecorder()
		handler.ServeHTTP(dRec, dReq)

		if dRec.Code != http.StatusOK {
			t.Fatalf("expected 200 game detail, got %v", dRec.Code)
		}
		var game domain.GameRecord
		json.NewDecoder(dRec.Body).Decode(&game)
		if game.AppID != appID {
			t.Errorf("expected appID %s, got %s", appID, game.AppID)
		}
	})

	// Workflow 7: Open data-quality report
	t.Run("Workflow 7: Data Quality Report", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/data-quality", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 data quality, got %v", rec.Code)
		}
		var dq domain.DataQualityReport
		json.NewDecoder(rec.Body).Decode(&dq)
		if dq.SourceRowCount != 124146 {
			t.Errorf("expected 124146 rows, got %d", dq.SourceRowCount)
		}
		if len(dq.FieldCompleteness) == 0 {
			t.Errorf("expected field completeness items")
		}
	})

	// Workflow 8: Export cohort results as CSV
	t.Run("Workflow 8: CSV Export", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/export/cohort?genre=Strategy", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 export, got %v", rec.Code)
		}
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/csv") {
			t.Errorf("expected text/csv, got %s", ct)
		}
	})

	// Verify SPA index.html is served for root
	t.Run("SPA Frontend Serving", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for root, got %v", rec.Code)
		}
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("expected text/html, got %s", ct)
		}
	})
}
