// Package config provides functions for loading and parsing
// the application configuration.
package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// Config represents the entire application configuration. It contains the server
// configuration and caching settings.
type Config struct {
	Server Server `json:"server"` // Server contains UDP server configuration.
	Cache  Cache  `json:"cache"`  // Cache contains caching configuration.
}

// Server defines the configuration for the UDP server, including the server's
// address and port.
type Server struct {
	Address string `json:"address"` // Address is the IP address or hostname the server will listen on.
	Port    int    `json:"port"`    // Port is the UDP port the server will bind to.
}

// Cache contains caching settings for the application, including the default
// Time-To-Live (TTL) value for cached data.
type Cache struct {
	DefaultTTL int `json:"default_ttl"` // DefaultTTL specifies the default Time-To-Live for cached items.
}

// NewConfig creates a new Config by reading and parsing a JSON file from the specified
// file path. It returns the Config object or an error if loading or parsing the file fails.
//
// path: The path to the configuration file.
func NewConfig(path string) (*Config, error) {
	log.Printf("Loading configuration from %s", path)
	data, err := readJSONFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return &config, nil
}

func readJSONFromFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}
	return data, nil
}
