package storage

import (
	"fmt"
	"sync"
	"time"
)

// cacheEntry holds a cached report with its insertion time.
type cacheEntry struct {
	data      []byte
	createdAt time.Time
}

// CachedStore wraps a Store with an in-memory LRU cache.
// Caches CID→report mappings to reduce redundant storage calls.
// Thread-safe for concurrent access.
type CachedStore struct {
	store   Store
	cache   map[string]*cacheEntry
	order   []string // LRU order: oldest first.
	mu      sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// CacheConfig holds configuration for the cache layer.
type CacheConfig struct {
	// MaxSize is the maximum number of entries in the cache. Default: 1000.
	MaxSize int
	// TTL is how long entries remain valid. Default: 1 hour.
	TTL time.Duration
}

// NewCachedStore wraps a Store with a caching layer.
func NewCachedStore(store Store, cfg *CacheConfig) *CachedStore {
	maxSize := 1000
	ttl := time.Hour

	if cfg != nil {
		if cfg.MaxSize > 0 {
			maxSize = cfg.MaxSize
		}
		if cfg.TTL > 0 {
			ttl = cfg.TTL
		}
	}

	return &CachedStore{
		store:   store,
		cache:   make(map[string]*cacheEntry),
		order:   make([]string, 0),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Upload stores a report via the underlying Store and caches the result.
func (c *CachedStore) Upload(report []byte) (string, error) {
	cidStr, err := c.store.Upload(report)
	if err != nil {
		return "", err
	}

	// Cache the uploaded report by its CID.
	c.put(cidStr, report)
	return cidStr, nil
}

// Retrieve fetches a report by CID, checking the cache first.
func (c *CachedStore) Retrieve(cidStr string) ([]byte, error) {
	// Check cache first.
	if data, ok := c.get(cidStr); ok {
		return data, nil
	}

	// Cache miss — fetch from the underlying store.
	data, err := c.store.Retrieve(cidStr)
	if err != nil {
		return nil, err
	}

	// Cache the retrieved report.
	c.put(cidStr, data)
	return data, nil
}

// get returns cached data if present and not expired.
func (c *CachedStore) get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// Check TTL.
	if time.Since(entry.createdAt) > c.ttl {
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		return nil, false
	}

	return entry.data, true
}

// put adds an entry to the cache, evicting the oldest if at capacity.
func (c *CachedStore) put(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already cached, update in place.
	if _, ok := c.cache[key]; ok {
		c.cache[key] = &cacheEntry{data: data, createdAt: time.Now()}
		return
	}

	// Evict oldest entries if at capacity.
	for len(c.cache) >= c.maxSize && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.cache, oldest)
	}

	c.cache[key] = &cacheEntry{data: data, createdAt: time.Now()}
	c.order = append(c.order, key)
}

// Size returns the current number of cached entries.
func (c *CachedStore) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Stats returns cache statistics.
func (c *CachedStore) Stats() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("entries=%d max=%d ttl=%s", len(c.cache), c.maxSize, c.ttl)
}
