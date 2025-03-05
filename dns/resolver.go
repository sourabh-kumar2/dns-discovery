// Package dns provides a Resolver for handling DNS queries.
//
// The resolver parses incoming DNS queries, looks up responses in the cache,
// and constructs appropriate DNS responses to be sent back to clients.
package dns

import (
	"context"
	"fmt"

	"github.com/sourabh-kumar2/dns-discovery/discovery"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

// Resolver handles DNS query resolution using an in-memory cache.
//
// The resolver is responsible for parsing incoming queries, looking up
// answers in the cache, and constructing DNS response packets.
type Resolver struct {
	cache *discovery.Cache // In-memory cache for DNS records
}

// NewResolver initializes and returns a new Resolver instance.
//
// Parameters:
// - cache: The in-memory cache used for resolving DNS queries.
//
// Returns:
// - A pointer to the initialized Resolver instance.
func NewResolver(cache *discovery.Cache) *Resolver {
	return &Resolver{cache: cache}
}

// Resolve processes a raw DNS query and returns the corresponding response.
//
// It parses the query, checks the cache for matching records, and constructs a
// valid DNS response. If no matching records are found, an NXDOMAIN response is returned.
//
// Parameters:
// - ctx: The request context for logging and tracing.
// - query: The raw DNS query packet received from the client.
//
// Returns:
// - A byte slice containing the serialized DNS response packet.
// - An error if query parsing or response construction fails.
func (r *Resolver) Resolve(ctx context.Context, query []byte) ([]byte, error) {
	header, questions, err := ParseQuery(ctx, query)
	if err != nil {
		logger.Log(zap.WarnLevel, "Error parsing query", zap.Error(err))
		return nil, fmt.Errorf("error parsing query: %w", err)
	}

	ctx = logger.WithTransactionID(ctx, header.TransactionID)

	resp, err := BuildDNSResponse(ctx, questions, header, r.cache)
	if err != nil {
		logger.Log(zap.WarnLevel, "Error building DNS response", zap.Error(err))
		return nil, fmt.Errorf("error building DNS response: %w", err)
	}

	return resp, nil
}
