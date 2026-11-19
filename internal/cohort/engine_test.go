package cohort

import (
	"strings"
	"testing"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

func TestBuildWhereClause(t *testing.T) {
	minY := 2020
	maxY := 2024
	minP := 4.99
	maxP := 19.99
	lin := true

	filter := domain.CohortFilter{
		Genres:         []string{"Strategy", "Simulation"},
		ReleaseYearMin: &minY,
		ReleaseYearMax: &maxY,
		PriceMin:       &minP,
		PriceMax:       &maxP,
		PriceType:      domain.PriceTypePaid,
		Linux:          &lin,
		QualityMode:    domain.QualityModeExcludePlaytests,
	}

	qf := BuildWhereClause(filter)

	if !strings.Contains(qf.WhereClause, "list_has_any(genres, [?, ?])") {
		t.Errorf("expected list_has_any(genres), got: %s", qf.WhereClause)
	}
	if !strings.Contains(qf.WhereClause, "release_year >= ?") || !strings.Contains(qf.WhereClause, "release_year <= ?") {
		t.Errorf("expected release year constraints, got: %s", qf.WhereClause)
	}
	if !strings.Contains(qf.WhereClause, "is_free = false AND price > 0") {
		t.Errorf("expected paid price constraint, got: %s", qf.WhereClause)
	}
	if !strings.Contains(qf.WhereClause, "linux = ?") {
		t.Errorf("expected linux constraint, got: %s", qf.WhereClause)
	}
	if !strings.Contains(qf.WhereClause, "is_playtest = false") {
		t.Errorf("expected playtest exclusion, got: %s", qf.WhereClause)
	}

	// Check arguments count
	// Genres: Strategy, Simulation (2)
	// Years: 2020, 2024 (2)
	// Prices: 4.99, 19.99 (2)
	// Linux: true (1)
	expectedArgsCount := 7
	if len(qf.Args) != expectedArgsCount {
		t.Errorf("expected %d args, got %d: %v", expectedArgsCount, len(qf.Args), qf.Args)
	}
}

func TestBuildWhereClause_Empty(t *testing.T) {
	qf := BuildWhereClause(domain.CohortFilter{})
	if qf.WhereClause != "1=1" {
		t.Errorf("expected 1=1 for empty filter, got %s", qf.WhereClause)
	}
	if len(qf.Args) != 0 {
		t.Errorf("expected 0 args for empty filter, got %d", len(qf.Args))
	}
}
