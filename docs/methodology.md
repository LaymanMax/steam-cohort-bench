# Statistical Methodology & Analytical Design

## 1. Principles of Distribution-Based Benchmarking

Market research for indie games is frequently distorted by simple averages. A handful of viral hits ($60 price tags, millions of reviews) pulls the arithmetic mean upward, presenting an unrealistic picture to an indie developer.

`steam-cohort-bench` enforces **distribution-based benchmarking**:

- **25th Percentile (P25)**: The lower quartile threshold.
- **50th Percentile (P50 / Median)**: The true middle observation.
- **75th Percentile (P75)**: The upper quartile threshold.
- **Interquartile Range (IQR)**: The middle 50% span between P25 and P75.

### Continuous Quantile Interpolation

All percentiles are calculated using continuous linear interpolation (the standard R-7 quantile method, natively implemented in DuckDB as `QUANTILE_CONT(x, p)`):
$$\text{index} = p \cdot (N - 1)$$
$$\text{quantile} = x_{\lfloor \text{index} \rfloor} \cdot (1 - w) + x_{\lceil \text{index} \rceil} \cdot w$$
where $w = \text{index} - \lfloor \text{index} \rfloor$.

---

## 2. Derivation of Specific Metrics

### A. Paid Game Pricing

- **Exclusion of Free Games**: Games with a price of `$0.00` are categorized as Free to Play. Including them in commercial price distributions would artificially push percentiles to zero.
- **Denominator Transparency**: Paid price percentiles report both the number of valid paid games (`sample_count`) and the number of excluded free games (`missing_count`).

### B. Ownership Intervals

- **No Pseudo-Precision**: SteamSpy estimates are reported in ranges (e.g. `20,000 - 50,000`). We never fabricate exact player counts or treat midpoints as ground truth.
- **Zero Ownership Representation**: Games reporting `0 - 0` are preserved as a distinct bracket rather than converted to `NULL`.

### C. Review Volume

- **Separation of Press Quotes**: The source dataset column `reviews` contains journalistic quotes. Player review volume is derived strictly from `positive + negative`.
- **Positive Sentiment Percentage**: Calculated as `positive / (positive + negative) * 100` only for titles with at least 1 review.

### D. Playtime Hours

- Steam logs playtime in minutes. For user readability, values are converted to hours (`minutes / 60.0`). Titles with zero reported minutes are excluded from playtime percentile denominators to reflect actual player behavior among players who launched the game.

### E. Platform Support

- Platform availability percentages are evaluated against titles with known operating system support, preventing missing metadata from diluting Linux or macOS prevalence rates.

---

## 3. Sample Quality Modes

The dataset includes thousands of sparse playtests, internal testing entries, and incomplete records. Users can choose among three documented quality modes:

1. **All Records (`all`)**: Evaluates every game in the dataset regardless of metadata sparsity.
2. **Exclude Probable Playtests (`exclude_playtests`)**: Excludes records matching playtest, beta, or alpha keywords in their title.
3. **More Complete Records (`more_complete`)**: Requires a valid release year, non-empty genre or category metadata, and excludes probable playtests.
