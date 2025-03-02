package config

type Config struct {
	Server Server `json:"server"`
	Cache  Cache  `json:"cache"`
}
type Server struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}
type Cache struct {
	DefaultTtl int `json:"default_ttl"`
}

func NewConfig() *Config {
	return &Config{
		Server: Server{
			Address: "127.0.0.1",
			Port:    8053,
		},
		Cache: Cache{
			DefaultTtl: 100,
		},
	}
}
