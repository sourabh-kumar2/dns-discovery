package discovery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cache := NewTestCache()

	tcs := []struct {
		name           string
		key            string
		qType          uint16
		value          []byte
		ttl            time.Duration
		expectedRecord *Record
	}{
		{
			name:  "Store and Retrieve TXT record",
			key:   "example.com",
			qType: 16, // TXT
			value: []byte("sample TXT response"),
			ttl:   5 * time.Second,
			expectedRecord: &Record{
				Value: []byte("sample TXT response"),
				TTL:   5 * time.Second,
			},
		},
		{
			name:  "Store and Retrieve A record",
			key:   "example.com",
			qType: 1,                      // A
			value: []byte{192, 168, 1, 1}, // Fake IP
			ttl:   10 * time.Second,
			expectedRecord: &Record{
				Value: []byte{192, 168, 1, 1},
				TTL:   10 * time.Second,
			},
		},
		{
			name:           "Non-existent record returns nil",
			key:            "missing.com",
			qType:          1, // A
			value:          nil,
			expectedRecord: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.value != nil {
				cache.Set(tc.key, tc.qType, tc.value, tc.ttl)
			}

			assert.Equal(t, cache.Get(tc.key, tc.qType), tc.expectedRecord)
		})
	}
}
