package dns

import (
	"context"
	"testing"

	"github.com/sourabh-kumar2/dns-discovery/discovery"
	"github.com/stretchr/testify/assert"
)

func TestNewResolver(t *testing.T) {
	cache := discovery.NewCache()
	resolver := NewResolver(cache)

	assert.NotNil(t, resolver, "Resolver instance should not be nil")
}

func TestResolverResolveValidQuery(t *testing.T) {
	cache := discovery.NewCache()
	resolver := NewResolver(cache)

	// Preload cache with a test record
	cache.Set("example.com", 1, []byte{127, 0, 0, 1}, 300)

	// Simulated valid DNS query for "example.com A"
	query := []byte{
		0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, 0x00, 0x01,
	}

	ctx := context.Background()
	resp, err := resolver.Resolve(ctx, query)

	assert.NoError(t, err, "Expected no error for valid query")
	assert.NotNil(t, resp, "Expected a response")
	assert.Greater(t, len(resp), 12, "Response should be longer than the header")
}

func TestResolverResolveInvalidQuery(t *testing.T) {
	cache := discovery.NewCache()
	resolver := NewResolver(cache)

	// Simulated invalid query (too short)
	query := []byte{0x12, 0x34}

	ctx := context.Background()
	resp, err := resolver.Resolve(ctx, query)

	assert.Error(t, err, "Expected error for invalid query")
	assert.Nil(t, resp, "Expected no response for invalid query")
}

func TestResolverResolveNXDOMAIN(t *testing.T) {
	cache := discovery.NewCache()
	resolver := NewResolver(cache)

	// Simulated valid DNS query for a non-existent domain
	query := []byte{
		0x12, 0x34, 0x01, 0x00, 0x00, 1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x07, 'n', 'o', 'n', 'e', 'x', 'i', 's', 0x03, 'c', 'o', 'm', 0x00,
		0x00, 0x01, 0x00, 0x01,
	}

	ctx := context.Background()
	resp, err := resolver.Resolve(ctx, query)

	assert.NoError(t, err, "NXDOMAIN should not return an error")
	assert.NotNil(t, resp, "Expected a response")
	assert.Greater(t, len(resp), 12, "Response should be longer than the header")
}
