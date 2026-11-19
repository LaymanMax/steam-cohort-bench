package ingest

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/dataset"
	"github.com/laymanmax/steam-cohort-bench/internal/domain"
	_ "github.com/marcboeker/go-duckdb"
)

const (
	DefaultHuggingFaceParquetURL = "https://huggingface.co/datasets/FronkonGames/steam-games-dataset/resolve/main/data/train-00000-of-00001.parquet"
	DefaultDatasetRepo           = "https://huggingface.co/datasets/FronkonGames/steam-games-dataset"
	DefaultDatasetName           = "FronkonGames/steam-games-dataset"
	DefaultLicense               = "MIT"
	DefaultAttribution           = "Fronkon Games (https://github.com/FronkonGames), Steam, SteamSpy"
	DefaultPriceUnit             = "USD" // Steam standard store currency in this dataset
	ReadmeReportedRowCount       = 148909
)

// PipelineOptions defines options for the ingestion run.
type PipelineOptions struct {
	InputPath   string
	Download    bool
	DownloadURL string
	OutputDir   string
	VerifyOnly  bool
	Force       bool
	Limit       int
}

// Pipeline coordinates the download, validation, normalization, and publication of the dataset.
type Pipeline struct {
	logger *slog.Logger
	opts   PipelineOptions
}

// NewPipeline creates a new ingestion pipeline.
func NewPipeline(opts PipelineOptions, logger *slog.Logger) *Pipeline {
	if logger == nil {
		logger = slog.Default()
	}
	if opts.DownloadURL == "" {
		opts.DownloadURL = DefaultHuggingFaceParquetURL
	}
	if opts.OutputDir == "" {
		opts.OutputDir = "./data"
	}
	return &Pipeline{
		logger: logger,
		opts:   opts,
	}
}

// Run executes the complete ingestion lifecycle.
func (p *Pipeline) Run() error {
	p.logger.Info("starting ingestion pipeline", "download", p.opts.Download, "input", p.opts.InputPath, "verifyOnly", p.opts.VerifyOnly)

	rawParquetPath := p.opts.InputPath
	rawDir := filepath.Join(p.opts.OutputDir, "raw")
	processedDir := filepath.Join(p.opts.OutputDir, "processed")
	manifestDir := filepath.Join(p.opts.OutputDir, "manifests")

	if err := os.MkdirAll(rawDir, 0755); err != nil {
		return fmt.Errorf("create raw dir: %w", err)
	}
	if err := os.MkdirAll(processedDir, 0755); err != nil {
		return fmt.Errorf("create processed dir: %w", err)
	}
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}

	// 1. Download if requested
	if p.opts.Download {
		rawParquetPath = filepath.Join(rawDir, "steam-games.parquet")
		if _, err := os.Stat(rawParquetPath); err == nil && !p.opts.Force {
			p.logger.Info("raw parquet already exists locally, use --force to re-download", "path", rawParquetPath)
		} else {
			p.logger.Info("downloading source dataset from Hugging Face", "url", p.opts.DownloadURL, "destination", rawParquetPath)
			if err := p.downloadFile(p.opts.DownloadURL, rawParquetPath); err != nil {
				return fmt.Errorf("download failed: %w", err)
			}
			p.logger.Info("download completed successfully", "path", rawParquetPath)
		}
	}

	if rawParquetPath == "" {
		rawParquetPath = filepath.Join(rawDir, "steam-games.parquet")
	}

	// 2. Verify input exists and is readable
	info, err := os.Stat(rawParquetPath)
	if err != nil {
		return fmt.Errorf("source parquet not accessible at %s: %w", rawParquetPath, err)
	}
	p.logger.Info("verified input parquet exists", "path", rawParquetPath, "size_bytes", info.Size())

	// 3. Compute checksum
	checksum, err := p.computeSHA256(rawParquetPath)
	if err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}
	p.logger.Info("computed source file checksum", "sha256", checksum)

	// 4. Open DuckDB in-memory for profiling & transformation
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return fmt.Errorf("open duckdb: %w", err)
	}
	defer db.Close()

	// Verify readable by DuckDB and count rows
	var sourceRowCount int64
	var distinctAppIDs int64
	err = db.QueryRow(fmt.Sprintf("SELECT count(*), count(DISTINCT appID) FROM read_parquet('%s')", rawParquetPath)).
		Scan(&sourceRowCount, &distinctAppIDs)
	if err != nil {
		return fmt.Errorf("read source parquet with duckdb: %w", err)
	}
	duplicateIDs := sourceRowCount - distinctAppIDs

	p.logger.Info("source parquet profiling",
		"total_rows", sourceRowCount,
		"distinct_app_ids", distinctAppIDs,
		"duplicate_ids", duplicateIDs,
		"readme_discrepancy", ReadmeReportedRowCount-sourceRowCount,
	)

	if p.opts.VerifyOnly {
		p.logger.Info("verification complete (verify-only flag specified)")
		return nil
	}

	// 5. Build canonical representation atomically in a staging temp directory
	stagingDir, err := os.MkdirTemp("", "steam-ingest-*")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(stagingDir) // Clean up staging dir upon exit

	stagingParquet := filepath.Join(stagingDir, "canonical_games.parquet")
	stagingDB := filepath.Join(stagingDir, "analytics.duckdb")

	limitClause := ""
	if p.opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", p.opts.Limit)
		p.logger.Warn("limit applied for ingestion test/dev", "limit", p.opts.Limit)
	}

	p.logger.Info("generating canonical normalized parquet and analytical tables", "staging_file", stagingParquet)

	// Transform and write to canonical Parquet with ZSTD compression
	transformSQL := fmt.Sprintf(`
	COPY (
		SELECT 
			CAST(appID AS VARCHAR) as app_id,
			COALESCE(name, '') as name,
			COALESCE(release_date, '') as release_date,
			TRY_CAST(regexp_extract(release_date, '(19\d\d|20\d\d)', 1) AS INTEGER) as release_year,
			COALESCE(price, 0.0) as price,
			(COALESCE(price, 0.0) = 0.0) as is_free,
			COALESCE(estimated_owners, '') as estimated_owners,
			COALESCE(TRY_CAST(replace(regexp_extract(estimated_owners, '([\d,]+)\s*-\s*([\d,]+)', 1), ',', '') AS BIGINT), 0) as owners_lower,
			COALESCE(TRY_CAST(replace(regexp_extract(estimated_owners, '([\d,]+)\s*-\s*([\d,]+)', 2), ',', '') AS BIGINT), 0) as owners_upper,
			CASE 
				WHEN owners_lower = 0 AND owners_upper = 0 THEN '0 - 0'
				WHEN owners_lower = 0 AND owners_upper <= 20000 THEN '0 - 20K'
				WHEN owners_lower >= 20000 AND owners_upper <= 50000 THEN '20K - 50K'
				WHEN owners_lower >= 50000 AND owners_upper <= 100000 THEN '50K - 100K'
				WHEN owners_lower >= 100000 AND owners_upper <= 200000 THEN '100K - 200K'
				WHEN owners_lower >= 200000 AND owners_upper <= 500000 THEN '200K - 500K'
				WHEN owners_lower >= 500000 AND owners_upper <= 1000000 THEN '500K - 1M'
				WHEN owners_lower >= 1000000 AND owners_upper <= 2000000 THEN '1M - 2M'
				WHEN owners_lower >= 2000000 AND owners_upper <= 5000000 THEN '2M - 5M'
				WHEN owners_lower >= 5000000 AND owners_upper <= 10000000 THEN '5M - 10M'
				WHEN owners_lower >= 10000000 AND owners_upper <= 20000000 THEN '10M - 20M'
				WHEN owners_lower >= 20000000 AND owners_upper <= 50000000 THEN '20M - 50M'
				WHEN owners_lower >= 50000000 AND owners_upper <= 100000000 THEN '50M - 100M'
				WHEN owners_lower >= 100000000 THEN '100M+'
				ELSE 'Unknown'
			END as owners_bracket,
			COALESCE(peak_ccu, 0) as peak_ccu,
			COALESCE(required_age, 0) as required_age,
			COALESCE(dlc_count, 0) as dlc_count,
			COALESCE(windows, false) as windows,
			COALESCE(mac, false) as mac,
			COALESCE(linux, false) as linux,
			COALESCE(positive, 0) as positive,
			COALESCE(negative, 0) as negative,
			(COALESCE(positive, 0) + COALESCE(negative, 0)) as total_reviews,
			CASE WHEN (COALESCE(positive, 0) + COALESCE(negative, 0)) > 0 
				THEN ROUND((COALESCE(positive, 0) * 100.0) / (COALESCE(positive, 0) + COALESCE(negative, 0)), 2)
				ELSE NULL 
			END as review_score_pct,
			metacritic_score,
			user_score,
			COALESCE(average_playtime_forever, 0) as average_playtime_forever,
			COALESCE(average_playtime_2weeks, 0) as average_playtime_2weeks,
			COALESCE(median_playtime_forever, 0) as median_playtime_forever,
			ROUND(COALESCE(median_playtime_forever, 0) / 60.0, 2) as median_playtime_hours,
			COALESCE(median_playtime_2weeks, 0) as median_playtime_2weeks,
			COALESCE(genres, []) as genres,
			COALESCE(categories, []) as categories,
			COALESCE(supported_languages, []) as supported_languages,
			COALESCE(full_audio_languages, []) as full_audio_languages,
			len(COALESCE(supported_languages, [])) as supported_lang_count,
			len(COALESCE(full_audio_languages, [])) as audio_lang_count,
			COALESCE(developers, []) as developers,
			COALESCE(publishers, []) as publishers,
			COALESCE(short_description, '') as short_description,
			COALESCE(header_image, '') as header_image,
			COALESCE(website, '') as website,
			COALESCE(support_url, '') as support_url,
			(lower(name) LIKE '%%playtest%%' OR lower(name) LIKE '%%closed beta%%' OR lower(name) LIKE '%%open beta%%' OR lower(name) LIKE '%%alpha test%%') as is_playtest,
			(len(COALESCE(genres, [])) = 0 AND len(COALESCE(categories, [])) = 0 AND (COALESCE(estimated_owners, '') IN ('0 - 0', '')) AND COALESCE(price, 0.0) = 0.0 AND (COALESCE(positive, 0) + COALESCE(negative, 0)) = 0) as is_sparse
		FROM read_parquet('%s')
		%s
	) TO '%s' (FORMAT PARQUET, COMPRESSION ZSTD);
	`, rawParquetPath, limitClause, stagingParquet)

	if _, err := db.Exec(transformSQL); err != nil {
		return fmt.Errorf("canonical parquet generation failed: %w", err)
	}

	// 6. Verify canonical parquet
	var normCount int64
	err = db.QueryRow(fmt.Sprintf("SELECT count(*) FROM read_parquet('%s')", stagingParquet)).Scan(&normCount)
	if err != nil || normCount == 0 {
		return fmt.Errorf("validation of generated canonical parquet failed: %w", err)
	}
	p.logger.Info("canonical parquet validated", "normalized_rows", normCount)

	// 7. Extract distinct metadata for manifest (genres, categories, years, brackets)
	availableGenres, err := p.extractDistinctArrayElements(db, stagingParquet, "genres")
	if err != nil {
		return fmt.Errorf("extract genres: %w", err)
	}
	availableCategories, err := p.extractDistinctArrayElements(db, stagingParquet, "categories")
	if err != nil {
		return fmt.Errorf("extract categories: %w", err)
	}
	availableYears, err := p.extractDistinctYears(db, stagingParquet)
	if err != nil {
		return fmt.Errorf("extract years: %w", err)
	}
	availableBrackets, err := p.extractDistinctBrackets(db, stagingParquet)
	if err != nil {
		return fmt.Errorf("extract brackets: %w", err)
	}

	// 8. Build persistent DuckDB database in staging
	stagingDuckDB, err := sql.Open("duckdb", stagingDB)
	if err != nil {
		return fmt.Errorf("create staging duckdb: %w", err)
	}
	createTableSQL := fmt.Sprintf(`CREATE TABLE games AS SELECT * FROM read_parquet('%s');`, stagingParquet)
	if _, err := stagingDuckDB.Exec(createTableSQL); err != nil {
		stagingDuckDB.Close()
		return fmt.Errorf("populate staging duckdb table: %w", err)
	}
	stagingDuckDB.Close()

	// 9. Generate dataset manifest
	manifest := &domain.DatasetManifest{
		DatasetName:         DefaultDatasetName,
		DatasetRepository:   DefaultDatasetRepo,
		Revision:            checksum[:12],
		SourceFileChecksum:  checksum,
		SourceFilename:      filepath.Base(rawParquetPath),
		SourceFormat:        "parquet",
		SourceRowCount:      sourceRowCount,
		NormalizedRowCount:  normCount,
		DistinctAppIDs:      distinctAppIDs,
		DuplicateIDs:        duplicateIDs,
		IngestedAt:          time.Now().UTC(),
		License:             DefaultLicense,
		Attribution:         DefaultAttribution,
		PriceUnit:           DefaultPriceUnit,
		ReadmeCount:         ReadmeReportedRowCount,
		ViewerCount:         sourceRowCount,
		AvailableGenres:     availableGenres,
		AvailableCategories: availableCategories,
		AvailableYears:      availableYears,
		AvailableBrackets:   availableBrackets,
	}

	stagingManifest := filepath.Join(stagingDir, "manifest.json")
	if err := dataset.SaveManifest(stagingManifest, manifest); err != nil {
		return fmt.Errorf("save staging manifest: %w", err)
	}

	// 10. Atomic Publication
	finalParquet := filepath.Join(processedDir, "canonical_games.parquet")
	finalDB := filepath.Join(processedDir, "analytics.duckdb")
	finalManifest := filepath.Join(manifestDir, "manifest.json")

	if err := copyFile(stagingParquet, finalParquet); err != nil {
		return fmt.Errorf("atomic publish canonical parquet: %w", err)
	}
	if err := copyFile(stagingDB, finalDB); err != nil {
		return fmt.Errorf("atomic publish analytics duckdb: %w", err)
	}
	if err := copyFile(stagingManifest, finalManifest); err != nil {
		return fmt.Errorf("atomic publish manifest: %w", err)
	}

	p.logger.Info("atomic publication complete!",
		"canonical_parquet", finalParquet,
		"analytics_db", finalDB,
		"manifest", finalManifest,
	)

	return nil
}

func (p *Pipeline) extractDistinctArrayElements(db *sql.DB, parquetPath, column string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT unnest(%s) as val 
		FROM read_parquet('%s') 
		WHERE len(%s) > 0 
		ORDER BY val ASC
	`, column, parquetPath, column)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err == nil && s != "" {
			items = append(items, s)
		}
	}
	return items, nil
}

func (p *Pipeline) extractDistinctYears(db *sql.DB, parquetPath string) ([]int, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT release_year 
		FROM read_parquet('%s') 
		WHERE release_year IS NOT NULL 
		ORDER BY release_year DESC
	`, parquetPath)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err == nil {
			years = append(years, y)
		}
	}
	return years, nil
}

func (p *Pipeline) extractDistinctBrackets(db *sql.DB, parquetPath string) ([]string, error) {
	query := fmt.Sprintf(`
		SELECT owners_bracket, min(owners_lower) as ml 
		FROM read_parquet('%s') 
		GROUP BY owners_bracket 
		ORDER BY ml ASC
	`, parquetPath)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brackets []string
	for rows.Next() {
		var b string
		var ml int64
		if err := rows.Scan(&b, &ml); err == nil && b != "" {
			brackets = append(brackets, b)
		}
	}
	return brackets, nil
}

func (p *Pipeline) downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 15 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d downloading from %s", resp.StatusCode, url)
	}

	tempDest := dest + ".download"
	out, err := os.Create(tempDest)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempDest)
		return err
	}

	return os.Rename(tempDest, dest)
}

func (p *Pipeline) computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmpDst := dst + ".tmp"
	out, err := os.Create(tmpDst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmpDst)
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	return os.Rename(tmpDst, dst)
}
