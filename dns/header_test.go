package dns

import (
	"encoding/hex"
	"testing"
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
			hexInput:   "123481800001000200000001",
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
			name:       "Header with Zero Counts",
			hexInput:   "ABCD01000000000000000000",
			expectErr:  false,
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
			if err != nil {
				t.Fatalf("Failed to decode hex input: %v", err)
			}

			header, err := parseDNSHeader(data)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if header.TransactionID != tc.expectedID {
				t.Errorf("TransactionID: expected %x, got %x", tc.expectedID, header.TransactionID)
			}
			if header.QDCount != tc.expectedQD {
				t.Errorf("QDCount: expected %d, got %d", tc.expectedQD, header.QDCount)
			}
			if header.ANCount != tc.expectedAN {
				t.Errorf("ANCount: expected %d, got %d", tc.expectedAN, header.ANCount)
			}
			if header.NSCount != tc.expectedNS {
				t.Errorf("NSCount: expected %d, got %d", tc.expectedNS, header.NSCount)
			}
			if header.ARCount != tc.expectedAR {
				t.Errorf("ARCount: expected %d, got %d", tc.expectedAR, header.ARCount)
			}
		})
	}
}
