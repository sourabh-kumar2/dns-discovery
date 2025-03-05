package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

type fileRecord struct {
	Domain string `json:"domain"` // Fully qualified domain name
	QType  uint16 `json:"qtype"`  // DNS record type (e.g., A = 1, TXT = 16)
	Value  string `json:"value"`  // Record value (IP address, TXT data, etc.)
	TTL    int    `json:"ttl"`    // Time-to-live in seconds
}

func loadFromFile(filename string) (map[string]Record, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var records []fileRecord
	if err := json.Unmarshal(file, &records); err != nil {
		return nil, fmt.Errorf("failed to parse JSON records: %w", err)
	}

	recordMap := make(map[string]Record)
	for _, rec := range records {
		if rec.Domain == "" || rec.QType == 0 || rec.TTL <= 0 {
			logger.Log(zap.WarnLevel, "Skipping invalid record", zap.Any("record", rec))
			continue
		}
		value, pErr := parseValue(rec.QType, rec.Value)
		if pErr != nil {
			logger.Log(zap.WarnLevel, "Skipping invalid record", zap.Any("record", rec))
			continue
		}
		if rec.QType == 1 {
			value = net.ParseIP(rec.Value).To4()
		}
		key := formatKey(rec.Domain, rec.QType)
		recordMap[key] = Record{
			Value: value,
			TTL:   time.Duration(rec.TTL),
		}
	}

	logger.Log(zap.InfoLevel, "Loaded DNS records from file", zap.Int("count", len(recordMap)))
	return recordMap, nil
}

func parseValue(qType uint16, value string) ([]byte, error) {
	switch qType {
	case 1:
		ip := net.ParseIP(value).To4()
		if ip == nil {
			return nil, fmt.Errorf("invalid IPv4 address: %s", value)
		}
		return ip, nil
	case 16:
		return []byte(value), nil
	case 28:
		ip := net.ParseIP(value).To16()
		if ip == nil {
			return nil, fmt.Errorf("invalid IPv6 address: %s", value)
		}
		return ip, nil
	default:
		return nil, fmt.Errorf("invalid qtype %d", qType)
	}
}
