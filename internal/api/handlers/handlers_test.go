package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/laymanmax/steam-cohort-bench/internal/api/middleware"
	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/domain"
	"github.com/laymanmax/steam-cohort-bench/internal/logging"
	"github.com/laymanmax/steam-cohort-bench/internal/storage"
)

func setupTestServer(t *testing.T) http.Handler {
	cfg := config.Load()
	cwd, _ := os.Getwd()
	if !strings.HasSuffix(cwd, "steam-cohort-bench") {
		cfg.DataDir = "../../../data"
		cfg.RawParquetPath = "../../../data/raw/steam-games.parquet"
		cfg.ProcessedParquetPath = "../../../data/processed/canonical_games.parquet"
		cfg.AnalyticsDBPath = "../../../data/processed/analytics.duckdb"
		cfg.ManifestPath = "../../../data/manifests/manifest.json"
	}
	_ = logging.InitLogger("error")
	store, err := storage.NewStorage(cfg)
	if err != nil {
		t.Logf("storage warning: %v", err)
	}

	h := NewHandler(store)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", h.Health)
	mux.HandleFunc("GET /api/v1/meta", h.Meta)
	mux.HandleFunc("GET /api/v1/overview", h.Overview)
	mux.HandleFunc("POST /api/v1/cohorts/benchmark", h.BenchmarkCohort)
	mux.HandleFunc("POST /api/v1/cohorts/compare", h.CompareCohorts)
	mux.HandleFunc("POST /api/v1/profiles/benchmark", h.BenchmarkProfile)
	mux.HandleFunc("GET /api/v1/games", h.GetCatalog)
	mux.HandleFunc("GET /api/v1/games/{appID}", h.GetGame)
	mux.HandleFunc("GET /api/v1/data-quality", h.GetDataQuality)
	mux.HandleFunc("GET /api/v1/export/cohort", h.ExportCohortCSV)
	mux.HandleFunc("GET /api/v1/export/games", h.ExportGamesCSV)

	handler := middleware.CORS(mux)
	return handler
}

func TestHealthEndpoint(t *testing.T) {
	handler := setupTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if res["server"] != "steam-cohort-bench" {
		t.Errorf("expected server name steam-cohort-bench, got %v", res["server"])
	}
}

func TestMetaEndpoint(t *testing.T) {
	handler := setupTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/meta", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var meta domain.DatasetManifest
	if err := json.NewDecoder(rec.Body).Decode(&meta); err != nil {
		t.Fatalf("decode meta: %v", err)
	}

	if meta.SourceRowCount != 124146 {
		t.Errorf("expected 124146 rows, got %d", meta.SourceRowCount)
	}
}

func TestOverviewEndpoint(t *testing.T) {
	handler := setupTestServer(t)
	req := httptest.NewRequest("GET", "/api/v1/overview", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res domain.CohortBenchmarkResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode overview: %v", err)
	}

	if res.MarketSize.TotalGames != 124146 {
		t.Errorf("expected 124146 total games, got %d", res.MarketSize.TotalGames)
	}
	if res.Pricing.PaidPriceStats.P50 == nil {
		t.Errorf("expected paid price median to be present")
	}
}

func TestCohortBenchmarkEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	filter := domain.CohortFilter{
		Name:        "Strategy Cohort",
		Genres:      []string{"Strategy"},
		QualityMode: domain.QualityModeExcludePlaytests,
	}
	body, _ := json.Marshal(filter)

	req := httptest.NewRequest("POST", "/api/v1/cohorts/benchmark", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var res domain.CohortBenchmarkResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode benchmark: %v", err)
	}

	if res.MarketSize.TotalGames == 0 {
		t.Errorf("expected matching strategy games > 0")
	}
	if res.Platforms.WindowsCount == 0 {
		t.Errorf("expected windows games > 0")
	}
}

func TestCohortCompareEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	reqPayload := CompareCohortsRequest{
		Cohorts: []domain.CohortFilter{
			{Name: "Action", Genres: []string{"Action"}},
			{Name: "Strategy", Genres: []string{"Strategy"}},
		},
	}
	body, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest("POST", "/api/v1/cohorts/compare", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var res domain.CohortComparisonResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode comparison: %v", err)
	}

	if len(res.Cohorts) != 2 {
		t.Errorf("expected 2 cohorts compared, got %d", len(res.Cohorts))
	}
}

func TestProfileBenchmarkEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	profile := domain.HypotheticalProfile{
		Name:                  "Project Nova",
		Genre:                 "Strategy",
		SecondaryGenre:        "Simulation",
		ReleaseYear:           2025,
		ProposedPrice:         14.99,
		Windows:               true,
		Linux:                 true,
		LocalizationLanguages: []string{"English", "German", "Japanese"},
	}
	body, _ := json.Marshal(profile)

	req := httptest.NewRequest("POST", "/api/v1/profiles/benchmark", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var res domain.ProfileBenchmarkResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode profile benchmark: %v", err)
	}

	if res.Pricing.ProposedPrice != 14.99 {
		t.Errorf("expected proposed price 14.99, got %v", res.Pricing.ProposedPrice)
	}
	if res.Pricing.PositionDescription == "" {
		t.Errorf("expected non-empty position description")
	}
}

func TestCatalogEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/games?genre=Indie&page=1&page_size=10&sort_by=total_reviews&sort_dir=desc", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var res domain.GameCatalogResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode catalog: %v", err)
	}

	if len(res.Games) != 10 {
		t.Errorf("expected 10 games, got %d", len(res.Games))
	}
	if res.TotalCount == 0 {
		t.Errorf("expected total indie games > 0")
	}
}

func TestGetGameDetails(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/games/496350", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var game domain.GameRecord
	if err := json.NewDecoder(rec.Body).Decode(&game); err != nil {
		t.Fatalf("decode game: %v", err)
	}

	if game.AppID != "496350" {
		t.Errorf("expected appID 496350, got %s", game.AppID)
	}
	if game.Name == "" {
		t.Errorf("expected non-empty game name")
	}
}

func TestDataQualityEndpoint(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/data-quality", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var res domain.DataQualityReport
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode data quality: %v", err)
	}

	if res.SourceRowCount != 124146 {
		t.Errorf("expected 124146 source rows, got %d", res.SourceRowCount)
	}
	if res.ProbablePlaytestCount == 0 {
		t.Errorf("expected playtest count > 0")
	}
}

func TestExportCohortCSV(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/export/cohort?genre=Action", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("expected text/csv, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Total Games") || !strings.Contains(body, "Paid Price Median") {
		t.Errorf("expected CSV to contain key metric labels, got: %s", body)
	}
}

func TestExportGamesCSV(t *testing.T) {
	handler := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/export/games?genre=Strategy&limit=20", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) < 2 {
		t.Errorf("expected header + at least 1 game line, got %d lines", len(lines))
	}
}
