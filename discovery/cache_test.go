package discovery

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	cache := NewCache()

	tcs := []struct {
		name      string
		key       string
		qType     uint16
		value     []byte
		ttl       time.Duration
		waitTime  time.Duration
		expectNil bool
	}{
		{
			name:      "Store and Retrieve TXT record",
			key:       "example.com",
			qType:     16, // TXT
			value:     []byte("sample TXT response"),
			ttl:       5 * time.Second,
			waitTime:  0,
			expectNil: false,
		},
		{
			name:      "Store and Retrieve A record",
			key:       "example.com",
			qType:     1,                      // A
			value:     []byte{192, 168, 1, 1}, // Fake IP
			ttl:       10 * time.Second,
			waitTime:  0,
			expectNil: false,
		},
		{
			name:      "Record expires after TTL",
			key:       "expired.com",
			qType:     1, // A
			value:     []byte{127, 0, 0, 1},
			ttl:       2 * time.Second,
			waitTime:  3 * time.Second,
			expectNil: true,
		},
		{
			name:      "Non-existent record returns nil",
			key:       "missing.com",
			qType:     1, // A
			value:     nil,
			ttl:       0,
			waitTime:  0,
			expectNil: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.value != nil {
				cache.Set(tc.key, tc.qType, tc.value, tc.ttl)
			}

			time.Sleep(tc.waitTime)

			result, found := cache.Get(tc.key, tc.qType)

			if tc.expectNil {
				assert.False(t, found, "Expected record to be expired/missing")
				assert.Nil(t, result, "Expected nil response")
			} else {
				assert.True(t, found, "Expected record to exist")
				assert.Equal(t, tc.value, result, "Value mismatch")
			}
		})
	}
}
