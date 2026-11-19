package analytics

import (
	"math"
	"testing"
)

func TestCalculatePercentiles(t *testing.T) {
	// Sample data: [10, 20, 30, 40, 50]
	// n = 5
	// p25: idx = 0.25 * 4 = 1 -> 20.0
	// p50: idx = 0.50 * 4 = 2 -> 30.0
	// p75: idx = 0.75 * 4 = 3 -> 40.0
	vals := []float64{10, 20, 30, 40, 50}
	stats := CalculatePercentiles(vals, "test_metric", "dollars", 5)

	if stats.SampleCount != 5 {
		t.Fatalf("expected SampleCount=5, got %d", stats.SampleCount)
	}
	if stats.MissingCount != 0 {
		t.Fatalf("expected MissingCount=0, got %d", stats.MissingCount)
	}
	if stats.P25 == nil || *stats.P25 != 20.0 {
		t.Errorf("expected P25=20.0, got %v", stats.P25)
	}
	if stats.P50 == nil || *stats.P50 != 30.0 {
		t.Errorf("expected P50=30.0, got %v", stats.P50)
	}
	if stats.P75 == nil || *stats.P75 != 40.0 {
		t.Errorf("expected P75=40.0, got %v", stats.P75)
	}
	if stats.Mean == nil || *stats.Mean != 30.0 {
		t.Errorf("expected Mean=30.0, got %v", stats.Mean)
	}
}

func TestCalculatePercentiles_Empty(t *testing.T) {
	stats := CalculatePercentiles([]float64{}, "test_metric", "dollars", 10)
	if stats.SampleCount != 0 {
		t.Errorf("expected SampleCount=0, got %d", stats.SampleCount)
	}
	if stats.MissingCount != 10 {
		t.Errorf("expected MissingCount=10, got %d", stats.MissingCount)
	}
	if stats.P25 != nil || stats.P50 != nil || stats.P75 != nil {
		t.Errorf("expected nil percentiles for empty slice, got %+v", stats)
	}
}

func TestCalculatePercentiles_SingleElement(t *testing.T) {
	stats := CalculatePercentiles([]float64{15.99}, "test_metric", "dollars", 1)
	if *stats.P25 != 15.99 || *stats.P50 != 15.99 || *stats.P75 != 15.99 {
		t.Errorf("expected 15.99 for all percentiles of single element, got p25=%v, p50=%v, p75=%v", stats.P25, stats.P50, stats.P75)
	}
}

func TestCalculatePercentiles_Interpolation(t *testing.T) {
	// [10, 20, 30, 40]
	// n = 4
	// p50: idx = 0.5 * 3 = 1.5 -> midpoint between index 1 (20) and index 2 (30) = 25.0
	vals := []float64{10, 20, 30, 40}
	stats := CalculatePercentiles(vals, "test_metric", "dollars", 4)
	if stats.P50 == nil || math.Abs(*stats.P50-25.0) > 1e-6 {
		t.Errorf("expected P50=25.0, got %v", stats.P50)
	}
}

func TestBuildHistogram(t *testing.T) {
	vals := []float64{1.0, 4.0, 7.5, 9.99, 14.99, 25.0}
	edges := []float64{0.0, 5.0, 10.0, 20.0, 50.0}
	bins := BuildHistogram(vals, edges)

	if len(bins) != 4 {
		t.Fatalf("expected 4 bins, got %d", len(bins))
	}
	// bin 0: [0, 5) -> 1.0, 4.0 -> count 2
	if bins[0].Count != 2 {
		t.Errorf("bin 0 count = %d, want 2", bins[0].Count)
	}
	// bin 1: [5, 10) -> 7.5, 9.99 -> count 2
	if bins[1].Count != 2 {
		t.Errorf("bin 1 count = %d, want 2", bins[1].Count)
	}
	// bin 2: [10, 20) -> 14.99 -> count 1
	if bins[2].Count != 1 {
		t.Errorf("bin 2 count = %d, want 1", bins[2].Count)
	}
	// bin 3: [20, 50] -> 25.0 -> count 1
	if bins[3].Count != 1 {
		t.Errorf("bin 3 count = %d, want 1", bins[3].Count)
	}
}

func TestBuildCommonPriceRanges(t *testing.T) {
	prices := []float64{2.99, 4.99, 9.99, 14.99, 19.99, 24.99, 49.99}
	ranges := BuildCommonPriceRanges(prices)
	if len(ranges) != 6 {
		t.Fatalf("expected 6 ranges, got %d", len(ranges))
	}
	if ranges[0].Count != 2 { // < 5.0 (2.99, 4.99)
		t.Errorf("under 5 count = %d, want 2", ranges[0].Count)
	}
	if ranges[1].Count != 1 { // 5-10 (9.99)
		t.Errorf("5-10 count = %d, want 1", ranges[1].Count)
	}
}
