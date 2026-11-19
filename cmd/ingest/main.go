package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/ingest"
	"github.com/laymanmax/steam-cohort-bench/internal/logging"
)

func main() {
	cfg := config.Load()

	downloadFlag := flag.Bool("download", false, "Download source dataset from Hugging Face")
	inputFlag := flag.String("input", cfg.RawParquetPath, "Path to local raw steam-games.parquet file")
	outputDirFlag := flag.String("output-dir", cfg.DataDir, "Directory to output processed dataset and manifests")
	verifyOnlyFlag := flag.Bool("verify-only", false, "Only verify input parquet readability, schema, and checksum without transforming")
	forceFlag := flag.Bool("force", false, "Force re-download or overwrite existing processed data")
	limitFlag := flag.Int("limit", 0, "Limit number of records ingested (for testing/development only)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "steam-cohort-bench data ingestion tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/ingest [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	logger := logging.InitLogger(cfg.LogLevel)

	opts := ingest.PipelineOptions{
		InputPath:  *inputFlag,
		Download:   *downloadFlag,
		OutputDir:  *outputDirFlag,
		VerifyOnly: *verifyOnlyFlag,
		Force:      *forceFlag,
		Limit:      *limitFlag,
	}

	pipeline := ingest.NewPipeline(opts, logger)
	if err := pipeline.Run(); err != nil {
		logger.Error("ingestion failed", "error", err)
		os.Exit(1)
	}

	logger.Info("ingestion completed successfully")
}
