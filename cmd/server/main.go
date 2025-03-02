package main

import (
	"github.com/sourabh-kumar2/dns-discovery/config"
	"log"
	"net"
)

func main() {
	cfg := config.NewConfig()
	startUDPServer(&cfg.Server)
}

func startUDPServer(server *config.Server) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(server.Address),
		Port: server.Port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatalf("Error starting UDP server on %s:%d - %v", server.Address, server.Port, err)
	}
	defer conn.Close()

	log.Printf("UDP server started on %s:%d", server.Address, server.Port)

	handleIncomingMessages(conn)
}

func handleIncomingMessages(conn *net.UDPConn) {
	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP connection: %v", err)
			continue
		}

		log.Printf("Received %d bytes from %s: %v", n, addr.IP, buf[:n])
	}
}
