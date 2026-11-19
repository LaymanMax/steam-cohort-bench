package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/api/handlers"
	"github.com/laymanmax/steam-cohort-bench/internal/api/middleware"
	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/storage"
)

// NewRouter creates and configures the HTTP router with routes and middlewares.
func NewRouter(cfg *config.Config, store *storage.Storage, logger *slog.Logger) http.Handler {
	h := handlers.NewHandler(store)
	mux := http.NewServeMux()

	// API Routes
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
	mux.HandleFunc("POST /api/v1/cohorts/export", h.ExportCohortCSV)
	mux.HandleFunc("GET /api/v1/export/games", h.ExportGamesCSV)

	// Static file handler with SPA fallback
	staticHandler := spaFileServer(cfg.StaticDir)
	mux.Handle("/", staticHandler)

	// Apply Middlewares: Timeout -> Recovery -> Logger -> CORS
	handler := middleware.Timeout(30 * time.Second)(mux)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.Logger(logger)(handler)
	handler = middleware.CORS(handler)

	return handler
}

func spaFileServer(staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If path starts with /api/, do not serve index.html
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		info, err := os.Stat(path)

		// If file exists and is not a directory, serve it
		if err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}

		// Otherwise serve index.html for SPA client-side routing
		indexPath := filepath.Join(staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}

		// In development or if web/dist is not built yet, return friendly message
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>steam-cohort-bench</title><style>body{font-family:sans-serif;background:#0f172a;color:#f8fafc;padding:3rem;text-align:center;}</style></head><body><h1>steam-cohort-bench API</h1><p>API is active at <code>/api/v1/</code>.</p><p>Frontend assets are not yet compiled in <code>web/dist</code>. Run <code>pnpm run build</code> in <code>web/</code> or run Vite dev server at <code>http://localhost:5173</code>.</p></body></html>`))
	})
}
