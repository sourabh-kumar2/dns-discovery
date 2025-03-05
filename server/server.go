// Package server implements a UDP-based DNS server.
//
// It listens for DNS queries, processes incoming packets, and sends responses.
// The server supports graceful shutdown and concurrent request handling.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sourabh-kumar2/dns-discovery/dns"
	"github.com/sourabh-kumar2/dns-discovery/logger"
	"go.uber.org/zap"
)

// Server represents a UDP-based DNS server.
//
// It listens for DNS queries, processes them using a resolver, and sends responses.
type Server struct {
	conn     *net.UDPConn   // UDP connection for handling requests
	done     chan struct{}  // Channel to signal server shutdown
	wg       sync.WaitGroup // WaitGroup to track active requests
	resolver *dns.Resolver  // Resolver to process incoming queries
}

// NewServer initializes and returns a new DNS server.
//
// Parameters:
// - addr: The IP address to bind the server to.
// - port: The UDP port to listen on.
// - resolver: The resolver responsible for handling DNS queries.
//
// Returns:
// - A pointer to the initialized Server instance.
// - An error if the server fails to start.
func NewServer(addr string, port int, resolver *dns.Resolver) (*Server, error) {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP(addr),
		Port: port,
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Log(zap.FatalLevel, "Error starting UDP server",
			zap.String("server", addr),
			zap.Int("port", port),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error starting UDP server: %w", err)
	}

	return &Server{
		conn:     conn,
		done:     make(chan struct{}),
		resolver: resolver,
	}, nil
}

// handleIncomingMessages continuously listens for incoming UDP packets and processes them.
//
// It spawns a new goroutine for each request to allow concurrent processing.
func (s *Server) handleIncomingMessages(ctx context.Context) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			logger.Log(zap.WarnLevel, "Stopping message handling.")
			return
		default:
			_ = s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}

				logger.Log(zap.ErrorLevel, "Error reading from UDP connection", zap.Error(err))
				continue
			}

			data := make([]byte, n)
			copy(data, buf[:n])

			logger.Log(zap.InfoLevel, fmt.Sprintf("Received %d bytes from %s", n, addr.IP))
			s.wg.Add(1)
			go s.processPacket(ctx, addr, data)
		}
	}
}

// Start begins listening for incoming DNS requests and processing them.
//
// This function should be called as a goroutine to allow for asynchronous operation.
func (s *Server) Start(ctx context.Context) {
	defer func() {
		close(s.done)
		_ = s.conn.Close()
	}()

	logger.LogWithContext(
		ctx, zap.InfoLevel, "Server started listening",
		zap.Any("address", s.conn.LocalAddr().String()),
	)

	s.handleIncomingMessages(ctx)
}

// processPacket handles a single DNS query from a client.
//
// It parses the query, resolves it using the configured resolver, and sends a response.
//
// Parameters:
// - ctx: The request context.
// - addr: The address of the client sending the request.
// - buf: The raw DNS query data.
func (s *Server) processPacket(ctx context.Context, addr *net.UDPAddr, buf []byte) {
	defer s.wg.Done()

	ctx = logger.WithRequestID(ctx, uuid.NewString())

	resp, err := s.resolver.Resolve(ctx, buf)
	if err != nil {
		logger.Log(zap.WarnLevel, "Error building DNS response", zap.Error(err))
		return
	}

	_, err = s.conn.WriteToUDP(resp, addr)
	if err != nil {
		logger.LogWithContext(ctx, zap.ErrorLevel, "Error writing DNS response", zap.Error(err))
		return
	}
	logger.LogWithContext(ctx, zap.InfoLevel, "DNS response written to UDP")
}

// Stop gracefully shuts down the server.
//
// It waits for all active request handlers to finish before terminating.
func (s *Server) Stop() {
	<-s.done

	s.wg.Wait()
	logger.Log(zap.InfoLevel, "Shutdown complete.")
}
