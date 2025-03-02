package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	Server Server `json:"server"`
	Cache  Cache  `json:"cache"`
}
type Server struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}
type Cache struct {
	DefaultTTL int `json:"default_ttl"`
}

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
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}
	return data, nil
}
