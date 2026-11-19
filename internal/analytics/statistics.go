package analytics

import (
	"fmt"
	"math"
	"sort"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

// CalculatePercentiles calculates p25, p50, p75 using continuous linear interpolation (R-7 / standard quantile).
// The input slice does NOT need to be pre-sorted (it will be copied and sorted).
func CalculatePercentiles(values []float64, metricName, unit string, totalCohortSize int64) domain.PercentileMetric {
	validCount := int64(len(values))
	missingCount := totalCohortSize - validCount
	if missingCount < 0 {
		missingCount = 0
	}

	result := domain.PercentileMetric{
		Metric:       metricName,
		Unit:         unit,
		SampleCount:  validCount,
		MissingCount: missingCount,
	}

	if validCount == 0 {
		return result
	}

	// Copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	minVal := sorted[0]
	maxVal := sorted[len(sorted)-1]
	result.Min = &minVal
	result.Max = &maxVal

	var sum float64
	for _, v := range sorted {
		sum += v
	}
	mean := sum / float64(validCount)
	result.Mean = &mean

	p25 := quantile(sorted, 0.25)
	p50 := quantile(sorted, 0.50)
	p75 := quantile(sorted, 0.75)

	result.P25 = &p25
	result.P50 = &p50
	result.P75 = &p75

	return result
}

// quantile computes the p-th quantile (0 <= p <= 1) using continuous linear interpolation.
func quantile(sorted []float64, p float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sorted[0]
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[n-1]
	}

	// R-7 method: index = p * (n - 1)
	idx := p * float64(n-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	weight := idx - float64(lower)

	if lower == upper {
		return sorted[lower]
	}
	return sorted[lower]*(1.0-weight) + sorted[upper]*weight
}

// BuildHistogram generates uniform or custom bounded bins for numeric values.
func BuildHistogram(values []float64, binEdges []float64) []domain.HistogramBin {
	if len(binEdges) < 2 {
		return nil
	}

	numBins := len(binEdges) - 1
	counts := make([]int64, numBins)
	total := int64(len(values))

	for _, v := range values {
		for i := 0; i < numBins; i++ {
			lower := binEdges[i]
			upper := binEdges[i+1]
			// Last bin includes upper bound
			if i == numBins-1 {
				if v >= lower && v <= upper {
					counts[i]++
					break
				}
			} else {
				if v >= lower && v < upper {
					counts[i]++
					break
				}
			}
		}
	}

	bins := make([]domain.HistogramBin, numBins)
	for i := 0; i < numBins; i++ {
		lower := binEdges[i]
		upper := binEdges[i+1]
		pct := 0.0
		if total > 0 {
			pct = (float64(counts[i]) / float64(total)) * 100.0
		}
		label := formatBinLabel(lower, upper, i == numBins-1)
		bins[i] = domain.HistogramBin{
			Lower:      lower,
			Upper:      upper,
			Label:      label,
			Count:      counts[i],
			Percentage: math.Round(pct*10) / 10,
		}
	}

	return bins
}

func formatBinLabel(lower, upper float64, isLast bool) string {
	if lower == 0 && upper >= 1000 {
		return fmt.Sprintf("$%.0f+", lower)
	}
	if isLast && upper >= 1000 {
		return fmt.Sprintf("%.0f+", lower)
	}
	return fmt.Sprintf("%.0f - %.0f", lower, upper)
}

// BuildCommonPriceRanges groups paid games into market standard price points.
func BuildCommonPriceRanges(prices []float64) []domain.DistributionItem {
	var (
		cUnder5 int64
		c5to10  int64
		c10to15 int64
		c15to20 int64
		c20to30 int64
		c30Plus int64
	)
	total := int64(len(prices))

	for _, p := range prices {
		switch {
		case p < 5.0:
			cUnder5++
		case p < 10.0:
			c5to10++
		case p < 15.0:
			c10to15++
		case p < 20.0:
			c15to20++
		case p < 30.0:
			c20to30++
		default:
			c30Plus++
		}
	}

	calcPct := func(c int64) float64 {
		if total == 0 {
			return 0
		}
		return math.Round((float64(c)/float64(total))*1000) / 10
	}

	return []domain.DistributionItem{
		{Key: "under_5", Label: "$0.01 - $4.99", Count: cUnder5, Percentage: calcPct(cUnder5)},
		{Key: "5_to_10", Label: "$5.00 - $9.99", Count: c5to10, Percentage: calcPct(c5to10)},
		{Key: "10_to_15", Label: "$10.00 - $14.99", Count: c10to15, Percentage: calcPct(c10to15)},
		{Key: "15_to_20", Label: "$15.00 - $19.99", Count: c15to20, Percentage: calcPct(c15to20)},
		{Key: "20_to_30", Label: "$20.00 - $29.99", Count: c20to30, Percentage: calcPct(c20to30)},
		{Key: "30_plus", Label: "$30.00+", Count: c30Plus, Percentage: calcPct(c30Plus)},
	}
}

// CheckCohortWarnings inspects sample counts and data quality to generate clear analytical caveats.
func CheckCohortWarnings(cohortName string, totalGames, validSamples int64) []string {
	var warnings []string
	if totalGames < 10 {
		warnings = append(warnings, fmt.Sprintf("Cohort %q has very few games (%d total). Statistical distributions may be volatile and sensitive to individual outliers.", cohortName, totalGames))
	} else if totalGames < 30 {
		warnings = append(warnings, fmt.Sprintf("Cohort %q has a modest sample size (%d total). Consider broadening filters if looking for general market trends.", cohortName, totalGames))
	}

	if totalGames > 0 && validSamples > 0 {
		missingRate := float64(totalGames-validSamples) / float64(totalGames)
		if missingRate > 0.5 {
			warnings = append(warnings, fmt.Sprintf("Over 50%% (%.1f%%) of games in cohort %q have missing or zero values for this metric.", missingRate*100, cohortName))
		}
	}

	return warnings
}
