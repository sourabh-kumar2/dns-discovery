// Package main builds the binary.
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sourabh-kumar2/dns-discovery/config"
)

func main() {
	configPath := flag.String("config", "cmd/server/config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatal(err)
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
	log.Printf("Received signal %v. Shutting down...", sig)
	cancel() // signal goroutines to stop

	// Wait for all goroutines to finish, with a timeout if necessary.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown complete.")
	case <-time.After(10 * time.Second):
		log.Println("Timeout during shutdown; forcing exit.")
	}
}

func startUDPServer(ctx context.Context, server *config.Server) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(server.Address),
		Port: server.Port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Error starting UDP server on %s:%d - %v", server.Address, server.Port, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	log.Printf("UDP server started on %s:%d", server.Address, server.Port)

	handleIncomingMessages(ctx, conn)
}

func handleIncomingMessages(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping message handling.")
			return
		default:
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Error reading from UDP connection: %v", err)
				continue
			}

			log.Printf("Received %d bytes from %s", n, addr.IP)
			go processPacket(ctx, buf[:n])
		}
	}
}

func processPacket(_ context.Context, buf []byte) {
	log.Printf("Received %s bytes from UDP", buf)
}
