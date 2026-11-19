package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/laymanmax/steam-cohort-bench/internal/api/response"
	"github.com/laymanmax/steam-cohort-bench/internal/domain"
	"github.com/laymanmax/steam-cohort-bench/internal/storage"
)

// Handler holds dependencies for API route handlers.
type Handler struct {
	storage *storage.Storage
}

// NewHandler initializes a new API handler.
func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}

// Health checks server and dataset readiness.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	isReady := h.storage.IsReady()
	status := "ready"
	if !isReady {
		status = "dataset_not_loaded"
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"dataset_ready":   isReady,
		"server":          "steam-cohort-bench",
		"version":         "1.0.0",
		"requires_ingest": !isReady,
	})
}

// Meta returns dataset metadata, revisions, available dimensions, and data-quality summary.
func (h *Handler) Meta(w http.ResponseWriter, r *http.Request) {
	manifest := h.storage.GetManifest()
	if manifest == nil || !h.storage.IsReady() {
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"dataset_ready":        false,
			"message":              "No dataset has been ingested yet. Please run ingestion with `go run ./cmd/ingest` or download the dataset.",
			"available_years":      []int{},
			"available_genres":     []string{},
			"available_categories": []string{},
		})
		return
	}

	response.JSON(w, http.StatusOK, manifest)
}

// Overview returns global high-level market statistics and overall distributions.
func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not yet loaded", nil)
		return
	}

	// Global overview is simply an unconstrained cohort
	filter := domain.CohortFilter{
		Name:        "Global Market Overview",
		QualityMode: domain.QualityModeAll,
	}

	res, err := h.storage.ExecuteCohortBenchmark(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to compute overview: %v", err), nil)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// BenchmarkCohort accepts a cohort definition and computes distribution percentiles and breakdowns.
func (h *Handler) BenchmarkCohort(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	var filter domain.CohortFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload", map[string]string{"error": err.Error()})
		return
	}

	res, err := h.storage.ExecuteCohortBenchmark(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("cohort benchmark failed: %v", err), nil)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// CompareCohortsRequest payload with multiple cohorts.
type CompareCohortsRequest struct {
	Cohorts []domain.CohortFilter `json:"cohorts"`
}

// CompareCohorts computes side-by-side benchmark statistics for up to 4 cohorts.
func (h *Handler) CompareCohorts(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	var req CompareCohortsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", map[string]string{"error": err.Error()})
		return
	}

	if len(req.Cohorts) == 0 {
		response.Error(w, http.StatusBadRequest, "must provide at least 1 cohort for comparison", nil)
		return
	}
	if len(req.Cohorts) > 4 {
		response.Error(w, http.StatusBadRequest, "comparison is limited to a maximum of 4 cohorts", nil)
		return
	}

	var results []domain.CohortBenchmarkResult
	var warnings []string

	for i, f := range req.Cohorts {
		if f.ID == "" {
			f.ID = fmt.Sprintf("cohort_%d", i+1)
		}
		if f.Name == "" {
			f.Name = fmt.Sprintf("Cohort %d", i+1)
		}
		res, err := h.storage.ExecuteCohortBenchmark(r.Context(), f)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed benchmarking cohort %s: %v", f.Name, err), nil)
			return
		}
		results = append(results, *res)
	}

	// Check sample overlap / small sample warnings
	for _, res := range results {
		if res.MarketSize.TotalGames < 30 {
			warnings = append(warnings, fmt.Sprintf("%q has a small sample (%d games); statistical comparisons should be interpreted cautiously.", res.CohortName, res.MarketSize.TotalGames))
		}
	}

	response.JSON(w, http.StatusOK, domain.CohortComparisonResult{
		Cohorts:  results,
		Warnings: warnings,
	})
}

// BenchmarkProfile evaluates a hypothetical game specification against comparable market peers.
func (h *Handler) BenchmarkProfile(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	var profile domain.HypotheticalProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON payload", map[string]string{"error": err.Error()})
		return
	}

	if profile.Genre == "" {
		response.Error(w, http.StatusBadRequest, "primary genre is required for hypothetical profile", nil)
		return
	}

	res, err := h.storage.BenchmarkHypotheticalProfile(r.Context(), profile)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to benchmark profile: %v", err), nil)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// GetCatalog handles searchable, filtered, and paginated game queries.
func (h *Handler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	params := storage.CatalogQueryParams{
		Search:      q.Get("search"),
		Bracket:     q.Get("bracket"),
		PriceType:   q.Get("price_type"),
		QualityMode: q.Get("quality_mode"),
		SortBy:      q.Get("sort_by"),
		SortDir:     q.Get("sort_dir"),
		Page:        page,
		PageSize:    pageSize,
	}

	if g := q.Get("genre"); g != "" {
		params.Genres = strings.Split(g, ",")
	}
	if c := q.Get("category"); c != "" {
		params.Categories = strings.Split(c, ",")
	}
	if yMin, err := strconv.Atoi(q.Get("year_min")); err == nil {
		params.YearMin = &yMin
	}
	if yMax, err := strconv.Atoi(q.Get("year_max")); err == nil {
		params.YearMax = &yMax
	}
	if pMin, err := strconv.ParseFloat(q.Get("price_min"), 64); err == nil {
		params.PriceMin = &pMin
	}
	if pMax, err := strconv.ParseFloat(q.Get("price_max"), 64); err == nil {
		params.PriceMax = &pMax
	}
	if rMin, err := strconv.ParseInt(q.Get("review_min"), 10, 64); err == nil {
		params.ReviewMin = &rMin
	}
	if win := q.Get("windows"); win != "" {
		b := win == "true"
		params.Windows = &b
	}
	if mac := q.Get("mac"); mac != "" {
		b := mac == "true"
		params.Mac = &b
	}
	if lin := q.Get("linux"); lin != "" {
		b := lin == "true"
		params.Linux = &b
	}

	res, err := h.storage.QueryGameCatalog(r.Context(), params)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("catalog query failed: %v", err), nil)
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// GetGame retrieves detailed attributes and data quality flags for an individual game by AppID.
func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	appID := r.PathValue("appID")
	if appID == "" {
		response.Error(w, http.StatusBadRequest, "appID path parameter is required", nil)
		return
	}

	game, err := h.storage.GetGameByAppID(r.Context(), appID)
	if err != nil {
		response.Error(w, http.StatusNotFound, fmt.Sprintf("game with AppID %s not found", appID), nil)
		return
	}

	response.JSON(w, http.StatusOK, game)
}

// GetDataQuality provides full audit metrics on missing data, playtests, and row counts.
func (h *Handler) GetDataQuality(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	report, err := h.storage.GetDataQualityReport(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate data quality report: %v", err), nil)
		return
	}

	response.JSON(w, http.StatusOK, report)
}

// ExportCohortCSV exports cohort statistics as a CSV file.
func (h *Handler) ExportCohortCSV(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	var filter domain.CohortFilter
	if r.Method == http.MethodPost {
		_ = json.NewDecoder(r.Body).Decode(&filter)
	} else {
		// Parse query params for GET
		q := r.URL.Query()
		if g := q.Get("genre"); g != "" {
			filter.Genres = strings.Split(g, ",")
		}
		filter.QualityMode = domain.QualityMode(q.Get("quality_mode"))
	}

	data, err := h.storage.ExportCohortSummaryCSV(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("export failed: %v", err), nil)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cohort-summary.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// ExportGamesCSV exports filtered matching games as a CSV file.
func (h *Handler) ExportGamesCSV(w http.ResponseWriter, r *http.Request) {
	if !h.storage.IsReady() {
		response.Error(w, http.StatusServiceUnavailable, "analytical dataset is not loaded", nil)
		return
	}

	q := r.URL.Query()
	var filter domain.CohortFilter
	if g := q.Get("genre"); g != "" {
		filter.Genres = strings.Split(g, ",")
	}
	if b := q.Get("bracket"); b != "" {
		filter.OwnersBrackets = []string{b}
	}
	filter.QualityMode = domain.QualityMode(q.Get("quality_mode"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	data, err := h.storage.ExportGamesCSV(r.Context(), filter, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("export failed: %v", err), nil)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"steam-games-export.csv\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
