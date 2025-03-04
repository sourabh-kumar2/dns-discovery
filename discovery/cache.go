// Package discovery blah blah.
package discovery

import (
	"context"
	"fmt"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

// Record represents a cached DNS response.
type Record struct {
	Value      []byte
	Expiration time.Time
}

// Cache stores DNS records with TTL support.
type Cache struct {
	mu     sync.RWMutex
	data   map[string]Record
	ctx    context.Context
	cancel context.CancelFunc
}

// NewCache initializes a new cache instance.
func NewCache() *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &Cache{
		data:   make(map[string]Record),
		ctx:    ctx,
		cancel: cancel,
	}
	go cache.cleanupExpiredRecords()
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
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a DNS record from the cache if it exists and is not expired.
func (c *Cache) Get(domain string, qType uint16) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := formatKey(domain, qType)
	record, exists := c.data[key]
	if !exists || time.Now().After(record.Expiration) {
		return nil, false
	}
	return record.Value, true
}

// Shutdown graceful shutdown.
func (c *Cache) Shutdown() {
	c.cancel()
}

// cleanupExpiredRecords periodically removes expired entries.
func (c *Cache) cleanupExpiredRecords() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			expiredCount := 0
			for key, record := range c.data {
				if time.Now().After(record.Expiration) {
					delete(c.data, key)
					expiredCount++
				}
			}
			c.mu.Unlock()
			logger.Log(zap.InfoLevel, "Cache cleanup completed", zap.Int("expired_records", expiredCount))

		case <-c.ctx.Done(): // Exit when shutdown signal is received
			logger.Log(zap.InfoLevel, "Cache cleanup stopped due to shutdown")
			return
		}
	}
}
