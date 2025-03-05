// Package main builds the binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sourabh-kumar2/dns-discovery/config"
	"github.com/sourabh-kumar2/dns-discovery/discovery"
	"github.com/sourabh-kumar2/dns-discovery/dns"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"github.com/sourabh-kumar2/dns-discovery/server"
	"go.uber.org/zap"
)

func init() {
	if err := logger.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logger.Log(zap.InfoLevel, "Initialized logger")
}

func main() {
	configPath := flag.String("config", "cmd/server/config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		logger.Log(zap.FatalLevel, "Failed to initialize config", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := discovery.NewCache()
	cache.Set("example.com", 1, []byte{127, 0, 0, 2}, 300*time.Second)
	cache.Set("example.com", 16, []byte("example text"), 300*time.Second)
	resolver := dns.NewResolver(cache)

	srv, err := server.NewServer(cfg.Server.Address, cfg.Server.Port, resolver)
	if err != nil {
		logger.Log(zap.FatalLevel, "Failed to initialize server", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go srv.Start(ctx)

	sig := <-sigChan
	logger.Log(zap.InfoLevel, fmt.Sprintf("Received signal %v. Shutting down...", sig))

	cancel()

	srv.Stop()
}
