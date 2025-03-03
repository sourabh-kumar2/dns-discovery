package dns

import (
	"encoding/binary"
	"fmt"
)

// Question represents a DNS question section.
//
// The question section specifies the domain name being queried,
// along with the record type and class.
//
// Fields:
//   - DomainName: The domain name being queried, e.g., "example.com".
//   - QType: The type of DNS record being requested (e.g., A, AAAA, MX, TXT).
//   - QClass: The class of the query, typically 1 (IN for Internet).
type Question struct {
	DomainName string // The domain name being queried
	QType      uint16 // The type of DNS record requested
	QClass     uint16 // The class of the query, typically IN (1)
}

// parseDNSQuestion parses a DNS question section, handling compression if present.
//
// Parameters:
// - data: The raw DNS packet.
// - offset: The starting offset of the question section.
//
// Returns:
// - A pointer to the parsed Question struct.
// - The new offset after parsing.
// - An error if parsing fails.
func parseDNSQuestion(data []byte, offset int) (*Question, int, error) {
	maxLen := len(data)

	domainName, newOffset, err := decodeDomainName(data, offset)
	if err != nil {
		return nil, 0, err
	}

	offset = newOffset // Update offset after reading domain name

	// Ensure enough bytes remain for QType and QClass
	if offset+4 > maxLen {
		return nil, 0, fmt.Errorf("incomplete question section")
	}

	qType := binary.BigEndian.Uint16(data[offset : offset+2])
	qClass := binary.BigEndian.Uint16(data[offset+2 : offset+4])
	offset += 4

	return &Question{
		DomainName: string(domainName),
		QType:      qType,
		QClass:     qClass,
	}, offset, nil
}

// decodeDomainName extracts a domain name from a DNS message.
//
// It handles both standard domain name encoding and DNS compression (pointers).
// If a pointer is encountered, it jumps to the referenced offset and continues decoding.
// To prevent infinite loops caused by malformed packets, it tracks visited offsets.
//
// Parameters:
//   - data: The raw DNS message as a byte slice.
//   - offset: The starting position of the domain name in the message.
//
// Returns:
//   - A byte slice representing the decoded domain name (without the trailing null byte).
//   - The next offset position after reading the domain name.
//   - An error if the domain name is malformed, contains an invalid pointer, or forms a pointer loop.
func decodeDomainName(data []byte, offset int) ([]byte, int, error) {
	var domainName []byte
	originalOffset := offset
	maxLen := len(data)
	jumped := false
	visited := make(map[int]bool) // Track visited offsets

	for offset < maxLen {
		// Detect infinite loop by checking visited offsets
		if visited[offset] {
			return nil, 0, fmt.Errorf("detected pointer loop at offset %d", offset)
		}
		visited[offset] = true

		length := int(data[offset])
		if length == 0 { // End of domain name
			offset++
			break
		}

		// Handle compression (pointer)
		if length&0xC0 == 0xC0 {
			if offset+1 >= maxLen {
				return nil, 0, fmt.Errorf("invalid pointer")
			}
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if pointer >= maxLen {
				return nil, 0, fmt.Errorf("pointer out of range")
			}

			// Check if we've already jumped
			if jumped {
				return nil, 0, fmt.Errorf("multiple jumps in compressed domain name")
			}

			offset = pointer
			jumped = true
			continue
		}

		// Check for valid length
		if offset+length > maxLen {
			return nil, 0, fmt.Errorf("invalid label length")
		}

		// Append label
		if len(domainName) > 0 {
			domainName = append(domainName, '.')
		}
		domainName = append(domainName, data[offset+1:offset+1+length]...)
		offset += length + 1
	}

	// If we followed a pointer, return the original offset for proper QType reading
	if jumped {
		return domainName, originalOffset + 2, nil
	}
	return domainName, offset, nil
}
