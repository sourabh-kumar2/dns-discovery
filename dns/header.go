package dns

import (
	"encoding/binary"
	"fmt"
)

// Header represents the DNS message header.
//
// The DNS header consists of 12 bytes and contains important fields
// that define the structure and behavior of a DNS query or response.
//
// Fields:
//   - TransactionID: A 16-bit identifier assigned by the client.
//   - Flags: A 16-bit field specifying query/response type, operation code, and flags.
//   - QDCount: The number of entries in the Question section.
//   - ANCount: The number of resource records in the Answer section.
//   - NSCount: The number of resource records in the Authority section.
//   - ARCount: The number of resource records in the Additional section.
type Header struct {
	TransactionID uint16 // Unique identifier for the DNS request/response
	Flags         uint16 // Flags indicating query type and response settings
	QDCount       uint16 // Number of questions in the query
	ANCount       uint16 // Number of answers in the response
	NSCount       uint16 // Number of authority records
	ARCount       uint16 // Number of additional records
}

// parseDNSHeader parses the DNS packet header from the given byte slice.
// It ensures the packet is at least 12 bytes long (DNS header size) and extracts the header fields.
//
// Validation:
// - The function checks that the provided data is at least 12 bytes long.
// - It ensures that the Query Count (QDCount) is greater than zero, as a valid query must have at least one question.
//
// Returns:
// - A pointer to the parsed Header struct if successful.
// - An error if the packet is too short or has an invalid QDCount.
func parseDNSHeader(data []byte) (*Header, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("invalid DNS packet, too short")
	}

	header := &Header{
		TransactionID: binary.BigEndian.Uint16(data[0:2]),
		Flags:         binary.BigEndian.Uint16(data[2:4]),
		QDCount:       binary.BigEndian.Uint16(data[4:6]),
		ANCount:       binary.BigEndian.Uint16(data[6:8]),
		NSCount:       binary.BigEndian.Uint16(data[8:10]),
		ARCount:       binary.BigEndian.Uint16(data[10:12]),
	}

	if header.QDCount == 0 {
		return nil, fmt.Errorf("invalid DNS packet, no questions present (QDCount = 0)")
	}

	// To protect from flooding of questions. DOS
	if header.QDCount >= 10 {
		return nil, fmt.Errorf("invalid DNS packet, QDCount > 10")
	}

	opcode := (header.Flags >> 11) & 0xF
	// Validate Opcode
	if opcode > 2 { // Only 0 (Standard Query), 1 (Inverse Query), 2 (Server Status) are valid
		return nil, fmt.Errorf("invalid DNS header, unsupported opcode: %d", opcode)
	}

	return header, nil
}
