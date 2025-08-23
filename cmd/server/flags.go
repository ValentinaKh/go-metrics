package main

import (
	"flag"
)

func parseFlags() (host *string) {

	tHost := flag.String("a", "localhost:8080", "address for endpoint")

	flag.Parse()

	return tHost
}
