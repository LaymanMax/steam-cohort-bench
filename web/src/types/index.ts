export type QualityMode = "all" | "exclude_playtests" | "more_complete";
export type PriceType = "all" | "free" | "paid";

export interface CohortFilter {
  id?: string;
  name?: string;
  color?: string;
  genres?: string[];
  categories?: string[];
  tags?: string[];
  release_year_min?: number;
  release_year_max?: number;
  price_min?: number;
  price_max?: number;
  price_type?: PriceType;
  windows?: boolean;
  mac?: boolean;
  linux?: boolean;
  owners_brackets?: string[];
  review_count_min?: number;
  review_count_max?: number;
  quality_mode?: QualityMode;
}

export interface PercentileMetric {
  metric: string;
  unit: string;
  description?: string;
  sample_count: number;
  missing_count: number;
  p25?: number | null;
  p50?: number | null; // Median
  p75?: number | null;
  mean?: number | null;
  min?: number | null;
  max?: number | null;
}

export interface DistributionItem {
  key: string;
  label: string;
  count: number;
  percentage: number;
}

export interface HistogramBin {
  lower: number;
  upper: number;
  label: string;
  count: number;
  percentage: number;
}

export interface MarketSizeSection {
  total_games: number;
  distinct_app_ids: number;
  eligible_games: number;
  release_year_distribution: DistributionItem[];
  genre_distribution: DistributionItem[];
  category_distribution: DistributionItem[];
}

export interface PricingSection {
  free_games_count: number;
  free_games_percentage: number;
  paid_games_count: number;
  paid_games_percentage: number;
  paid_price_stats: PercentileMetric;
  price_histogram: HistogramBin[];
  common_price_ranges: DistributionItem[];
}

export interface AudienceSection {
  ownership_bracket_distribution: DistributionItem[];
  median_ownership_bracket: string;
  total_with_ownership: number;
}

export interface PlaytimeSection {
  playtime_forever_stats_hours: PercentileMetric;
  playtime_2weeks_stats_hours: PercentileMetric;
  playtime_forever_histogram: HistogramBin[];
  free_vs_paid_playtime_median: DistributionItem[];
}

export interface ReviewsSection {
  total_reviews_stats: PercentileMetric;
  positive_score_percentage_stats: PercentileMetric;
  review_count_histogram: HistogramBin[];
  review_volume_brackets: DistributionItem[];
}

export interface PlatformCombination {
  platforms: string;
  count: number;
  percentage: number;
}

export interface PlatformSection {
  windows_count: number;
  windows_percentage: number;
  mac_count: number;
  mac_percentage: number;
  linux_count: number;
  linux_percentage: number;
  known_sample_count: number;
  platform_combinations: PlatformCombination[];
}

export interface LocalizationSection {
  listed_language_count_stats: PercentileMetric;
  full_audio_count_stats: PercentileMetric;
  top_supported_languages: DistributionItem[];
  top_audio_languages: DistributionItem[];
}

export interface CohortBenchmarkResult {
  cohort_id: string;
  cohort_name: string;
  color?: string;
  filter: CohortFilter;
  execution_time_ms: number;
  market_size: MarketSizeSection;
  pricing: PricingSection;
  audience: AudienceSection;
  playtime: PlaytimeSection;
  reviews: ReviewsSection;
  platforms: PlatformSection;
  localization: LocalizationSection;
  warnings?: string[];
}

export interface CohortComparisonResult {
  cohorts: CohortBenchmarkResult[];
  warnings?: string[];
}

export interface HypotheticalProfile {
  name?: string;
  genre: string;
  secondary_genre?: string;
  release_year: number;
  proposed_price: number;
  is_free: boolean;
  windows: boolean;
  mac: boolean;
  linux: boolean;
  localization_languages: string[];
  target_owners_bracket?: string;
  target_review_count?: number;
}

export interface ProfileBenchmarkResult {
  profile: HypotheticalProfile;
  cohort_summary: MarketSizeSection;
  pricing_context: {
    proposed_price: number;
    cohort_paid_price_stats: PercentileMetric;
    position_description: string;
    percentile_rank_estimate: number;
  };
  audience_context: AudienceSection;
  playtime_context: PlaytimeSection;
  platform_context: {
    proposed_combination: string;
    cohort_combinations: PlatformCombination[];
    is_common_combination: boolean;
    windows_rate: number;
    mac_rate: number;
    linux_rate: number;
  };
  localization_context: {
    proposed_language_count: number;
    cohort_language_count_stats: PercentileMetric;
    languages_analysis: DistributionItem[];
  };
  warnings?: string[];
}

export interface GameRecord {
  app_id: string;
  name: string;
  release_date: string;
  release_year?: number | null;
  price: number;
  is_free: boolean;
  estimated_owners: string;
  owners_lower: number;
  owners_upper: number;
  owners_bracket: string;
  peak_ccu: number;
  required_age: number;
  dlc_count: number;
  windows: boolean;
  mac: boolean;
  linux: boolean;
  positive: number;
  negative: number;
  total_reviews: number;
  review_score_pct?: number | null;
  metacritic_score?: number | null;
  average_playtime_forever: number;
  median_playtime_forever: number;
  median_playtime_hours: number;
  genres: string[];
  categories: string[];
  supported_languages: string[];
  full_audio_languages: string[];
  developers: string[];
  publishers: string[];
  header_image?: string;
  short_description?: string;
  website?: string;
  is_playtest: boolean;
  is_sparse: boolean;
  quality_flags: string[];
}

export interface GameCatalogResponse {
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
  games: GameRecord[];
}

export interface DataQualityReport {
  source_row_count: number;
  distinct_app_ids: number;
  duplicate_ids: number;
  readme_reported_count: number;
  discrepancy_explanation: string;
  probable_playtest_count: number;
  sparse_records_count: number;
  missing_release_dates: number;
  missing_genres: number;
  missing_categories: number;
  missing_price: number;
  zero_owners_count: number;
  missing_reviews: number;
  missing_playtime: number;
  field_completeness: DistributionItem[];
  ingestion_warnings: string[];
}

export interface DatasetManifest {
  dataset_name: string;
  dataset_repository: string;
  revision: string;
  source_file_checksum: string;
  source_filename: string;
  source_format: string;
  source_row_count: number;
  normalized_row_count: number;
  distinct_app_ids: number;
  duplicate_ids: number;
  ingested_at: string;
  license: string;
  attribution: string;
  price_unit: string;
  readme_count: number;
  viewer_count: number;
  available_genres: string[];
  available_categories: string[];
  available_years: number[];
  available_brackets: string[];
}
