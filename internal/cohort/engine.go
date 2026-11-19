package cohort

import (
	"fmt"
	"strings"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

// QueryFilter contains the generated WHERE clause and parameter arguments.
type QueryFilter struct {
	WhereClause string
	Args        []interface{}
}

// BuildWhereClause constructs a parameterized SQL WHERE clause from a CohortFilter.
// It uses predefined allowlists and safe parameter binding to prevent any injection.
func BuildWhereClause(f domain.CohortFilter) QueryFilter {
	var conditions []string
	var args []interface{}

	// 1. Genres: matches games that contain ANY of the specified genres
	if len(f.Genres) > 0 {
		var genreLiterals []string
		for _, g := range f.Genres {
			trimmed := strings.TrimSpace(g)
			if trimmed != "" {
				args = append(args, trimmed)
				genreLiterals = append(genreLiterals, "?")
			}
		}
		if len(genreLiterals) > 0 {
			conditions = append(conditions, fmt.Sprintf("list_has_any(genres, [%s])", strings.Join(genreLiterals, ", ")))
		}
	}

	// 2. Categories: matches games that contain ANY of the specified categories
	if len(f.Categories) > 0 {
		var catLiterals []string
		for _, c := range f.Categories {
			trimmed := strings.TrimSpace(c)
			if trimmed != "" {
				args = append(args, trimmed)
				catLiterals = append(catLiterals, "?")
			}
		}
		if len(catLiterals) > 0 {
			conditions = append(conditions, fmt.Sprintf("list_has_any(categories, [%s])", strings.Join(catLiterals, ", ")))
		}
	}

	// 3. Release Year
	if f.ReleaseYearMin != nil {
		conditions = append(conditions, "release_year >= ?")
		args = append(args, *f.ReleaseYearMin)
	}
	if f.ReleaseYearMax != nil {
		conditions = append(conditions, "release_year <= ?")
		args = append(args, *f.ReleaseYearMax)
	}

	// 4. Price range
	if f.PriceMin != nil {
		conditions = append(conditions, "price >= ?")
		args = append(args, *f.PriceMin)
	}
	if f.PriceMax != nil {
		conditions = append(conditions, "price <= ?")
		args = append(args, *f.PriceMax)
	}

	// 5. Price type (free, paid, all)
	switch f.PriceType {
	case domain.PriceTypeFree:
		conditions = append(conditions, "is_free = true")
	case domain.PriceTypePaid:
		conditions = append(conditions, "is_free = false AND price > 0")
	}

	// 6. Platform support
	if f.Windows != nil {
		conditions = append(conditions, "windows = ?")
		args = append(args, *f.Windows)
	}
	if f.Mac != nil {
		conditions = append(conditions, "mac = ?")
		args = append(args, *f.Mac)
	}
	if f.Linux != nil {
		conditions = append(conditions, "linux = ?")
		args = append(args, *f.Linux)
	}

	// 7. Ownership brackets
	if len(f.OwnersBrackets) > 0 {
		var bracketPlaceholders []string
		for _, b := range f.OwnersBrackets {
			trimmed := strings.TrimSpace(b)
			if trimmed != "" {
				args = append(args, trimmed)
				bracketPlaceholders = append(bracketPlaceholders, "?")
			}
		}
		if len(bracketPlaceholders) > 0 {
			conditions = append(conditions, fmt.Sprintf("owners_bracket IN (%s)", strings.Join(bracketPlaceholders, ", ")))
		}
	}

	// 8. Review count
	if f.ReviewCountMin != nil {
		conditions = append(conditions, "total_reviews >= ?")
		args = append(args, *f.ReviewCountMin)
	}
	if f.ReviewCountMax != nil {
		conditions = append(conditions, "total_reviews <= ?")
		args = append(args, *f.ReviewCountMax)
	}

	// 9. Quality Mode
	switch f.QualityMode {
	case domain.QualityModeExcludePlaytests:
		conditions = append(conditions, "is_playtest = false")
	case domain.QualityModeMoreComplete:
		conditions = append(conditions, "is_playtest = false AND release_year IS NOT NULL AND (len(genres) > 0 OR len(categories) > 0)")
	default:
		// QualityModeAll -> no filter
	}

	whereClause := "1=1"
	if len(conditions) > 0 {
		whereClause = strings.Join(conditions, " AND ")
	}

	return QueryFilter{
		WhereClause: whereClause,
		Args:        args,
	}
}
