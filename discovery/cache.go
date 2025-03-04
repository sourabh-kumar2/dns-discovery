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
)

// Record represents a cached DNS response.
type Record struct {
	Value []byte
	TTL   time.Duration
}

// Cache stores DNS records with TTL support.
type Cache struct {
	mu   sync.RWMutex
	data map[string]Record
}

// NewCache initializes a new cache instance.
func NewCache() *Cache {
	cache := &Cache{
		data: make(map[string]Record),
	}
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
