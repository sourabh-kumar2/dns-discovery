package main

import (
	"log"
	"net"
)

const (
	port  = 8053
	ipStr = "127.0.0.1"
)

func main() {
	udpAddr := net.UDPAddr{
		IP:   net.ParseIP(ipStr),
		Port: port,
	}
	conn, err := net.ListenUDP("udp", &udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, readErr := conn.ReadFromUDP(buf)
		if readErr != nil {
			log.Fatalln("read error:", readErr)
		}
		log.Printf("Read %d bytes from %s", n, addr)
		log.Println(string(buf[:n]))
	}
}
