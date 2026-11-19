package cohort

import (
	"testing"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

func TestCacheRevisionInvalidation(t *testing.T) {
	cache := NewMemoryCache(5*time.Minute, 100)
	cache.SetRevision("rev-1")

	filter := &domain.CohortFilter{
		Genres: []string{"Action"},
	}
	key := GenerateCacheKey("rev-1", filter)

	res := &domain.CohortBenchmarkResult{
		CohortID:   "test-1",
		CohortName: "Action Games",
	}
	cache.Set(key, res)

	// Fetch should succeed
	cached, ok := cache.Get(key)
	if !ok || cached.CohortID != "test-1" {
		t.Fatalf("expected to find cached item, got %v, %v", cached, ok)
	}

	// Change revision -> cache must be invalidated
	cache.SetRevision("rev-2")
	cached, ok = cache.Get(key)
	if ok || cached != nil {
		t.Fatalf("expected cache to be invalidated on revision change")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := NewMemoryCache(10*time.Millisecond, 100)
	cache.SetRevision("rev-1")

	filter := &domain.CohortFilter{Genres: []string{"Indie"}}
	key := GenerateCacheKey("rev-1", filter)

	cache.Set(key, &domain.CohortBenchmarkResult{CohortID: "indie"})
	time.Sleep(20 * time.Millisecond)

	_, ok := cache.Get(key)
	if ok {
		t.Fatalf("expected item to expire")
	}
}
