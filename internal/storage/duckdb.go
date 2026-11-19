package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/analytics"
	"github.com/laymanmax/steam-cohort-bench/internal/cohort"
	"github.com/laymanmax/steam-cohort-bench/internal/config"
	"github.com/laymanmax/steam-cohort-bench/internal/dataset"
	"github.com/laymanmax/steam-cohort-bench/internal/domain"
	_ "github.com/marcboeker/go-duckdb"
)

// Storage coordinates analytical querying against DuckDB.
type Storage struct {
	mu          sync.RWMutex
	db          *sql.DB
	cfg         *config.Config
	manifest    *domain.DatasetManifest
	cache       *cohort.MemoryCache
	isReady     bool
	parquetPath string
}

// NewStorage initializes analytical storage with DuckDB.
func NewStorage(cfg *config.Config) (*Storage, error) {
	s := &Storage{
		cfg:         cfg,
		cache:       cohort.NewMemoryCache(cfg.CacheTTL, cfg.MaxCacheItems),
		parquetPath: cfg.ProcessedParquetPath,
	}

	if err := s.Reload(); err != nil {
		// Log warning, allow startup even if dataset is not yet ingested (first-run mode)
		s.isReady = false
	}

	return s, nil
}

// Reload checks and connects to the processed analytical database and manifest.
func (s *Storage) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Check if manifest exists
	m, err := dataset.LoadManifest(s.cfg.ManifestPath)
	if err != nil {
		s.isReady = false
		return fmt.Errorf("load manifest: %w", err)
	}
	s.manifest = m

	// 2. Check if canonical parquet exists
	if _, err := os.Stat(s.cfg.ProcessedParquetPath); err != nil {
		s.isReady = false
		return fmt.Errorf("canonical parquet missing at %s: %w", s.cfg.ProcessedParquetPath, err)
	}

	// 3. Open DuckDB connection (in-memory with view to parquet, for clean concurrent access)
	if s.db != nil {
		s.db.Close()
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		s.isReady = false
		return fmt.Errorf("open duckdb: %w", err)
	}

	// Create games view pointing to canonical parquet
	viewSQL := fmt.Sprintf(`CREATE OR REPLACE VIEW games AS SELECT * FROM read_parquet('%s');`, s.cfg.ProcessedParquetPath)
	if _, err := db.Exec(viewSQL); err != nil {
		db.Close()
		s.isReady = false
		return fmt.Errorf("create games view: %w", err)
	}

	s.db = db
	s.isReady = true
	s.cache.SetRevision(m.Revision)
	return nil
}

// IsReady returns whether an analytical dataset has been ingested and loaded.
func (s *Storage) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isReady
}

// GetManifest returns the current dataset manifest.
func (s *Storage) GetManifest() *domain.DatasetManifest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manifest
}

// ExecuteCohortBenchmark runs a full multi-dimensional analytical benchmark for a cohort.
func (s *Storage) ExecuteCohortBenchmark(ctx context.Context, filter domain.CohortFilter) (*domain.CohortBenchmarkResult, error) {
	s.mu.RLock()
	if !s.isReady || s.db == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("analytical dataset not loaded: please run ingestion first")
	}
	manifestRev := ""
	if s.manifest != nil {
		manifestRev = s.manifest.Revision
	}
	s.mu.RUnlock()

	// 1. Check cache
	cacheKey := cohort.GenerateCacheKey(manifestRev, &filter)
	if cached, ok := s.cache.Get(cacheKey); ok {
		return cached, nil
	}

	startTime := time.Now()

	// 2. Build parameterized WHERE clause
	qf := cohort.BuildWhereClause(filter)

	// 3. Run high-level aggregation query
	aggQuery := fmt.Sprintf(`
	SELECT 
		count(*) as total_games,
		count(DISTINCT app_id) as distinct_app_ids,
		count(*) FILTER (WHERE is_free = true) as free_count,
		count(*) FILTER (WHERE is_free = false AND price > 0) as paid_count,
		
		-- Paid Price percentiles
		quantile_cont(price, 0.25) FILTER (WHERE is_free = false AND price > 0) as price_p25,
		quantile_cont(price, 0.50) FILTER (WHERE is_free = false AND price > 0) as price_p50,
		quantile_cont(price, 0.75) FILTER (WHERE is_free = false AND price > 0) as price_p75,
		avg(price) FILTER (WHERE is_free = false AND price > 0) as price_mean,
		min(price) FILTER (WHERE is_free = false AND price > 0) as price_min,
		max(price) FILTER (WHERE is_free = false AND price > 0) as price_max,

		-- Playtime forever (minutes)
		quantile_cont(median_playtime_forever, 0.25) FILTER (WHERE median_playtime_forever > 0) as play_p25,
		quantile_cont(median_playtime_forever, 0.50) FILTER (WHERE median_playtime_forever > 0) as play_p50,
		quantile_cont(median_playtime_forever, 0.75) FILTER (WHERE median_playtime_forever > 0) as play_p75,
		avg(median_playtime_forever) FILTER (WHERE median_playtime_forever > 0) as play_mean,
		count(*) FILTER (WHERE median_playtime_forever > 0) as play_sample_count,

		-- Playtime 2weeks (minutes)
		quantile_cont(median_playtime_2weeks, 0.25) FILTER (WHERE median_playtime_2weeks > 0) as play2w_p25,
		quantile_cont(median_playtime_2weeks, 0.50) FILTER (WHERE median_playtime_2weeks > 0) as play2w_p50,
		quantile_cont(median_playtime_2weeks, 0.75) FILTER (WHERE median_playtime_2weeks > 0) as play2w_p75,
		count(*) FILTER (WHERE median_playtime_2weeks > 0) as play2w_sample_count,

		-- Total reviews
		quantile_cont(total_reviews, 0.25) FILTER (WHERE total_reviews > 0) as rev_p25,
		quantile_cont(total_reviews, 0.50) FILTER (WHERE total_reviews > 0) as rev_p50,
		quantile_cont(total_reviews, 0.75) FILTER (WHERE total_reviews > 0) as rev_p75,
		avg(total_reviews) FILTER (WHERE total_reviews > 0) as rev_mean,
		count(*) FILTER (WHERE total_reviews > 0) as rev_sample_count,

		-- Review score pct
		quantile_cont(review_score_pct, 0.25) FILTER (WHERE review_score_pct IS NOT NULL) as score_p25,
		quantile_cont(review_score_pct, 0.50) FILTER (WHERE review_score_pct IS NOT NULL) as score_p50,
		quantile_cont(review_score_pct, 0.75) FILTER (WHERE review_score_pct IS NOT NULL) as score_p75,
		count(*) FILTER (WHERE review_score_pct IS NOT NULL) as score_sample_count,

		-- Platforms
		count(*) FILTER (WHERE windows = true) as win_count,
		count(*) FILTER (WHERE mac = true) as mac_count,
		count(*) FILTER (WHERE linux = true) as linux_count,
		count(*) as known_platforms_count,

		-- Supported languages count
		quantile_cont(supported_lang_count, 0.25) as lang_p25,
		quantile_cont(supported_lang_count, 0.50) as lang_p50,
		quantile_cont(supported_lang_count, 0.75) as lang_p75,
		avg(supported_lang_count) as lang_mean,

		-- Audio languages count
		quantile_cont(audio_lang_count, 0.25) as audio_p25,
		quantile_cont(audio_lang_count, 0.50) as audio_p50,
		quantile_cont(audio_lang_count, 0.75) as audio_p75,
		avg(audio_lang_count) as audio_mean,

		-- Median playtime for free vs paid
		quantile_cont(median_playtime_forever, 0.50) FILTER (WHERE is_free = true AND median_playtime_forever > 0) as free_play_median,
		quantile_cont(median_playtime_forever, 0.50) FILTER (WHERE is_free = false AND median_playtime_forever > 0) as paid_play_median

	FROM games
	WHERE %s
	`, qf.WhereClause)

	var (
		totalGames, distinctAppIDs, freeCount, paidCount int64
		pP25, pP50, pP75, pMean, pMin, pMax              sql.NullFloat64
		plP25, plP50, plP75, plMean                      sql.NullFloat64
		playSampleCount                                  int64
		pl2wP25, pl2wP50, pl2wP75                        sql.NullFloat64
		play2wSampleCount                                int64
		revP25, revP50, revP75, revMean                  sql.NullFloat64
		revSampleCount                                   int64
		scP25, scP50, scP75                              sql.NullFloat64
		scoreSampleCount                                 int64
		winCount, macCount, linuxCount, knownPlatforms   int64
		lP25, lP50, lP75, lMean                          sql.NullFloat64
		aP25, aP50, aP75, aMean                          sql.NullFloat64
		freePlayMed, paidPlayMed                         sql.NullFloat64
	)

	err := s.db.QueryRowContext(ctx, aggQuery, qf.Args...).Scan(
		&totalGames, &distinctAppIDs, &freeCount, &paidCount,
		&pP25, &pP50, &pP75, &pMean, &pMin, &pMax,
		&plP25, &plP50, &plP75, &plMean, &playSampleCount,
		&pl2wP25, &pl2wP50, &pl2wP75, &play2wSampleCount,
		&revP25, &revP50, &revP75, &revMean, &revSampleCount,
		&scP25, &scP50, &scP75, &scoreSampleCount,
		&winCount, &macCount, &linuxCount, &knownPlatforms,
		&lP25, &lP50, &lP75, &lMean,
		&aP25, &aP50, &aP75, &aMean,
		&freePlayMed, &paidPlayMed,
	)
	if err != nil {
		return nil, fmt.Errorf("execute cohort aggregate: %w", err)
	}

	cohortName := filter.Name
	if cohortName == "" {
		if len(filter.Genres) > 0 {
			cohortName = strings.Join(filter.Genres, " / ")
		} else {
			cohortName = "All Games"
		}
	}

	// 4. Query distributions (release year, genres, categories, ownership brackets, platform combinations, languages, review brackets)
	relYearDist := s.queryDistribution(ctx, fmt.Sprintf(`SELECT CAST(release_year AS VARCHAR) as k, CAST(release_year AS VARCHAR) as l, count(*) as c FROM games WHERE %s AND release_year IS NOT NULL GROUP BY release_year ORDER BY release_year DESC LIMIT 25`, qf.WhereClause), qf.Args, totalGames)
	genreDist := s.queryDistribution(ctx, fmt.Sprintf(`SELECT unnest(genres) as k, unnest(genres) as l, count(*) as c FROM games WHERE %s AND len(genres) > 0 GROUP BY k ORDER BY c DESC LIMIT 15`, qf.WhereClause), qf.Args, totalGames)
	catDist := s.queryDistribution(ctx, fmt.Sprintf(`SELECT unnest(categories) as k, unnest(categories) as l, count(*) as c FROM games WHERE %s AND len(categories) > 0 GROUP BY k ORDER BY c DESC LIMIT 15`, qf.WhereClause), qf.Args, totalGames)
	ownersDist := s.queryDistribution(ctx, fmt.Sprintf(`SELECT owners_bracket as k, owners_bracket as l, count(*) as c FROM games WHERE %s GROUP BY owners_bracket ORDER BY min(owners_lower) ASC`, qf.WhereClause), qf.Args, totalGames)
	platformCombos := s.queryPlatformCombinations(ctx, qf.WhereClause, qf.Args, totalGames)
	topLangs := s.queryDistribution(ctx, fmt.Sprintf(`SELECT unnest(supported_languages) as k, unnest(supported_languages) as l, count(*) as c FROM games WHERE %s AND len(supported_languages) > 0 GROUP BY k ORDER BY c DESC LIMIT 15`, qf.WhereClause), qf.Args, totalGames)
	topAudio := s.queryDistribution(ctx, fmt.Sprintf(`SELECT unnest(full_audio_languages) as k, unnest(full_audio_languages) as l, count(*) as c FROM games WHERE %s AND len(full_audio_languages) > 0 GROUP BY k ORDER BY c DESC LIMIT 15`, qf.WhereClause), qf.Args, totalGames)
	reviewBrackets := s.queryDistribution(ctx, fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN total_reviews = 0 THEN '0 reviews'
				WHEN total_reviews <= 10 THEN '1 - 10'
				WHEN total_reviews <= 50 THEN '11 - 50'
				WHEN total_reviews <= 100 THEN '51 - 100'
				WHEN total_reviews <= 500 THEN '101 - 500'
				WHEN total_reviews <= 1000 THEN '501 - 1,000'
				WHEN total_reviews <= 5000 THEN '1,001 - 5,000'
				ELSE '5,000+'
			END as k,
			k as l,
			count(*) as c
		FROM games
		WHERE %s
		GROUP BY k
		ORDER BY min(total_reviews) ASC
	`, qf.WhereClause), qf.Args, totalGames)

	// Price common ranges & histogram
	priceRanges := s.queryDistribution(ctx, fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN price < 5.0 THEN '$0.01 - $4.99'
				WHEN price < 10.0 THEN '$5.00 - $9.99'
				WHEN price < 15.0 THEN '$10.00 - $14.99'
				WHEN price < 20.0 THEN '$15.00 - $19.99'
				WHEN price < 30.0 THEN '$20.00 - $29.99'
				ELSE '$30.00+'
			END as k,
			k as l,
			count(*) as c
		FROM games
		WHERE %s AND is_free = false AND price > 0
		GROUP BY k
		ORDER BY min(price) ASC
	`, qf.WhereClause), qf.Args, paidCount)

	priceHistogram := s.queryHistogram(ctx, fmt.Sprintf(`SELECT price FROM games WHERE %s AND is_free = false AND price > 0`, qf.WhereClause), qf.Args, []float64{0, 5, 10, 15, 20, 30, 50, 70, 100})
	playtimeHistogram := s.queryHistogram(ctx, fmt.Sprintf(`SELECT median_playtime_forever / 60.0 FROM games WHERE %s AND median_playtime_forever > 0`, qf.WhereClause), qf.Args, []float64{0, 1, 2, 5, 10, 20, 50, 100})
	reviewHistogram := s.queryHistogram(ctx, fmt.Sprintf(`SELECT CAST(total_reviews AS DOUBLE) FROM games WHERE %s AND total_reviews > 0`, qf.WhereClause), qf.Args, []float64{0, 10, 50, 100, 500, 1000, 5000, 20000, 100000})

	// Median ownership bracket calculation
	medianOwnersBracket := "Unknown"
	if len(ownersDist) > 0 {
		var cumulative int64
		target := totalGames / 2
		for _, b := range ownersDist {
			cumulative += b.Count
			if cumulative >= target {
				medianOwnersBracket = b.Label
				break
			}
		}
	}

	// Calculate percentages
	var freePct, paidPct float64
	if totalGames > 0 {
		freePct = math.Round((float64(freeCount)/float64(totalGames))*1000) / 10
		paidPct = math.Round((float64(paidCount)/float64(totalGames))*1000) / 10
	}

	var winPct, macPct, linPct float64
	if knownPlatforms > 0 {
		winPct = math.Round((float64(winCount)/float64(knownPlatforms))*1000) / 10
		macPct = math.Round((float64(macCount)/float64(knownPlatforms))*1000) / 10
		linPct = math.Round((float64(linuxCount)/float64(knownPlatforms))*1000) / 10
	}

	// Playtime in hours
	convertMinutesToHours := func(m sql.NullFloat64) *float64 {
		if !m.Valid {
			return nil
		}
		hrs := math.Round((m.Float64/60.0)*100) / 100
		return &hrs
	}

	// Free vs paid playtime comparison
	freePlayHrs := convertMinutesToHours(freePlayMed)
	paidPlayHrs := convertMinutesToHours(paidPlayMed)
	freeVal := 0.0
	paidVal := 0.0
	if freePlayHrs != nil {
		freeVal = *freePlayHrs
	}
	if paidPlayHrs != nil {
		paidVal = *paidPlayHrs
	}
	freeVsPaidPlaytime := []domain.DistributionItem{
		{Key: "free", Label: "Free Games Median Playtime", Count: int64(freeVal * 60), Percentage: freeVal},
		{Key: "paid", Label: "Paid Games Median Playtime", Count: int64(paidVal * 60), Percentage: paidVal},
	}

	warnings := analytics.CheckCohortWarnings(cohortName, totalGames, paidCount)

	result := &domain.CohortBenchmarkResult{
		CohortID:        filter.ID,
		CohortName:      cohortName,
		Color:           filter.Color,
		Filter:          filter,
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		MarketSize: domain.MarketSizeSection{
			TotalGames:      totalGames,
			DistinctAppIDs:  distinctAppIDs,
			EligibleGames:   totalGames,
			ReleaseYearDist: relYearDist,
			GenreDist:       genreDist,
			CategoryDist:    catDist,
		},
		Pricing: domain.PricingSection{
			FreeGamesCount: freeCount,
			FreeGamesPct:   freePct,
			PaidGamesCount: paidCount,
			PaidGamesPct:   paidPct,
			PaidPriceStats: domain.PercentileMetric{
				Metric:       "paid_price",
				Unit:         "USD",
				Description:  "Reported base price among paid games in USD",
				SampleCount:  paidCount,
				MissingCount: totalGames - paidCount,
				P25:          nullFloatPtr(pP25),
				P50:          nullFloatPtr(pP50),
				P75:          nullFloatPtr(pP75),
				Mean:         nullFloatPtr(pMean),
				Min:          nullFloatPtr(pMin),
				Max:          nullFloatPtr(pMax),
			},
			PriceHistogram: priceHistogram,
			CommonRanges:   priceRanges,
		},
		Audience: domain.AudienceSection{
			OwnershipDist:       ownersDist,
			MedianOwnersBracket: medianOwnersBracket,
			TotalWithOwnership:  totalGames,
		},
		Playtime: domain.PlaytimeSection{
			ForeverStatsHours: domain.PercentileMetric{
				Metric:       "playtime_forever_hours",
				Unit:         "hours",
				Description:  "Reported median playtime forever converted to hours",
				SampleCount:  playSampleCount,
				MissingCount: totalGames - playSampleCount,
				P25:          convertMinutesToHours(plP25),
				P50:          convertMinutesToHours(plP50),
				P75:          convertMinutesToHours(plP75),
				Mean:         convertMinutesToHours(plMean),
			},
			TwoWeeksStatsHours: domain.PercentileMetric{
				Metric:       "playtime_2weeks_hours",
				Unit:         "hours",
				Description:  "Reported median playtime over past 2 weeks converted to hours",
				SampleCount:  play2wSampleCount,
				MissingCount: totalGames - play2wSampleCount,
				P25:          convertMinutesToHours(pl2wP25),
				P50:          convertMinutesToHours(pl2wP50),
				P75:          convertMinutesToHours(pl2wP75),
			},
			PlaytimeHistogram:  playtimeHistogram,
			FreeVsPaidPlaytime: freeVsPaidPlaytime,
		},
		Reviews: domain.ReviewsSection{
			TotalReviewStats: domain.PercentileMetric{
				Metric:       "total_reviews",
				Unit:         "reviews",
				Description:  "Total user reviews (positive + negative)",
				SampleCount:  revSampleCount,
				MissingCount: totalGames - revSampleCount,
				P25:          nullFloatPtr(revP25),
				P50:          nullFloatPtr(revP50),
				P75:          nullFloatPtr(revP75),
				Mean:         nullFloatPtr(revMean),
			},
			PositiveScorePct: domain.PercentileMetric{
				Metric:       "review_score_pct",
				Unit:         "percentage",
				Description:  "Positive review percentage among games with reviews",
				SampleCount:  scoreSampleCount,
				MissingCount: totalGames - scoreSampleCount,
				P25:          nullFloatPtr(scP25),
				P50:          nullFloatPtr(scP50),
				P75:          nullFloatPtr(scP75),
			},
			ReviewHistogram: reviewHistogram,
			ReviewBrackets:  reviewBrackets,
		},
		Platforms: domain.PlatformSection{
			WindowsCount: winCount,
			WindowsPct:   winPct,
			MacCount:     macCount,
			MacPct:       macPct,
			LinuxCount:   linuxCount,
			LinuxPct:     linPct,
			KnownSample:  knownPlatforms,
			Combinations: platformCombos,
		},
		Localization: domain.LocalizationSection{
			ListedLanguageCountStats: domain.PercentileMetric{
				Metric:       "supported_language_count",
				Unit:         "languages",
				Description:  "Number of supported interface/subtitle languages per game",
				SampleCount:  totalGames,
				MissingCount: 0,
				P25:          nullFloatPtr(lP25),
				P50:          nullFloatPtr(lP50),
				P75:          nullFloatPtr(lP75),
				Mean:         nullFloatPtr(lMean),
			},
			FullAudioCountStats: domain.PercentileMetric{
				Metric:       "full_audio_language_count",
				Unit:         "languages",
				Description:  "Number of listed full-audio voiceover languages per game",
				SampleCount:  totalGames,
				MissingCount: 0,
				P25:          nullFloatPtr(aP25),
				P50:          nullFloatPtr(aP50),
				P75:          nullFloatPtr(aP75),
				Mean:         nullFloatPtr(aMean),
			},
			TopSupportedLanguages: topLangs,
			TopAudioLanguages:     topAudio,
		},
		Warnings: warnings,
	}

	// Cache result
	s.cache.Set(cacheKey, result)
	return result, nil
}

// BenchmarkHypotheticalProfile compares a proposed game concept with comparable games.
func (s *Storage) BenchmarkHypotheticalProfile(ctx context.Context, p domain.HypotheticalProfile) (*domain.ProfileBenchmarkResult, error) {
	// Construct comparable cohort
	var genres []string
	if p.Genre != "" {
		genres = append(genres, p.Genre)
	}
	if p.SecondaryGenre != "" && p.SecondaryGenre != p.Genre {
		genres = append(genres, p.SecondaryGenre)
	}

	var yearMin, yearMax *int
	if p.ReleaseYear > 0 {
		yMin := p.ReleaseYear - 3
		yMax := p.ReleaseYear + 1
		yearMin = &yMin
		yearMax = &yMax
	}

	cf := domain.CohortFilter{
		Name:           fmt.Sprintf("Comparable to %s (%s)", p.Name, p.Genre),
		Genres:         genres,
		ReleaseYearMin: yearMin,
		ReleaseYearMax: yearMax,
		QualityMode:    domain.QualityModeExcludePlaytests,
	}

	cohortRes, err := s.ExecuteCohortBenchmark(ctx, cf)
	if err != nil {
		return nil, err
	}

	// Analyze pricing position
	paidStats := cohortRes.Pricing.PaidPriceStats
	positionDesc := "Within cohort observed range"
	percentileRank := 50.0

	if paidStats.P25 != nil && paidStats.P75 != nil && paidStats.P50 != nil {
		p25 := *paidStats.P25
		p50 := *paidStats.P50
		p75 := *paidStats.P75

		if p.ProposedPrice < p25 {
			positionDesc = "Below peer-group 25th percentile (budget tier)"
			percentileRank = 15.0
		} else if p.ProposedPrice > p75 {
			positionDesc = "Above peer-group 75th percentile (premium indie tier)"
			percentileRank = 85.0
		} else if p.ProposedPrice <= p50 {
			positionDesc = "Within peer-group lower interquartile range (p25 - median)"
			percentileRank = 37.5
		} else {
			positionDesc = "Within peer-group upper interquartile range (median - p75)"
			percentileRank = 62.5
		}
	}

	pricingContext := domain.ProfilePricingContext{
		ProposedPrice:       p.ProposedPrice,
		CohortStats:         paidStats,
		PositionDescription: positionDesc,
		PercentileRankEst:   percentileRank,
	}

	// Platform combination analysis
	var proposedParts []string
	if p.Windows {
		proposedParts = append(proposedParts, "Win")
	}
	if p.Mac {
		proposedParts = append(proposedParts, "Mac")
	}
	if p.Linux {
		proposedParts = append(proposedParts, "Linux")
	}
	proposedComboStr := strings.Join(proposedParts, " + ")
	if proposedComboStr == "" {
		proposedComboStr = "None specified"
	}

	isCommon := false
	for _, combo := range cohortRes.Platforms.Combinations {
		if strings.Contains(combo.Platforms, "All 3") && len(proposedParts) == 3 {
			isCommon = combo.Percentage >= 10.0
		} else if strings.Contains(combo.Platforms, "Windows only") && len(proposedParts) == 1 && p.Windows {
			isCommon = combo.Percentage >= 20.0
		}
	}

	platformContext := domain.ProfilePlatformContext{
		ProposedCombination: proposedComboStr,
		CohortCombinations:  cohortRes.Platforms.Combinations,
		IsCommonCombination: isCommon,
		WindowsRate:         cohortRes.Platforms.WindowsPct,
		MacRate:             cohortRes.Platforms.MacPct,
		LinuxRate:           cohortRes.Platforms.LinuxPct,
	}

	// Localization context
	var langAnalysis []domain.DistributionItem
	for _, lang := range p.LocalizationLanguages {
		var coveragePct float64
		for _, top := range cohortRes.Localization.TopSupportedLanguages {
			if strings.EqualFold(top.Key, lang) {
				coveragePct = top.Percentage
				break
			}
		}
		langAnalysis = append(langAnalysis, domain.DistributionItem{
			Key:        lang,
			Label:      lang,
			Percentage: coveragePct,
		})
	}

	locContext := domain.ProfileLocalizationContext{
		LanguageCount:        len(p.LocalizationLanguages),
		CohortLangCountStats: cohortRes.Localization.ListedLanguageCountStats,
		LanguagesAnalysis:    langAnalysis,
	}

	return &domain.ProfileBenchmarkResult{
		Profile:       p,
		CohortSummary: cohortRes.MarketSize,
		Pricing:       pricingContext,
		Audience:      cohortRes.Audience,
		Playtime:      cohortRes.Playtime,
		Platforms:     platformContext,
		Localization:  locContext,
		Warnings:      cohortRes.Warnings,
	}, nil
}

// CatalogQueryParams defines parameters for searching the game catalog.
type CatalogQueryParams struct {
	Search      string
	Genres      []string
	Categories  []string
	YearMin     *int
	YearMax     *int
	PriceMin    *float64
	PriceMax    *float64
	PriceType   string // "all", "free", "paid"
	Windows     *bool
	Mac         *bool
	Linux       *bool
	Bracket     string
	ReviewMin   *int64
	QualityMode string
	SortBy      string
	SortDir     string
	Page        int
	PageSize    int
}

// QueryGameCatalog retrieves a paginated list of games matching catalog filters.
func (s *Storage) QueryGameCatalog(ctx context.Context, p CatalogQueryParams) (*domain.GameCatalogResponse, error) {
	s.mu.RLock()
	if !s.isReady || s.db == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("analytical dataset not loaded")
	}
	s.mu.RUnlock()

	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 25
	}

	// Allowed sort columns
	allowedSorts := map[string]string{
		"name":                    "name",
		"release_date":            "release_date",
		"release_year":            "release_year",
		"price":                   "price",
		"total_reviews":           "total_reviews",
		"review_score_pct":        "review_score_pct",
		"median_playtime_forever": "median_playtime_forever",
		"owners_lower":            "owners_lower",
		"peak_ccu":                "peak_ccu",
	}

	sortCol := "total_reviews"
	if col, ok := allowedSorts[strings.ToLower(p.SortBy)]; ok {
		sortCol = col
	}

	sortDir := "DESC"
	if strings.ToUpper(p.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	// Build WHERE clause
	cf := domain.CohortFilter{
		Genres:         p.Genres,
		Categories:     p.Categories,
		ReleaseYearMin: p.YearMin,
		ReleaseYearMax: p.YearMax,
		PriceMin:       p.PriceMin,
		PriceMax:       p.PriceMax,
		PriceType:      domain.PriceType(p.PriceType),
		Windows:        p.Windows,
		Mac:            p.Mac,
		Linux:          p.Linux,
		ReviewCountMin: p.ReviewMin,
		QualityMode:    domain.QualityMode(p.QualityMode),
	}
	if p.Bracket != "" {
		cf.OwnersBrackets = []string{p.Bracket}
	}

	qf := cohort.BuildWhereClause(cf)
	where := qf.WhereClause
	args := qf.Args

	if p.Search != "" {
		where += " AND (lower(name) LIKE ? OR app_id = ?)"
		searchPattern := "%" + strings.ToLower(strings.TrimSpace(p.Search)) + "%"
		args = append(args, searchPattern, strings.TrimSpace(p.Search))
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT count(*) FROM games WHERE %s", where)
	var totalCount int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("count catalog games: %w", err)
	}

	offset := (p.Page - 1) * p.PageSize
	query := fmt.Sprintf(`
	SELECT 
		app_id, name, release_date, release_year, price, is_free,
		estimated_owners, owners_lower, owners_upper, owners_bracket,
		peak_ccu, required_age, dlc_count, windows, mac, linux,
		positive, negative, total_reviews, review_score_pct, metacritic_score,
		average_playtime_forever, median_playtime_forever, median_playtime_hours,
		genres, categories, supported_languages, full_audio_languages,
		developers, publishers, header_image, short_description, website,
		is_playtest, is_sparse
	FROM games
	WHERE %s
	ORDER BY %s %s NULLS LAST
	LIMIT %d OFFSET %d
	`, where, sortCol, sortDir, p.PageSize, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query catalog: %w", err)
	}
	defer rows.Close()

	var games []domain.GameRecord
	for rows.Next() {
		var g domain.GameRecord
		var (
			relYear                sql.NullInt64
			scorePct               sql.NullFloat64
			metaScore              sql.NullInt64
			genresRaw, catsRaw     interface{}
			suppLangsRaw, audioRaw interface{}
			devsRaw, pubsRaw       interface{}
		)

		err := rows.Scan(
			&g.AppID, &g.Name, &g.ReleaseDate, &relYear, &g.Price, &g.IsFree,
			&g.EstimatedOwners, &g.OwnersLower, &g.OwnersUpper, &g.OwnersBracket,
			&g.PeakCCU, &g.RequiredAge, &g.DLCCount, &g.Windows, &g.Mac, &g.Linux,
			&g.Positive, &g.Negative, &g.TotalReviews, &scorePct, &metaScore,
			&g.AveragePlaytimeForever, &g.MedianPlaytimeForever, &g.MedianPlaytimeHours,
			&genresRaw, &catsRaw, &suppLangsRaw, &audioRaw,
			&devsRaw, &pubsRaw, &g.HeaderImage, &g.ShortDescription, &g.Website,
			&g.IsPlaytest, &g.IsSparse,
		)
		if err != nil {
			return nil, fmt.Errorf("scan game: %w", err)
		}

		if relYear.Valid {
			y := int(relYear.Int64)
			g.ReleaseYear = &y
		}
		if scorePct.Valid {
			g.ReviewScorePct = &scorePct.Float64
		}
		if metaScore.Valid {
			g.MetacriticScore = &metaScore.Int64
		}
		g.Genres = toStringSlice(genresRaw)
		g.Categories = toStringSlice(catsRaw)
		g.SupportedLanguages = toStringSlice(suppLangsRaw)
		g.FullAudioLanguages = toStringSlice(audioRaw)
		g.Developers = toStringSlice(devsRaw)
		g.Publishers = toStringSlice(pubsRaw)

		g.QualityFlags = analyticsQualityFlags(g)
		games = append(games, g)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(p.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.GameCatalogResponse{
		TotalCount: totalCount,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
		Games:      games,
	}, nil
}

// GetGameByAppID fetches full details for a single game.
func (s *Storage) GetGameByAppID(ctx context.Context, appID string) (*domain.GameRecord, error) {
	s.mu.RLock()
	if !s.isReady || s.db == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("analytical dataset not loaded")
	}
	s.mu.RUnlock()

	query := `
	SELECT 
		app_id, name, release_date, release_year, price, is_free,
		estimated_owners, owners_lower, owners_upper, owners_bracket,
		peak_ccu, required_age, dlc_count, windows, mac, linux,
		positive, negative, total_reviews, review_score_pct, metacritic_score,
		average_playtime_forever, median_playtime_forever, median_playtime_hours,
		genres, categories, supported_languages, full_audio_languages,
		developers, publishers, header_image, short_description, website,
		is_playtest, is_sparse
	FROM games
	WHERE app_id = ?
	LIMIT 1
	`
	var g domain.GameRecord
	var (
		relYear                sql.NullInt64
		scorePct               sql.NullFloat64
		metaScore              sql.NullInt64
		genresRaw, catsRaw     interface{}
		suppLangsRaw, audioRaw interface{}
		devsRaw, pubsRaw       interface{}
	)

	err := s.db.QueryRowContext(ctx, query, appID).Scan(
		&g.AppID, &g.Name, &g.ReleaseDate, &relYear, &g.Price, &g.IsFree,
		&g.EstimatedOwners, &g.OwnersLower, &g.OwnersUpper, &g.OwnersBracket,
		&g.PeakCCU, &g.RequiredAge, &g.DLCCount, &g.Windows, &g.Mac, &g.Linux,
		&g.Positive, &g.Negative, &g.TotalReviews, &scorePct, &metaScore,
		&g.AveragePlaytimeForever, &g.MedianPlaytimeForever, &g.MedianPlaytimeHours,
		&genresRaw, &catsRaw, &suppLangsRaw, &audioRaw,
		&devsRaw, &pubsRaw, &g.HeaderImage, &g.ShortDescription, &g.Website,
		&g.IsPlaytest, &g.IsSparse,
	)
	if err != nil {
		return nil, err
	}

	if relYear.Valid {
		y := int(relYear.Int64)
		g.ReleaseYear = &y
	}
	if scorePct.Valid {
		g.ReviewScorePct = &scorePct.Float64
	}
	if metaScore.Valid {
		g.MetacriticScore = &metaScore.Int64
	}
	g.Genres = toStringSlice(genresRaw)
	g.Categories = toStringSlice(catsRaw)
	g.SupportedLanguages = toStringSlice(suppLangsRaw)
	g.FullAudioLanguages = toStringSlice(audioRaw)
	g.Developers = toStringSlice(devsRaw)
	g.Publishers = toStringSlice(pubsRaw)
	g.QualityFlags = analyticsQualityFlags(g)

	return &g, nil
}

// GetDataQualityReport compiles dataset-wide completeness, playtest counts, and missing-field rates.
func (s *Storage) GetDataQualityReport(ctx context.Context) (*domain.DataQualityReport, error) {
	s.mu.RLock()
	if !s.isReady || s.db == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("dataset not loaded")
	}
	manifest := s.manifest
	s.mu.RUnlock()

	query := `
	SELECT 
		count(*) as total,
		count(DISTINCT app_id) as distinct_ids,
		count(*) FILTER (WHERE is_playtest = true) as playtests,
		count(*) FILTER (WHERE is_sparse = true) as sparse,
		count(*) FILTER (WHERE release_year IS NULL) as missing_dates,
		count(*) FILTER (WHERE len(genres) = 0) as missing_genres,
		count(*) FILTER (WHERE len(categories) = 0) as missing_categories,
		count(*) FILTER (WHERE price IS NULL) as missing_price,
		count(*) FILTER (WHERE owners_lower = 0 AND owners_upper = 0) as zero_owners,
		count(*) FILTER (WHERE total_reviews = 0) as missing_reviews,
		count(*) FILTER (WHERE median_playtime_forever = 0) as missing_playtime
	FROM games;
	`
	var (
		total, distinctIDs, playtests, sparse, missingDates, missingGenres, missingCategories int64
		missingPrice, zeroOwners, missingReviews, missingPlaytime                             int64
	)
	err := s.db.QueryRowContext(ctx, query).Scan(
		&total, &distinctIDs, &playtests, &sparse,
		&missingDates, &missingGenres, &missingCategories,
		&missingPrice, &zeroOwners, &missingReviews, &missingPlaytime,
	)
	if err != nil {
		return nil, fmt.Errorf("query quality metrics: %w", err)
	}

	calcPct := func(count int64) float64 {
		if total == 0 {
			return 100.0
		}
		valid := total - count
		return math.Round((float64(valid)/float64(total))*1000) / 10
	}

	fieldCompleteness := []domain.DistributionItem{
		{Key: "app_id", Label: "Application ID", Count: distinctIDs, Percentage: 100.0},
		{Key: "name", Label: "Game Title", Count: total, Percentage: 100.0},
		{Key: "release_year", Label: "Release Year", Count: total - missingDates, Percentage: calcPct(missingDates)},
		{Key: "genres", Label: "Genre Metadata", Count: total - missingGenres, Percentage: calcPct(missingGenres)},
		{Key: "categories", Label: "Category Metadata", Count: total - missingCategories, Percentage: calcPct(missingCategories)},
		{Key: "price", Label: "Price Information", Count: total - missingPrice, Percentage: calcPct(missingPrice)},
		{Key: "owners", Label: "Non-Zero Ownership Estimates", Count: total - zeroOwners, Percentage: calcPct(zeroOwners)},
		{Key: "reviews", Label: "Non-Zero Player Reviews", Count: total - missingReviews, Percentage: calcPct(missingReviews)},
		{Key: "playtime", Label: "Non-Zero Median Playtime", Count: total - missingPlaytime, Percentage: calcPct(missingPlaytime)},
	}

	var warnings []string
	if playtests > 0 {
		warnings = append(warnings, fmt.Sprintf("%d titles (%.1f%%) match playtest keywords and can be excluded using quality modes.", playtests, (float64(playtests)/float64(total))*100))
	}
	if zeroOwners > 0 {
		warnings = append(warnings, fmt.Sprintf("%d titles report '0 - 0' estimated owners; these are preserved as distinct states rather than dropped as nulls.", zeroOwners))
	}
	if sparse > 0 {
		warnings = append(warnings, fmt.Sprintf("%d records are sparse (empty genres, categories, and zero reviews).", sparse))
	}

	readmeCount := int64(148909)
	if manifest != nil && manifest.ReadmeCount > 0 {
		readmeCount = manifest.ReadmeCount
	}

	return &domain.DataQualityReport{
		SourceRowCount:         total,
		DistinctAppIDs:         distinctIDs,
		DuplicateIDs:           total - distinctIDs,
		ReadmeReportedCount:    readmeCount,
		DiscrepancyExplanation: fmt.Sprintf("The Hugging Face README cites %d games, while the official Parquet export contains %d rows (%d row difference). The imported Parquet rows are used as the ground truth.", readmeCount, total, readmeCount-total),
		ProbablePlaytestCount:  playtests,
		SparseRecordsCount:     sparse,
		MissingReleaseDates:    missingDates,
		MissingGenres:          missingGenres,
		MissingCategories:      missingCategories,
		MissingPrice:           missingPrice,
		ZeroOwnersCount:        zeroOwners,
		MissingReviews:         missingReviews,
		MissingPlaytime:        missingPlaytime,
		FieldCompleteness:      fieldCompleteness,
		IngestionWarnings:      warnings,
	}, nil
}

// ExportCohortSummaryCSV creates a CSV summary of a cohort benchmark.
func (s *Storage) ExportCohortSummaryCSV(ctx context.Context, filter domain.CohortFilter) ([]byte, error) {
	bench, err := s.ExecuteCohortBenchmark(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// Write header
	w.Write([]string{"Metric", "Value", "Unit", "Sample Count", "Missing Count"})
	writeFloat := func(label string, val *float64, unit string, sample, missing int64) {
		vStr := "N/A"
		if val != nil {
			vStr = fmt.Sprintf("%.2f", *val)
		}
		w.Write([]string{label, vStr, unit, strconv.FormatInt(sample, 10), strconv.FormatInt(missing, 10)})
	}

	w.Write([]string{"Total Games", strconv.FormatInt(bench.MarketSize.TotalGames, 10), "games", strconv.FormatInt(bench.MarketSize.TotalGames, 10), "0"})
	w.Write([]string{"Free Games", strconv.FormatInt(bench.Pricing.FreeGamesCount, 10), "games", strconv.FormatInt(bench.MarketSize.TotalGames, 10), "0"})
	w.Write([]string{"Paid Games", strconv.FormatInt(bench.Pricing.PaidGamesCount, 10), "games", strconv.FormatInt(bench.MarketSize.TotalGames, 10), "0"})
	writeFloat("Paid Price P25", bench.Pricing.PaidPriceStats.P25, "USD", bench.Pricing.PaidGamesCount, bench.Pricing.PaidPriceStats.MissingCount)
	writeFloat("Paid Price Median (P50)", bench.Pricing.PaidPriceStats.P50, "USD", bench.Pricing.PaidGamesCount, bench.Pricing.PaidPriceStats.MissingCount)
	writeFloat("Paid Price P75", bench.Pricing.PaidPriceStats.P75, "USD", bench.Pricing.PaidGamesCount, bench.Pricing.PaidPriceStats.MissingCount)
	writeFloat("Median Playtime Forever P50", bench.Playtime.ForeverStatsHours.P50, "hours", bench.Playtime.ForeverStatsHours.SampleCount, bench.Playtime.ForeverStatsHours.MissingCount)
	w.Write([]string{"Median Ownership Bracket", bench.Audience.MedianOwnersBracket, "bracket", strconv.FormatInt(bench.MarketSize.TotalGames, 10), "0"})
	w.Write([]string{"Windows Support Rate", fmt.Sprintf("%.1f%%", bench.Platforms.WindowsPct), "percentage", strconv.FormatInt(bench.Platforms.KnownSample, 10), "0"})
	w.Write([]string{"Linux Support Rate", fmt.Sprintf("%.1f%%", bench.Platforms.LinuxPct), "percentage", strconv.FormatInt(bench.Platforms.KnownSample, 10), "0"})

	w.Flush()
	return buf.Bytes(), nil
}

// ExportGamesCSV streams matching games as CSV.
func (s *Storage) ExportGamesCSV(ctx context.Context, filter domain.CohortFilter, limit int) ([]byte, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}

	qf := cohort.BuildWhereClause(filter)
	query := fmt.Sprintf(`
	SELECT 
		app_id, name, release_date, release_year, price, is_free,
		estimated_owners, owners_bracket, total_reviews, review_score_pct,
		median_playtime_hours, windows, mac, linux,
		array_to_string(genres, '; ') as genres_str,
		array_to_string(categories, '; ') as cats_str
	FROM games
	WHERE %s
	ORDER BY total_reviews DESC
	LIMIT %d
	`, qf.WhereClause, limit)

	rows, err := s.db.QueryContext(ctx, query, qf.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	w.Write([]string{
		"AppID", "Name", "Release Date", "Release Year", "Price", "Is Free",
		"Estimated Owners", "Owners Bracket", "Total Reviews", "Positive Review %",
		"Median Playtime (Hours)", "Windows", "macOS", "Linux", "Genres", "Categories",
	})

	for rows.Next() {
		var (
			appID, name, relDate, owners, bracket, genresStr, catsStr string
			relYear                                                   sql.NullInt64
			price, playHours                                          float64
			isFree, win, mac, lin                                     bool
			reviews                                                   int64
			scorePct                                                  sql.NullFloat64
		)
		if err := rows.Scan(
			&appID, &name, &relDate, &relYear, &price, &isFree,
			&owners, &bracket, &reviews, &scorePct,
			&playHours, &win, &mac, &lin, &genresStr, &catsStr,
		); err != nil {
			return nil, err
		}

		scoreStr := ""
		if scorePct.Valid {
			scoreStr = fmt.Sprintf("%.1f", scorePct.Float64)
		}
		yearStr := ""
		if relYear.Valid {
			yearStr = strconv.Itoa(int(relYear.Int64))
		}

		w.Write([]string{
			appID, name, relDate, yearStr, fmt.Sprintf("%.2f", price), strconv.FormatBool(isFree),
			owners, bracket, strconv.FormatInt(reviews, 10), scoreStr,
			fmt.Sprintf("%.2f", playHours), strconv.FormatBool(win), strconv.FormatBool(mac), strconv.FormatBool(lin),
			genresStr, catsStr,
		})
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (s *Storage) queryDistribution(ctx context.Context, sqlQuery string, args []interface{}, total int64) []domain.DistributionItem {
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []domain.DistributionItem
	for rows.Next() {
		var k, l sql.NullString
		var c int64
		if err := rows.Scan(&k, &l, &c); err == nil && k.Valid {
			pct := 0.0
			if total > 0 {
				pct = math.Round((float64(c)/float64(total))*1000) / 10
			}
			items = append(items, domain.DistributionItem{
				Key:        k.String,
				Label:      l.String,
				Count:      c,
				Percentage: pct,
			})
		}
	}
	return items
}

func (s *Storage) queryPlatformCombinations(ctx context.Context, where string, args []interface{}, total int64) []domain.PlatformCombination {
	query := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN windows = true AND mac = true AND linux = true THEN 'All 3 (Win, Mac, Linux)'
				WHEN windows = true AND mac = true AND linux = false THEN 'Windows + macOS'
				WHEN windows = true AND mac = false AND linux = true THEN 'Windows + Linux'
				WHEN windows = true AND mac = false AND linux = false THEN 'Windows only'
				WHEN windows = false AND mac = true AND linux = false THEN 'macOS only'
				WHEN windows = false AND mac = false AND linux = true THEN 'Linux only'
				ELSE 'Other'
			END as combo,
			count(*) as c
		FROM games
		WHERE %s
		GROUP BY combo
		ORDER BY c DESC
	`, where)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var res []domain.PlatformCombination
	for rows.Next() {
		var combo string
		var c int64
		if err := rows.Scan(&combo, &c); err == nil {
			pct := 0.0
			if total > 0 {
				pct = math.Round((float64(c)/float64(total))*1000) / 10
			}
			res = append(res, domain.PlatformCombination{
				Platforms:  combo,
				Count:      c,
				Percentage: pct,
			})
		}
	}
	return res
}

func (s *Storage) queryHistogram(ctx context.Context, sqlQuery string, args []interface{}, edges []float64) []domain.HistogramBin {
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var vals []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err == nil {
			vals = append(vals, v)
		}
	}
	return analytics.BuildHistogram(vals, edges)
}

func nullFloatPtr(f sql.NullFloat64) *float64 {
	if f.Valid {
		v := math.Round(f.Float64*100) / 100
		return &v
	}
	return nil
}

func toStringSlice(val interface{}) []string {
	if val == nil {
		return []string{}
	}
	if s, ok := val.([]string); ok {
		return s
	}
	if arr, ok := val.([]interface{}); ok {
		res := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				res = append(res, str)
			} else if item != nil {
				res = append(res, fmt.Sprintf("%v", item))
			}
		}
		return res
	}
	return []string{}
}

func analyticsQualityFlags(g domain.GameRecord) []string {
	var flags []string
	if g.ReleaseYear == nil {
		flags = append(flags, "missing_release_date")
	}
	if len(g.Genres) == 0 {
		flags = append(flags, "missing_genre")
	}
	if len(g.Categories) == 0 {
		flags = append(flags, "missing_category")
	}
	if g.OwnersBracket == "0 - 0" {
		flags = append(flags, "zero_owners")
	}
	if g.TotalReviews == 0 {
		flags = append(flags, "zero_reviews")
	}
	if g.MedianPlaytimeForever == 0 {
		flags = append(flags, "zero_playtime")
	}
	if g.IsPlaytest {
		flags = append(flags, "probable_playtest")
	}
	if g.IsSparse {
		flags = append(flags, "sparse_record")
	}
	return flags
}
