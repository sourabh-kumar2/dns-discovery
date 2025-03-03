package dns

import (
	"testing"
)

func TestParseDNSQuestion(t *testing.T) {
	tcs := []struct {
		name      string
		data      []byte
		offset    int
		expected  Question
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
			expected: Question{
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
			expected: Question{
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
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			q, _, err := parseDNSQuestion(tc.data, tc.offset)
			if (err != nil) != tc.expectErr {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				if q.DomainName != tc.expected.DomainName || q.QType != tc.expected.QType || q.QClass != tc.expected.QClass {
					t.Errorf("Parsed question mismatch. Got %+v, expected %+v", q, tc.expected)
				}
			}
		})
	}
}

func TestDecodeDomainName(t *testing.T) {
	tcs := []struct {
		name      string
		data      []byte
		offset    int
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
