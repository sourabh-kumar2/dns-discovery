package dns

import (
	"testing"

	"github.com/sourabh-kumar2/dns-discovery/logger"
	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	err := logger.InitLogger()
	if err != nil {
		t.Fatal(err)
	}

	tcs := []struct {
		name       string
		data       []byte
		expectErr  bool
		expectLogs []string
	}{
		{
			name: "Valid DNS query - Single Question",
			data: []byte{
				0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
				0x00, 0x01, 0x00, 0x01,
			},
			expectErr:  false,
			expectLogs: []string{"Parsed DNS header", "Parsed DNS question", "Successfully parsed DNS query"},
		},
		{
			name:      "Invalid DNS query - Too Short",
			data:      []byte{0x12, 0x34}, // Too short to be valid
			expectErr: true,
			expectLogs: []string{
				"Failed to parse DNS header",
			},
		},
		{
			name: "Valid DNS query - Multiple Questions",
			data: []byte{
				0x56, 0x78, 0x01, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
				0x00, 0x01, 0x00, 0x01,
				0x03, 'w', 'w', 'w',
				0xC0, 0x0C, // Pointer to "example.com"
				0x00, 0x01, 0x00, 0x01,
			},
			expectErr:  false,
			expectLogs: []string{"Parsed DNS header", "Parsed DNS question", "Parsed DNS question", "Successfully parsed DNS query"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			logs := logger.CaptureLogs(func() {
				ParseQuery(tc.data)
			})

			// Verify expected log messages
			for _, expectedLog := range tc.expectLogs {
				found := false
				for _, log := range logs {
					if log.Message == expectedLog {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected log message not found: "+expectedLog)
			}
		})
	}
}
