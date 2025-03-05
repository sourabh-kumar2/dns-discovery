// Package server blah.
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

// Server type.
type Server struct {
	conn     *net.UDPConn
	done     chan struct{}
	wg       sync.WaitGroup
	resolver *dns.Resolver
}

// NewServer instance.
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

// Start the server.
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

// Stop the server.
func (s *Server) Stop() {
	<-s.done

	s.wg.Wait()
	logger.Log(zap.InfoLevel, "Shutdown complete.")
}
