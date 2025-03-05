package dns

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/sourabh-kumar2/dns-discovery/discovery"
	"github.com/sourabh-kumar2/dns-discovery/dns/internal"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

const (
	// QRResponse Query Response flag.
	QRResponse = 0x8000

	// RANotAvailable Recursion Available flag.
	RANotAvailable = 0x0080

	// NXDomain NXDOMAIN response code.
	NXDomain = 0x0003
)

// BuildDNSResponse constructs a DNS response packet based on the query and header.
//
// This function does the following:
// 1. Copies the DNS header from the query and modifies it to indicate a response.
// 2. Includes the question section as it is in the response.
// 3. (Future Implementation) Appends an answer section if a valid response is found.
//
// Parameters:
//   - query: The parsed DNS question containing the domain name, QType, and QClass.
//   - header: The parsed DNS header from the query.
//
// Returns:
//   - A byte slice representing the serialized DNS response packet.
//   - An error if serialization fails.
func BuildDNSResponse(ctx context.Context, questions []internal.Question, header *internal.Header, cache *discovery.Cache) ([]byte, error) {
	if len(questions) == 0 {
		logger.LogWithContext(ctx, zap.ErrorLevel, "No questions provided")
		return nil, errors.New("no questions provided")
	}

	header.Flags |= QRResponse | RANotAvailable
	header.ANCount = 0 // Will be updated dynamically
	header.ARCount = 0
	header.NSCount = 0

	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, header); err != nil {
		logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write DNS header", zap.Error(err))
		return nil, fmt.Errorf("failed to write DNS header: %w", err)
	}

	domainOffsets := make(map[string]int)
	for _, q := range questions {
		if err := encodeDomainName(&buf, q.DomainName, domainOffsets); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write Domain", zap.Error(err))
			return nil, fmt.Errorf("failed to write Domain: %w", err)
		}

		// Write QType and QClass.
		if err := binary.Write(&buf, binary.BigEndian, q.QType); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write QType", zap.Error(err))
			return nil, fmt.Errorf("failed to write QType: %w", err)
		}
		if err := binary.Write(&buf, binary.BigEndian, q.QClass); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write QClass", zap.Error(err))
			return nil, fmt.Errorf("failed to write QClass: %w", err)
		}
	}

	for _, q := range questions {
		record := cache.Get(q.DomainName, q.QType)
		if record == nil {
			logger.LogWithContext(ctx, zap.InfoLevel, "No record found for domain name: NXDOMAIN", zap.String("domain", q.DomainName))
			continue
		}
		logger.LogWithContext(ctx, zap.DebugLevel, "cache hit",
			zap.String("domain", q.DomainName),
			zap.Uint16("qtype", q.QType),
			zap.Any("record", record),
		)

		header.ANCount++

		if err := encodeDomainName(&buf, q.DomainName, domainOffsets); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write Domain", zap.Error(err))
			return nil, fmt.Errorf("failed to write Domain: %w", err)
		}
		if err := binary.Write(&buf, binary.BigEndian, q.QType); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write QType", zap.Error(err))
			return nil, fmt.Errorf("failed to write QType: %w", err)
		}
		if err := binary.Write(&buf, binary.BigEndian, q.QClass); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write QClass", zap.Error(err))
			return nil, fmt.Errorf("failed to write QClass: %w", err)
		}
		if err := binary.Write(&buf, binary.BigEndian, uint32(record.TTL)); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write TTL", zap.Error(err))
			return nil, fmt.Errorf("failed to write TTL: %w", err)
		}

		var rdataBuf bytes.Buffer
		if q.QType == 16 { // TXT Record
			txtData := record.Value
			if len(txtData) > 255 {
				return nil, fmt.Errorf("TXT record too long")
			}

			// TXT format requires a length byte before the actual string
			rdataBuf.WriteByte(byte(len(txtData)))
			rdataBuf.Write(txtData)
		} else {
			rdataBuf.Write(record.Value)
		}

		// Write RDLENGTH
		if err := binary.Write(&buf, binary.BigEndian, uint16(rdataBuf.Len())); err != nil {
			logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to write RDLENGTH", zap.Error(err))
			return nil, fmt.Errorf("failed to write RDLENGTH: %w", err)
		}

		buf.Write(rdataBuf.Bytes())
	}

	if header.ANCount == 0 {
		header.Flags |= NXDomain
	}

	// Update the ANCount in the header
	bufBytes := buf.Bytes()
	binary.BigEndian.PutUint16(bufBytes[2:], header.Flags)
	binary.BigEndian.PutUint16(bufBytes[6:], header.ANCount)

	logger.LogWithContext(ctx, zap.DebugLevel, "Successfully built DNS response",
		zap.String("raw response", fmt.Sprintf("%x", bufBytes)),
	)
	return bufBytes, nil
}

func encodeDomainName(buf *bytes.Buffer, domain string, domainOffsets map[string]int) error {
	if domain == "" {
		buf.WriteByte(0x00) // Root domain
	}

	if offset, ok := domainOffsets[domain]; ok {
		pointer := 0xC000 | offset
		if err := binary.Write(buf, binary.BigEndian, uint16(pointer)); err != nil {
			return fmt.Errorf("failed to write offset: %w", err)
		}
		return nil
	}

	currentOffset := buf.Len()
	domainOffsets[domain] = currentOffset

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) > 63 {
			return fmt.Errorf("label %q exceeds 63 characters", label)
		}

		if err := buf.WriteByte(byte(len(label))); err != nil {
			return fmt.Errorf("failed to write label length: %w", err)
		}
		if _, err := buf.WriteString(label); err != nil {
			return fmt.Errorf("failed to write label: %w", err)
		}
	}
	return buf.WriteByte(0x00)
}
