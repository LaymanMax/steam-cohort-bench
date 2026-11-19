package domain

import "time"

// QualityMode defines sample quality filtering mode.
type QualityMode string

const (
	QualityModeAll              QualityMode = "all"
	QualityModeExcludePlaytests QualityMode = "exclude_playtests"
	QualityModeMoreComplete     QualityMode = "more_complete"
)

// PriceType defines filter on free vs paid games.
type PriceType string

const (
	PriceTypeAll  PriceType = "all"
	PriceTypeFree PriceType = "free"
	PriceTypePaid PriceType = "paid"
)

// CohortFilter defines the multi-dimensional query criteria for selecting a cohort.
type CohortFilter struct {
	ID             string      `json:"id,omitempty"`
	Name           string      `json:"name,omitempty"`
	Color          string      `json:"color,omitempty"`
	Genres         []string    `json:"genres,omitempty"`
	Categories     []string    `json:"categories,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	ReleaseYearMin *int        `json:"release_year_min,omitempty"`
	ReleaseYearMax *int        `json:"release_year_max,omitempty"`
	PriceMin       *float64    `json:"price_min,omitempty"`
	PriceMax       *float64    `json:"price_max,omitempty"`
	PriceType      PriceType   `json:"price_type,omitempty"`
	Windows        *bool       `json:"windows,omitempty"`
	Mac            *bool       `json:"mac,omitempty"`
	Linux          *bool       `json:"linux,omitempty"`
	OwnersBrackets []string    `json:"owners_brackets,omitempty"`
	ReviewCountMin *int64      `json:"review_count_min,omitempty"`
	ReviewCountMax *int64      `json:"review_count_max,omitempty"`
	QualityMode    QualityMode `json:"quality_mode,omitempty"`
}

// PercentileMetric holds statistical percentile values with explicit sample and missing counts.
type PercentileMetric struct {
	Metric       string   `json:"metric"`
	Unit         string   `json:"unit"`
	Description  string   `json:"description,omitempty"`
	SampleCount  int64    `json:"sample_count"`
	MissingCount int64    `json:"missing_count"`
	P25          *float64 `json:"p25"`
	P50          *float64 `json:"p50"` // Median
	P75          *float64 `json:"p75"`
	Mean         *float64 `json:"mean,omitempty"`
	Min          *float64 `json:"min,omitempty"`
	Max          *float64 `json:"max,omitempty"`
}

// DistributionItem represents a single category or bucket count in a distribution.
type DistributionItem struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// HistogramBin represents a numeric bin for histograms.
type HistogramBin struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// MarketSizeSection holds high-level counts and distributions.
type MarketSizeSection struct {
	TotalGames      int64              `json:"total_games"`
	DistinctAppIDs  int64              `json:"distinct_app_ids"`
	EligibleGames   int64              `json:"eligible_games"`
	ReleaseYearDist []DistributionItem `json:"release_year_distribution"`
	GenreDist       []DistributionItem `json:"genre_distribution"`
	CategoryDist    []DistributionItem `json:"category_distribution"`
}

// PricingSection holds pricing percentiles, free/paid breakdown, and histograms.
type PricingSection struct {
	FreeGamesCount int64              `json:"free_games_count"`
	FreeGamesPct   float64            `json:"free_games_percentage"`
	PaidGamesCount int64              `json:"paid_games_count"`
	PaidGamesPct   float64            `json:"paid_games_percentage"`
	PaidPriceStats PercentileMetric   `json:"paid_price_stats"`
	PriceHistogram []HistogramBin     `json:"price_histogram"`
	CommonRanges   []DistributionItem `json:"common_price_ranges"`
}

// AudienceSection holds ownership interval distribution and median bracket.
type AudienceSection struct {
	OwnershipDist       []DistributionItem `json:"ownership_bracket_distribution"`
	MedianOwnersBracket string             `json:"median_ownership_bracket"`
	TotalWithOwnership  int64              `json:"total_with_ownership"`
}

// PlaytimeSection holds reported median playtime statistics in hours (and internal minutes).
type PlaytimeSection struct {
	ForeverStatsHours  PercentileMetric   `json:"playtime_forever_stats_hours"`
	TwoWeeksStatsHours PercentileMetric   `json:"playtime_2weeks_stats_hours"`
	PlaytimeHistogram  []HistogramBin     `json:"playtime_forever_histogram"`
	FreeVsPaidPlaytime []DistributionItem `json:"free_vs_paid_playtime_median"`
}

// ReviewsSection holds review volume percentiles and positive score distributions.
type ReviewsSection struct {
	TotalReviewStats PercentileMetric   `json:"total_reviews_stats"`
	PositiveScorePct PercentileMetric   `json:"positive_score_percentage_stats"`
	ReviewHistogram  []HistogramBin     `json:"review_count_histogram"`
	ReviewBrackets   []DistributionItem `json:"review_volume_brackets"`
}

// PlatformCombination represents games supporting a particular OS combo.
type PlatformCombination struct {
	Platforms  string  `json:"platforms"` // e.g. "Windows only", "Windows + Mac", "All 3"
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// PlatformSection holds platform availability metrics.
type PlatformSection struct {
	WindowsCount int64                 `json:"windows_count"`
	WindowsPct   float64               `json:"windows_percentage"`
	MacCount     int64                 `json:"mac_count"`
	MacPct       float64               `json:"mac_percentage"`
	LinuxCount   int64                 `json:"linux_count"`
	LinuxPct     float64               `json:"linux_percentage"`
	KnownSample  int64                 `json:"known_sample_count"`
	Combinations []PlatformCombination `json:"platform_combinations"`
}

// LocalizationSection holds language statistics.
type LocalizationSection struct {
	ListedLanguageCountStats PercentileMetric   `json:"listed_language_count_stats"`
	FullAudioCountStats      PercentileMetric   `json:"full_audio_count_stats"`
	TopSupportedLanguages    []DistributionItem `json:"top_supported_languages"`
	TopAudioLanguages        []DistributionItem `json:"top_audio_languages"`
}

// CohortBenchmarkResult contains all analytical sections for a single cohort.
type CohortBenchmarkResult struct {
	CohortID        string              `json:"cohort_id"`
	CohortName      string              `json:"cohort_name"`
	Color           string              `json:"color,omitempty"`
	Filter          CohortFilter        `json:"filter"`
	ExecutionTimeMs int64               `json:"execution_time_ms"`
	MarketSize      MarketSizeSection   `json:"market_size"`
	Pricing         PricingSection      `json:"pricing"`
	Audience        AudienceSection     `json:"audience"`
	Playtime        PlaytimeSection     `json:"playtime"`
	Reviews         ReviewsSection      `json:"reviews"`
	Platforms       PlatformSection     `json:"platforms"`
	Localization    LocalizationSection `json:"localization"`
	Warnings        []string            `json:"warnings,omitempty"`
}

// CohortComparisonResult contains results for up to 4 cohorts side-by-side.
type CohortComparisonResult struct {
	Cohorts  []CohortBenchmarkResult `json:"cohorts"`
	Warnings []string                `json:"warnings,omitempty"`
}

// HypotheticalProfile defines user-provided specifications for a prospective game.
type HypotheticalProfile struct {
	Name                  string   `json:"name,omitempty"`
	Genre                 string   `json:"genre"`
	SecondaryGenre        string   `json:"secondary_genre,omitempty"`
	ReleaseYear           int      `json:"release_year"`
	ProposedPrice         float64  `json:"proposed_price"`
	IsFree                bool     `json:"is_free"`
	Windows               bool     `json:"windows"`
	Mac                   bool     `json:"mac"`
	Linux                 bool     `json:"linux"`
	LocalizationLanguages []string `json:"localization_languages"`
	TargetOwnersBracket   string   `json:"target_owners_bracket,omitempty"`
	TargetReviewCount     *int64   `json:"target_review_count,omitempty"`
}

// ProfileBenchmarkResult compares a hypothetical profile to its comparable cohort.
type ProfileBenchmarkResult struct {
	Profile       HypotheticalProfile        `json:"profile"`
	CohortSummary MarketSizeSection          `json:"cohort_summary"`
	Pricing       ProfilePricingContext      `json:"pricing_context"`
	Audience      AudienceSection            `json:"audience_context"`
	Playtime      PlaytimeSection            `json:"playtime_context"`
	Platforms     ProfilePlatformContext     `json:"platform_context"`
	Localization  ProfileLocalizationContext `json:"localization_context"`
	Warnings      []string                   `json:"warnings,omitempty"`
}

type ProfilePricingContext struct {
	ProposedPrice       float64          `json:"proposed_price"`
	CohortStats         PercentileMetric `json:"cohort_paid_price_stats"`
	PositionDescription string           `json:"position_description"` // "Below p25", "Within IQR (p25-p75)", "Above p75"
	PercentileRankEst   float64          `json:"percentile_rank_estimate"`
}

type ProfilePlatformContext struct {
	ProposedCombination string                `json:"proposed_combination"`
	CohortCombinations  []PlatformCombination `json:"cohort_combinations"`
	IsCommonCombination bool                  `json:"is_common_combination"`
	WindowsRate         float64               `json:"windows_rate"`
	MacRate             float64               `json:"mac_rate"`
	LinuxRate           float64               `json:"linux_rate"`
}

type ProfileLocalizationContext struct {
	LanguageCount        int                `json:"proposed_language_count"`
	CohortLangCountStats PercentileMetric   `json:"cohort_language_count_stats"`
	LanguagesAnalysis    []DistributionItem `json:"languages_analysis"`
}

// GameRecord represents a full game record in the analytical database.
type GameRecord struct {
	AppID                  string   `json:"app_id"`
	Name                   string   `json:"name"`
	ReleaseDate            string   `json:"release_date"`
	ReleaseYear            *int     `json:"release_year"`
	Price                  float64  `json:"price"`
	IsFree                 bool     `json:"is_free"`
	EstimatedOwners        string   `json:"estimated_owners"`
	OwnersLower            int64    `json:"owners_lower"`
	OwnersUpper            int64    `json:"owners_upper"`
	OwnersBracket          string   `json:"owners_bracket"`
	PeakCCU                int64    `json:"peak_ccu"`
	RequiredAge            int64    `json:"required_age"`
	DLCCount               int64    `json:"dlc_count"`
	Windows                bool     `json:"windows"`
	Mac                    bool     `json:"mac"`
	Linux                  bool     `json:"linux"`
	Positive               int64    `json:"positive"`
	Negative               int64    `json:"negative"`
	TotalReviews           int64    `json:"total_reviews"`
	ReviewScorePct         *float64 `json:"review_score_pct"`
	MetacriticScore        *int64   `json:"metacritic_score"`
	AveragePlaytimeForever int64    `json:"average_playtime_forever"`
	MedianPlaytimeForever  int64    `json:"median_playtime_forever"`
	MedianPlaytimeHours    float64  `json:"median_playtime_hours"`
	Genres                 []string `json:"genres"`
	Categories             []string `json:"categories"`
	SupportedLanguages     []string `json:"supported_languages"`
	FullAudioLanguages     []string `json:"full_audio_languages"`
	Developers             []string `json:"developers"`
	Publishers             []string `json:"publishers"`
	HeaderImage            string   `json:"header_image,omitempty"`
	ShortDescription       string   `json:"short_description,omitempty"`
	Website                string   `json:"website,omitempty"`
	IsPlaytest             bool     `json:"is_playtest"`
	IsSparse               bool     `json:"is_sparse"`
	QualityFlags           []string `json:"quality_flags"`
}

// GameCatalogResponse is a paginated list of games.
type GameCatalogResponse struct {
	TotalCount int64        `json:"total_count"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
	Games      []GameRecord `json:"games"`
}

// DataQualityReport provides full visibility into data anomalies and completeness.
type DataQualityReport struct {
	SourceRowCount         int64              `json:"source_row_count"`
	DistinctAppIDs         int64              `json:"distinct_app_ids"`
	DuplicateIDs           int64              `json:"duplicate_ids"`
	ReadmeReportedCount    int64              `json:"readme_reported_count"`
	DiscrepancyExplanation string             `json:"discrepancy_explanation"`
	ProbablePlaytestCount  int64              `json:"probable_playtest_count"`
	SparseRecordsCount     int64              `json:"sparse_records_count"`
	MissingReleaseDates    int64              `json:"missing_release_dates"`
	MissingGenres          int64              `json:"missing_genres"`
	MissingCategories      int64              `json:"missing_categories"`
	MissingPrice           int64              `json:"missing_price"`
	ZeroOwnersCount        int64              `json:"zero_owners_count"`
	MissingReviews         int64              `json:"missing_reviews"`
	MissingPlaytime        int64              `json:"missing_playtime"`
	FieldCompleteness      []DistributionItem `json:"field_completeness"`
	IngestionWarnings      []string           `json:"ingestion_warnings"`
}

// DatasetManifest represents provenance metadata for the currently loaded analytical dataset.
type DatasetManifest struct {
	DatasetName         string    `json:"dataset_name"`
	DatasetRepository   string    `json:"dataset_repository"`
	Revision            string    `json:"revision"`
	SourceFileChecksum  string    `json:"source_file_checksum"`
	SourceFilename      string    `json:"source_filename"`
	SourceFormat        string    `json:"source_format"`
	SourceRowCount      int64     `json:"source_row_count"`
	NormalizedRowCount  int64     `json:"normalized_row_count"`
	DistinctAppIDs      int64     `json:"distinct_app_ids"`
	DuplicateIDs        int64     `json:"duplicate_ids"`
	IngestedAt          time.Time `json:"ingested_at"`
	License             string    `json:"license"`
	Attribution         string    `json:"attribution"`
	PriceUnit           string    `json:"price_unit"`
	ReadmeCount         int64     `json:"readme_count"`
	ViewerCount         int64     `json:"viewer_count"`
	AvailableGenres     []string  `json:"available_genres"`
	AvailableCategories []string  `json:"available_categories"`
	AvailableYears      []int     `json:"available_years"`
	AvailableBrackets   []string  `json:"available_brackets"`
}
