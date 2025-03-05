package main

import "flag"

type flags struct {
	address string
	port    int
	debug   bool
}

func parseFlags() *flags {
	f := &flags{}

	flag.StringVar(&f.address, "a", "127.0.0.1", "IP address to bind the DNS server")
	flag.IntVar(&f.port, "p", 8053, "Port number to listen on")
	flag.BoolVar(&f.debug, "debug", false, "Enable debug logging (set flag without value to enable)")

	flag.Parse()
	return f
}
