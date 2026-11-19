package normalize

import (
	"testing"
)

func TestParseReleaseYear(t *testing.T) {
	tests := []struct {
		input     string
		wantYear  *int
		wantMonth *int
	}{
		{"Aug 1, 2023", intPtr(2023), intPtr(8)},
		{"Jul 29, 2016", intPtr(2016), intPtr(7)},
		{"May 6, 2019", intPtr(2019), intPtr(5)},
		{"2025-10-31", intPtr(2025), intPtr(10)},
		{"Nov 2024", intPtr(2024), intPtr(11)},
		{"Coming 2026", intPtr(2026), nil},
		{"", nil, nil},
		{"Q1 2025", intPtr(2025), nil},
		{"To be announced", nil, nil},
	}

	for _, tt := range tests {
		y, m := ParseReleaseYear(tt.input)
		if (y == nil && tt.wantYear != nil) || (y != nil && tt.wantYear == nil) || (y != nil && *y != *tt.wantYear) {
			t.Errorf("ParseReleaseYear(%q) year = %v, want %v", tt.input, y, tt.wantYear)
		}
		if tt.wantMonth != nil && (m == nil || *m != *tt.wantMonth) {
			t.Errorf("ParseReleaseYear(%q) month = %v, want %v", tt.input, m, tt.wantMonth)
		}
	}
}

func TestParseOwnershipRange(t *testing.T) {
	tests := []struct {
		input       string
		wantLower   int64
		wantUpper   int64
		wantBracket string
	}{
		{"0 - 0", 0, 0, "0 - 0"},
		{"0 - 20000", 0, 20000, "0 - 20K"},
		{"0 - 20,000", 0, 20000, "0 - 20K"},
		{"20000 - 50000", 20000, 50000, "20K - 50K"},
		{"1000000 - 2000000", 1000000, 2000000, "1M - 2M"},
		{"100000000 - 200000000", 100000000, 200000000, "100M+"},
		{"", 0, 0, "Unknown"},
	}

	for _, tt := range tests {
		res := ParseOwnershipRange(tt.input)
		if res.Lower != tt.wantLower || res.Upper != tt.wantUpper || res.Bracket != tt.wantBracket {
			t.Errorf("ParseOwnershipRange(%q) = (%d, %d, %s), want (%d, %d, %s)",
				tt.input, res.Lower, res.Upper, res.Bracket, tt.wantLower, tt.wantUpper, tt.wantBracket)
		}
	}
}

func TestDeriveReviews(t *testing.T) {
	tot, pct := DeriveReviews(100, 25)
	if tot != 125 {
		t.Errorf("DeriveReviews(100, 25) total = %d, want 125", tot)
	}
	if pct == nil || *pct != 80.0 {
		t.Errorf("DeriveReviews(100, 25) pct = %v, want 80.0", pct)
	}

	totZero, pctZero := DeriveReviews(0, 0)
	if totZero != 0 || pctZero != nil {
		t.Errorf("DeriveReviews(0, 0) = (%d, %v), want (0, nil)", totZero, pctZero)
	}
}

func TestIsProbablePlaytest(t *testing.T) {
	if !IsProbablePlaytest("Black Dragon Mage Playtest") {
		t.Errorf("expected Black Dragon Mage Playtest to be flagged")
	}
	if !IsProbablePlaytest("Awesome Game Closed Beta") {
		t.Errorf("expected Closed Beta to be flagged")
	}
	if IsProbablePlaytest("Portal 2") {
		t.Errorf("did not expect Portal 2 to be flagged")
	}
}

func intPtr(i int) *int {
	return &i
}
