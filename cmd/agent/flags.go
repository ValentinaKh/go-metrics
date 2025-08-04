package main

import (
	"flag"
	"time"
)

func parseFlags() (host string, reportInterval time.Duration, pollInterval time.Duration) {

	tHost := flag.String("a", "localhost:8080", "address for endpoint")
	tReportInterval := flag.Uint("r", 10, "reportInterval")
	tPollInterval := flag.Uint("p", 2, "pollInterval")

	flag.Parse()

	return *tHost, time.Duration(*tReportInterval) * time.Second, time.Duration(*tPollInterval)
}
