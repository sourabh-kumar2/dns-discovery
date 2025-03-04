// Package main builds the binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sourabh-kumar2/dns-discovery/config"
	"github.com/sourabh-kumar2/dns-discovery/dns"
	"github.com/sourabh-kumar2/dns-discovery/logger"
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

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startUDPServer(ctx, &cfg.Server)
	}()

	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	logger.Log(zap.WarnLevel, fmt.Sprintf("Received signal %v. Shutting down...", sig))
	cancel() // signal goroutines to stop

	// Wait for all goroutines to finish, with a timeout if necessary.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log(zap.InfoLevel, "Shutdown complete.")
	case <-time.After(5 * time.Second):
		logger.Log(zap.InfoLevel, "Timeout during shutdown; forcing exit.")
	}
}

func startUDPServer(ctx context.Context, server *config.Server) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(server.Address),
		Port: server.Port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		logger.Log(zap.FatalLevel, "Error starting UDP server",
			zap.String("server", server.Address),
			zap.Int("port", server.Port),
			zap.Error(err),
		)
	}
	defer func() {
		_ = conn.Close()
	}()

	logger.Log(zap.InfoLevel, "UDP server started",
		zap.String("server", server.Address),
		zap.Int("port", server.Port),
	)

	handleIncomingMessages(ctx, conn)
}

func handleIncomingMessages(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			logger.Log(zap.WarnLevel, "Stopping message handling.")
			return
		default:
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				logger.Log(zap.ErrorLevel, "Error reading from UDP connection", zap.Error(err))
				continue
			}

			logger.Log(zap.InfoLevel, fmt.Sprintf("Received %d bytes from %s", n, addr.IP))
			go processPacket(ctx, buf[:n])
		}
	}
}

func processPacket(ctx context.Context, buf []byte) {
	logger.WithRequestID(ctx, uuid.NewString())
	dns.ParseQuery(logger.WithRequestID(ctx, uuid.NewString()), buf)
}
