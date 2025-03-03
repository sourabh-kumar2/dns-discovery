// Package dns provides functionality to parse and handle DNS queries.
// It extracts and logs the DNS header, questions, and prepares for response processing.
package dns

import (
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

// ParseQuery processes a raw DNS query packet.
// It extracts the DNS header and all question sections, logging relevant details.
func ParseQuery(data []byte) {
	if len(data) < 12 {
		logger.Logger.Error("Failed to parse DNS header",
			zap.String("reason", "packet too short"),
		)
		return
	}

	header, err := parseDNSHeader(data)
	if err != nil {
		logger.Logger.Error("Failed to parse DNS header", zap.Error(err))
		return
	}
	logger.Logger.Debug("Parsed DNS header", zap.Any("header", header))

	offset := 12
	var questions []Question

	for i := 0; i < int(header.QDCount); i++ {
		question, newOffset, err := parseDNSQuestion(data, offset)
		if err != nil {
			logger.Logger.Error("Failed to parse DNS question", zap.Int("questionIndex", i+1), zap.Error(err))
			return
		}

		logger.Logger.Debug("Parsed DNS question", zap.Int("questionIndex", i+1), zap.Any("question", question))
		questions = append(questions, *question)
		offset = newOffset
	}

	logger.Logger.Info("Successfully parsed DNS query", zap.Int("questionsCount", len(questions)))
}
