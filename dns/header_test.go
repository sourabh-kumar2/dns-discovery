package dns

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDNSHeader(t *testing.T) {
	tcs := []struct {
		name       string
		hexInput   string
		expectErr  bool
		expectedID uint16
		expectedQD uint16
		expectedAN uint16
		expectedNS uint16
		expectedAR uint16
	}{
		{
			name:       "Valid DNS Header",
			hexInput:   "123400000001000200000001",
			expectErr:  false,
			expectedID: 0x1234,
			expectedQD: 1,
			expectedAN: 2,
			expectedNS: 0,
			expectedAR: 1,
		},
		{
			name:      "Invalid Header (Too Short)",
			hexInput:  "1234",
			expectErr: true,
		},
		{
			name:      "Invalid Header (QDCount = 0)",
			hexInput:  "ABCD000000000200000001",
			expectErr: true,
		},
		{
			name:      "Invalid Opcode (out of range)",
			hexInput:  "5674F0000001000200000001", // F = 1111 (Opcode 15, invalid)
			expectErr: true,
		},
		{
			name:      "Too many QDs",
			hexInput:  "123400001111000000000000",
			expectErr: true,
		},
		{
			name:       "Header with Zero Counts",
			hexInput:   "ABCD01000000000000000000",
			expectErr:  true, // Now fails since QDCount = 0
			expectedID: 0xABCD,
			expectedQD: 0,
			expectedAN: 0,
			expectedNS: 0,
			expectedAR: 0,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			data, err := hex.DecodeString(tc.hexInput)
			assert.NoError(t, err, "Failed to decode hex input")

			header, err := parseDNSHeader(data)

			if tc.expectErr {
				assert.Error(t, err, "Expected error but got nil")
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotNil(t, header, "Header should not be nil")
			assert.Equal(t, tc.expectedID, header.TransactionID, "TransactionID mismatch")
			assert.Equal(t, tc.expectedQD, header.QDCount, "QDCount mismatch")
			assert.Equal(t, tc.expectedAN, header.ANCount, "ANCount mismatch")
			assert.Equal(t, tc.expectedNS, header.NSCount, "NSCount mismatch")
			assert.Equal(t, tc.expectedAR, header.ARCount, "ARCount mismatch")
		})
	}
}
