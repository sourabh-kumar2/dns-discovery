// Package discovery provides an in-memory DNS cache for service discovery.
// It allows storing and retrieving DNS records with TTL support, ensuring
// responses are dynamically updated based on cached entries.
//
// This package is primarily used for resolving service endpoints by mapping
// domain names to predefined responses, such as TXT or A records.
package discovery

import (
	"fmt"
	"sync"
	"time"

	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

// Record represents a cached DNS response.
type Record struct {
	Value []byte
	TTL   time.Duration
}

// Cache stores DNS records with TTL support.
type Cache struct {
	mu     sync.RWMutex
	data   map[string]Record
	stopCh chan struct{}
}

// NewCache initializes a cache and starts a background goroutine
// to periodically reload records from a JSON file.
//
// - `filename`: Path to the JSON file containing DNS records.
// - `interval`: Frequency of cache updates (e.g., `30 * time.Second`).
//
// Call `cache.Stop()` to gracefully stop the background ticker.
func NewCache(filename string, interval time.Duration) *Cache {
	cache := &Cache{data: make(map[string]Record)}
	cache.stopCh = make(chan struct{})

	if err := cache.refresh(filename); err != nil {
		logger.Log(zap.ErrorLevel, "Failed to hydrate cache", zap.Error(err))
	}
	logger.Log(zap.InfoLevel, "Cache initialized", zap.String("filename", filename))

	go cache.startUpdater(filename, interval)
	return cache
}

// formatKey generates a unique key using QType and domain.
func formatKey(domain string, qType uint16) string {
	return fmt.Sprintf("__%d__.%s", qType, domain)
}

// Set stores a DNS record in the cache with a TTL.
func (c *Cache) Set(domain string, qType uint16, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := formatKey(domain, qType)
	c.data[key] = Record{
		Value: value,
		TTL:   ttl,
	}
}

// Get retrieves a DNS record from the cache if it exists and is not expired.
func (c *Cache) Get(domain string, qType uint16) *Record {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := formatKey(domain, qType)
	record, exists := c.data[key]
	if !exists {
		return nil
	}
	return &record
}

// Update the existing cache data.
func (c *Cache) Update(newRecords map[string]Record) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = newRecords
}

func (c *Cache) startUpdater(filename string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.refresh(filename)
			if err != nil {
				logger.Log(zap.WarnLevel, "Failed to load records", zap.Error(err))
				continue
			}
			logger.Log(zap.InfoLevel, "Cache updated dynamically")
		case <-c.stopCh:
			logger.Log(zap.InfoLevel, "Stopping cache update ticker")
			return
		}
	}
}

func (c *Cache) refresh(filename string) error {
	newRecords, err := loadFromFile(filename)
	if err != nil {
		logger.Log(zap.WarnLevel, "Failed to load records", zap.Error(err))
		return err
	}
	c.Update(newRecords)
	return nil
}

// Stop gracefully stops the background cache update process.
func (c *Cache) Stop() {
	close(c.stopCh)
	logger.Log(zap.InfoLevel, "Stopping cache")
}

// NewTestCache is for testing.
func NewTestCache() *Cache {
	return &Cache{data: make(map[string]Record)}
}
