package main

import (
	"flag"
	"log"
)

type flags struct {
	address  string
	port     int
	debug    bool
	filename string // Path to the JSON file
	interval int    // Cache refresh interval (seconds)

}

func parseFlags() *flags {
	f := &flags{}

	flag.StringVar(&f.address, "address", "127.0.0.1", "IP address to bind the DNS server")
	flag.IntVar(&f.port, "port", 8053, "Port number to listen on")
	flag.BoolVar(&f.debug, "debug", false, "Enable debug logging (set flag without value to enable)")
	flag.StringVar(&f.filename, "filename", "records.json", "Path to DNS records JSON file")
	flag.IntVar(&f.interval, "interval", 30, "Cache refresh interval in seconds")

	flag.Parse()

	log.Printf(
		"\naddress: %s\nport: %d\ndebug: %t\nfilename: %s\ninterval: %d\n",
		f.address,
		f.port,
		f.debug,
		f.filename,
		f.interval,
	)

	return f
}
