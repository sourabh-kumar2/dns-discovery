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

	return header, nil
}
