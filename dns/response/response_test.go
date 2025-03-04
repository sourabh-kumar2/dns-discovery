package response

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/sourabh-kumar2/dns-discovery/discovery"
	"github.com/sourabh-kumar2/dns-discovery/dns"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"github.com/stretchr/testify/assert"
)

func TestBuildDNSResponse(t *testing.T) {
	tcs := []struct {
		name       string
		questions  []dns.Question
		header     *dns.Header
		cacheSetup func(*discovery.Cache)
		expectErr  bool
		validate   func(t *testing.T, response []byte)
	}{
		{
			name: "Valid single-question response",
			questions: []dns.Question{
				{
					DomainName: "example.com",
					QType:      16, // TXT record
					QClass:     1,  // IN (Internet)
				},
			},
			header: &dns.Header{
				TransactionID: 0x1234,
				Flags:         0x0100, // QR = 0 (query)
				QDCount:       1,
				ANCount:       0, // Set dynamically in response
				NSCount:       0,
				ARCount:       0,
			},
			cacheSetup: func(c *discovery.Cache) {
				c.Set("example.com", 16, []byte("hello world"), 30*time.Second)
			},
			expectErr: false,
			validate: func(t *testing.T, response []byte) {
				assertValidDNSResponse(t, response, 1, 1)
			},
		},
		{
			name: "Valid multiple-question response",
			questions: []dns.Question{
				{
					DomainName: "example.com",
					QType:      16,
					QClass:     1,
				},
				{
					DomainName: "service.local",
					QType:      16,
					QClass:     1,
				},
			},
			header: &dns.Header{
				TransactionID: 0x5678,
				Flags:         0x0100,
				QDCount:       2,
				ANCount:       0,
				NSCount:       0,
				ARCount:       0,
			},
			cacheSetup: func(c *discovery.Cache) {
				c.Set("example.com", 16, []byte("example txt"), 30*time.Second)
				c.Set("service.local", 16, []byte("service txt"), 30*time.Second)
			},
			expectErr: false,
			validate: func(t *testing.T, response []byte) {
				assertValidDNSResponse(t, response, 2, 2)
			},
		},
		{
			name: "Cache miss (NXDOMAIN response)",
			questions: []dns.Question{
				{
					DomainName: "unknown.com",
					QType:      16,
					QClass:     1,
				},
			},
			header: &dns.Header{
				TransactionID: 0x9999,
				Flags:         0x0100,
				QDCount:       1,
				ANCount:       0,
				NSCount:       0,
				ARCount:       0,
			},
			cacheSetup: func(_ *discovery.Cache) {}, // No cache entry
			expectErr:  false,
			validate: func(t *testing.T, response []byte) {
				assertValidDNSResponse(t, response, 1, 0) // No answers (NXDOMAIN)
			},
		},
		{
			name:       "No questions provided",
			questions:  []dns.Question{},
			header:     &dns.Header{TransactionID: 0x0001, Flags: 0x0100, QDCount: 0},
			cacheSetup: func(_ *discovery.Cache) {},
			expectErr:  true,
			validate:   nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			logger.CaptureLogs(func() {
				cache := discovery.NewCache()
				tc.cacheSetup(cache)

				resp, err := BuildDNSResponse(context.Background(), tc.questions, tc.header, cache)

				if tc.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					tc.validate(t, resp)
				}
			})
		})
	}
}

// Utility function to validate the DNS response format.
func assertValidDNSResponse(t *testing.T, response []byte, expectedQDCount, expectedANCount int) {
	assert.GreaterOrEqual(t, len(response), 12, "Response should have at least 12 bytes for the header")

	qdCount := binary.BigEndian.Uint16(response[4:6])
	anCount := binary.BigEndian.Uint16(response[6:8])

	assert.Equal(t, uint16(expectedQDCount), qdCount, "Mismatch in QDCount")
	assert.Equal(t, uint16(expectedANCount), anCount, "Mismatch in ANCount")
}
