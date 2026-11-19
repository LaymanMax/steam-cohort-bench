package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/api"
	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/logging"
	"github.com/laymanmax/steam-cohort-bench/internal/storage"
)

func main() {
	cfg := config.Load()
	logger := logging.InitLogger(cfg.LogLevel)

	logger.Info("starting steam-cohort-bench server",
		"port", cfg.Port,
		"host", cfg.Host,
		"data_dir", cfg.DataDir,
		"processed_parquet", cfg.ProcessedParquetPath,
	)

	store, err := storage.NewStorage(cfg)
	if err != nil {
		logger.Warn("storage initialization warning (dataset may need ingestion)", "error", err)
	}

	router := api.NewRouter(cfg, store, logger)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening", "address", fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	logger.Info("shutting down HTTP server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server exited cleanly")
}
