package normalize

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	yearRegex  = regexp.MustCompile(`\b(19\d\d|20\d\d)\b`)
	ownerRegex = regexp.MustCompile(`([\d,]+)\s*-\s*([\d,]+)`)
)

// ParsedOwnership contains parsed bounds and formatted label.
type ParsedOwnership struct {
	Original string
	Lower    int64
	Upper    int64
	Bracket  string
}

// ParseReleaseYear extracts the 4-digit release year and optional month from a date string.
func ParseReleaseYear(raw string) (*int, *int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	// Try standard date layouts first
	layouts := []string{
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan, 2006",
		"2 January, 2006",
		"Jan 2006",
		"January 2006",
		"2006-01-02",
		"2006/01/02",
		"02-01-2006",
		"02/01/2006",
		"2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			y := t.Year()
			m := int(t.Month())
			return &y, &m
		}
	}

	// Fallback to regex extraction for year
	matches := yearRegex.FindStringSubmatch(raw)
	if len(matches) > 1 {
		if y, err := strconv.Atoi(matches[1]); err == nil && y >= 1980 && y <= 2035 {
			return &y, nil
		}
	}

	return nil, nil
}

// ParseOwnershipRange parses the lower and upper bounds of an estimated owners string.
func ParseOwnershipRange(raw string) ParsedOwnership {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return ParsedOwnership{
			Original: "",
			Lower:    0,
			Upper:    0,
			Bracket:  "Unknown",
		}
	}

	matches := ownerRegex.FindStringSubmatch(cleaned)
	if len(matches) == 3 {
		lowerStr := strings.ReplaceAll(matches[1], ",", "")
		upperStr := strings.ReplaceAll(matches[2], ",", "")
		lower, err1 := strconv.ParseInt(lowerStr, 10, 64)
		upper, err2 := strconv.ParseInt(upperStr, 10, 64)
		if err1 == nil && err2 == nil {
			bracket := FormatOwnershipBracket(lower, upper)
			return ParsedOwnership{
				Original: cleaned,
				Lower:    lower,
				Upper:    upper,
				Bracket:  bracket,
			}
		}
	}

	return ParsedOwnership{
		Original: cleaned,
		Lower:    0,
		Upper:    0,
		Bracket:  cleaned,
	}
}

// FormatOwnershipBracket returns a consistent human-readable bracket label.
func FormatOwnershipBracket(lower, upper int64) string {
	if lower == 0 && upper == 0 {
		return "0 - 0"
	}
	if lower == 0 && upper <= 20000 {
		return "0 - 20K"
	}
	if lower >= 20000 && upper <= 50000 {
		return "20K - 50K"
	}
	if lower >= 50000 && upper <= 100000 {
		return "50K - 100K"
	}
	if lower >= 100000 && upper <= 200000 {
		return "100K - 200K"
	}
	if lower >= 200000 && upper <= 500000 {
		return "200K - 500K"
	}
	if lower >= 500000 && upper <= 1000000 {
		return "500K - 1M"
	}
	if lower >= 1000000 && upper <= 2000000 {
		return "1M - 2M"
	}
	if lower >= 2000000 && upper <= 5000000 {
		return "2M - 5M"
	}
	if lower >= 5000000 && upper <= 10000000 {
		return "5M - 10M"
	}
	if lower >= 10000000 && upper <= 20000000 {
		return "10M - 20M"
	}
	if lower >= 20000000 && upper <= 50000000 {
		return "20M - 50M"
	}
	if lower >= 50000000 && upper <= 100000000 {
		return "50M - 100M"
	}
	if lower >= 100000000 {
		return "100M+"
	}
	return strconv.FormatInt(lower, 10) + " - " + strconv.FormatInt(upper, 10)
}

// DeriveReviews calculates total review count and score percentage.
func DeriveReviews(pos, neg int64) (total int64, scorePct *float64) {
	if pos < 0 {
		pos = 0
	}
	if neg < 0 {
		neg = 0
	}
	total = pos + neg
	if total > 0 {
		pct := (float64(pos) / float64(total)) * 100.0
		return total, &pct
	}
	return 0, nil
}

// IsProbablePlaytest detects if a title indicates a playtest or internal test version.
func IsProbablePlaytest(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "playtest") ||
		strings.Contains(lower, "closed beta") ||
		strings.Contains(lower, "open beta") ||
		strings.Contains(lower, "alpha test") ||
		strings.Contains(lower, "internal test") ||
		strings.HasSuffix(lower, " test") ||
		strings.HasPrefix(lower, "test ")
}

// IsSparseRecord flags records that lack basic market metadata.
func IsSparseRecord(name string, genres, categories []string, owners string, price float64, totalReviews int64) bool {
	if len(genres) == 0 && len(categories) == 0 && (owners == "0 - 0" || owners == "") && price == 0 && totalReviews == 0 {
		return true
	}
	return false
}

// ClassifyQualityFlags generates individual data-quality tags for a game record.
func ClassifyQualityFlags(name string, releaseYear *int, genres, categories []string, owners string, price float64, totalReviews int64, playtimeForever int64, isPlaytest, isSparse bool) []string {
	var flags []string
	if releaseYear == nil {
		flags = append(flags, "missing_release_date")
	}
	if len(genres) == 0 {
		flags = append(flags, "missing_genre")
	}
	if len(categories) == 0 {
		flags = append(flags, "missing_category")
	}
	if owners == "0 - 0" {
		flags = append(flags, "zero_owners")
	} else if owners == "" {
		flags = append(flags, "missing_owners")
	}
	if totalReviews == 0 {
		flags = append(flags, "zero_reviews")
	}
	if playtimeForever == 0 {
		flags = append(flags, "zero_playtime")
	}
	if isPlaytest {
		flags = append(flags, "probable_playtest")
	}
	if isSparse {
		flags = append(flags, "sparse_record")
	}
	return flags
}
