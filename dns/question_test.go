package dns

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDNSQuestion(t *testing.T) {
	tcs := []struct {
		name      string
		data      []byte
		offset    uint16
		expected  *Question
		expectErr bool
	}{
		{
			name: "Valid standard question - example.com A record",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e', // "example"
				0x03, 'c', 'o', 'm', // "com"
				0x00,       // Null terminator
				0x00, 0x01, // Type A
				0x00, 0x01, // Class IN
			},
			offset: 0,
			expected: &Question{
				DomainName: "example.com",
				QType:      1,
				QClass:     1,
			},
			expectErr: false,
		},
		{
			name: "Valid compressed question - example.com A record",
			data: []byte{
				// Full domain at offset 12:
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,

				// Compressed reference to offset 12 (0x0C)
				0xC0, 0x00, // Pointer to offset 12
				0x00, 0x01, // Type A
				0x00, 0x01, // Class IN
			},
			offset: 13,
			expected: &Question{
				DomainName: "example.com",
				QType:      1,
				QClass:     1,
			},
			expectErr: false,
		},
		{
			name: "Invalid compressed question - pointer loop",
			data: []byte{
				0xC0, 0x00, // Pointer loops to itself
				0x00, 0x01,
				0x00, 0x01,
			},
			expectErr: true,
		},
		{
			name: "Invalid compressed question - truncated pointer",
			data: []byte{
				0xC0, // Incomplete pointer (missing offset byte)
			},
			offset:    0,
			expectErr: true,
		},
		{
			name: "Root Domain Name",
			data: []byte{0x00, 0x00, 0x01, 0x00, 0x01},
			expected: &Question{
				DomainName: "",
				QType:      1,
				QClass:     1,
			},
		},
		{
			name: "Label Too Long",
			data: []byte{
				0x4F, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
				0x01, 'c', 0x00, 0x00, 0x01, 0x00, 0x01,
			},
			expectErr: true,
		},
		{
			name: "Domain Name Too Long",
			data: func() []byte {
				d := make([]byte, 256)
				d[0] = 0x3F
				d[64] = 0x3F
				d[128] = 0x3F
				d[192] = 0x3F
				return d
			}(),
			expectErr: true,
		},
		{
			name: "Invalid QType",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
				0xFF, 0xFF,
				0x00, 0x01,
			},
			expectErr: true,
		},
		{
			name: "Invalid QClass",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
				0x00, 0x01,
				0x01, 0xFF,
			},
			expectErr: true,
		},
		{
			name: "Malformed Question",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
			},
			expectErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			q, _, err := parseDNSQuestion(tc.data, tc.offset)
			if (err != nil) != tc.expectErr {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				assert.Equal(t, tc.expected, q)
			}
		})
	}
}

func TestDecodeDomainName(t *testing.T) {
	tcs := []struct {
		name      string
		data      []byte
		offset    uint16
		expectErr bool
		expected  string
	}{
		{
			name: "Valid standard question - example.com A record",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
			},
			offset:   0,
			expected: "example.com",
		},
		{
			name: "Valid compressed question - example.com A record",
			data: []byte{
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm', 0x00,
				0xC0, 0x00,
			},
			offset:   13,
			expected: "example.com",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			name, _, err := decodeDomainName(tc.data, tc.offset)
			if (err != nil) != tc.expectErr {
				t.Errorf("Unexpected error: %v", err)
			}
			if string(name) != tc.expected {
				t.Errorf("Unexpected domain name. Got %s, expected %s", name, tc.expected)
			}
		})
	}
}
