package cohort

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/laymanmax/steam-cohort-bench/internal/domain"
)

// CacheEntry holds a cached benchmark result and expiration time.
type CacheEntry struct {
	Result    *domain.CohortBenchmarkResult
	ExpiresAt time.Time
}

// MemoryCache is an in-memory thread-safe cache for cohort benchmark results.
type MemoryCache struct {
	mu          sync.RWMutex
	items       map[string]CacheEntry
	order       []string // for basic eviction when max capacity reached
	ttl         time.Duration
	maxCapacity int
	revision    string
}

// NewMemoryCache creates a new cohort result cache.
func NewMemoryCache(ttl time.Duration, maxCapacity int) *MemoryCache {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if maxCapacity <= 0 {
		maxCapacity = 500
	}
	return &MemoryCache{
		items:       make(map[string]CacheEntry),
		order:       make([]string, 0, maxCapacity),
		ttl:         ttl,
		maxCapacity: maxCapacity,
	}
}

// SetRevision sets the current dataset revision and purges cache if revision changed.
func (c *MemoryCache) SetRevision(rev string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.revision != rev {
		c.revision = rev
		c.items = make(map[string]CacheEntry)
		c.order = c.order[:0]
	}
}

// GenerateCacheKey produces a deterministic hash from the dataset revision and cohort filter.
func GenerateCacheKey(revision string, filter *domain.CohortFilter) string {
	data, _ := json.Marshal(filter)
	h := sha256.New()
	h.Write([]byte(revision))
	h.Write([]byte(":"))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Get retrieves a cached cohort result if not expired.
func (c *MemoryCache) Get(key string) (*domain.CohortBenchmarkResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Result, true
}

// Set stores a benchmark result with the configured TTL.
func (c *MemoryCache) Set(key string, result *domain.CohortBenchmarkResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If over capacity, evict oldest
	if len(c.items) >= c.maxCapacity {
		if len(c.order) > 0 {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.items, oldest)
		}
	}

	c.items[key] = CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	c.order = append(c.order, key)
}

// Clear empties the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheEntry)
	c.order = c.order[:0]
}
