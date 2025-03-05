package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	validQTypes = map[uint16]bool{
		1:  true, // A
		2:  true, // NS
		5:  true, // CNAME
		6:  true, // SOA
		15: true, // MX
		16: true, // TXT
		28: true, // AAAA
	}

	validQClasses = map[uint16]bool{
		1: true, // IN (Internet)
		3: true, // CH (Chaosnet)
		4: true, // HS (Hesiod)
	}
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

// ParseQuestion parses a DNS question section, handling compression if present.
//
// Parameters:
// - data: The raw DNS packet.
// - offset: The starting offset of the question section.
//
// Returns:
// - A pointer to the parsed Question struct.
// - The new offset after parsing.
// - An error if parsing fails.
func ParseQuestion(data []byte, offset uint16) (*Question, uint16, error) {
	maxLen := uint16(len(data))

	domainName, newOffset, err := decodeDomainName(data, offset)
	if err != nil {
		return nil, 0, err
	}

	offset = newOffset // Update offset after reading domain name

	// Ensure enough bytes remain for QType and QClass
	if offset+4 > maxLen {
		return nil, 0, errors.New("incomplete question section")
	}

	qType := binary.BigEndian.Uint16(data[offset : offset+2])
	if !validQTypes[qType] {
		return nil, 0, errors.New("invalid QType")
	}
	qClass := binary.BigEndian.Uint16(data[offset+2 : offset+4])
	if !validQClasses[qClass] {
		return nil, 0, errors.New("invalid QClass")
	}
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
func decodeDomainName(data []byte, offset uint16) ([]byte, uint16, error) {
	// Ensure offset is within bounds.
	if offset >= uint16(len(data)) {
		return nil, 0, fmt.Errorf("offset out of range")
	}

	// If the domain name is the root domain, it is represented by a single 0 byte.
	if data[offset] == 0x00 {
		return []byte{}, offset + 1, nil
	}

	returnOffset := offset               // We will update this only on the first pointer jump.
	var domain []byte                    // Holds the assembled domain name.
	visited := make(map[uint16]struct{}) // Tracks visited offsets to prevent loops.
	jumped := false                      // Indicates if a pointer jump has occurred.

	for {
		// Check bounds.
		if offset >= uint16(len(data)) {
			return nil, 0, fmt.Errorf("offset out of range")
		}

		// Detect infinite loop.
		if _, ok := visited[offset]; ok {
			return nil, 0, fmt.Errorf("detected pointer loop at offset %d", offset)
		}
		visited[offset] = struct{}{}

		length := uint16(data[offset])
		// End of domain name.
		if length == 0 {
			offset++ // Move past the null terminator.
			break
		}

		// Check if this is a pointer (first two bits are set).
		if length&0xC0 == 0xC0 {
			// A pointer uses two bytes.
			if offset+1 >= uint16(len(data)) {
				return nil, 0, fmt.Errorf("invalid pointer at offset %d", offset)
			}
			// Extract pointer offset (mask with 0x3FFF to remove the two high bits).
			pointer := binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF
			if pointer >= uint16(len(data)) {
				return nil, 0, fmt.Errorf("pointer out of range")
			}
			// If this is the first pointer jump, record the return offset.
			if !jumped {
				returnOffset = offset + 2
			}
			jumped = true
			// Jump to the pointer offset.
			offset = pointer
			continue
		}

		// For a normal label, move past the length byte.
		offset++
		// Validate that the label fits in the data.
		if offset+length > uint16(len(data)) {
			return nil, 0, fmt.Errorf("invalid label length at offset %d", offset-1)
		}
		// Append the label to our domain.
		domain = append(domain, data[offset:offset+length]...)
		offset += length

		// Append a dot if the next byte is not the end of the domain or a pointer.
		if offset < uint16(len(data)) && data[offset] != 0 && (data[offset]&0xC0) != 0xC0 {
			domain = append(domain, '.')
		}
	}

	// Return the domain and the correct offset:
	// If we jumped via a pointer, use returnOffset; otherwise, use the current offset.
	if jumped {
		return domain, returnOffset, nil
	}
	return domain, offset, nil
}
