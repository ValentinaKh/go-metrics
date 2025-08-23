package main

import (
	"flag"
	"os"
)

func parseArgs() string {
	var host string
	flag.StringVar(&host, "a", "localhost:8080", "address for endpoint")

	flag.Parse()

	if r, ok := os.LookupEnv("ADDRESS"); ok {
		host = r
	}
	return host
}
