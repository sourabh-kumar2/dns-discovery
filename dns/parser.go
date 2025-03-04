// Package dns provides functionality to parse and handle DNS queries.
// It extracts and logs the DNS header, questions, and prepares for response processing.
package dns

import (
	"context"
	"errors"
	"fmt"

	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

const (
	// headerLength The DNS header consists of 12 bytes and contains important fields fixed 12.
	headerLength = 12
)

// ParseQuery processes a raw DNS query packet.
// It extracts the DNS header and all question sections, logging relevant details.
func ParseQuery(ctx context.Context, data []byte) (*Header, []Question, error) {
	if len(data) < headerLength {
		logger.LogWithContext(ctx, zap.ErrorLevel, "Failed to parse DNS header",
			zap.String("reason", "packet too short"),
		)
		return nil, nil, errors.New("packet too short")
	}

	header, err := parseDNSHeader(data)
	if err != nil {
		logger.LogWithContext(ctx, zap.WarnLevel, "Failed to parse DNS header", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to parse DNS header: %w", err)
	}
	ctx = logger.WithTransactionID(ctx, header.TransactionID)
	logger.LogWithContext(ctx, zap.DebugLevel, "Parsed DNS header", zap.Any("header", header))

	offset := uint16(headerLength)
	var questions []Question

	for i := 0; i < int(header.QDCount); i++ {
		question, newOffset, err := parseDNSQuestion(data, offset)
		if err != nil {
			logger.LogWithContext(ctx, zap.WarnLevel, "Failed to parse DNS question", zap.Int("questionIndex", i+1), zap.Error(err))
			return nil, nil, fmt.Errorf("failed to parse DNS question: %w", err)
		}

		logger.LogWithContext(ctx, zap.DebugLevel, "Parsed DNS question", zap.Int("questionIndex", i+1), zap.Any("question", question))
		questions = append(questions, *question)
		offset = newOffset
	}

	logger.LogWithContext(ctx, zap.DebugLevel, "Successfully parsed DNS query", zap.Int("questionsCount", len(questions)))
	return header, questions, nil
}
