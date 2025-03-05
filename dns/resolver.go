package dns

import (
	"context"
	"fmt"

	"github.com/sourabh-kumar2/dns-discovery/discovery"

	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

// Resolver type.
type Resolver struct {
	cache *discovery.Cache
}

// NewResolver creates.
func NewResolver(cache *discovery.Cache) *Resolver {
	return &Resolver{cache: cache}
}

// Resolve the dns query and returns the response
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
