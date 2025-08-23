package main

import (
	"flag"
	"os"
	"strconv"
	"time"
)

func mustParseArgs() (string, time.Duration, time.Duration) {
	var host string
	var reportInterval, pollInterval uint64

	flag.StringVar(&host, "a", "localhost:8080", "address for endpoint")
	flag.Uint64Var(&reportInterval, "r", 10, "reportInterval")
	flag.Uint64Var(&pollInterval, "p", 2, "pollInterval")

	flag.Parse()

	if r, ok := os.LookupEnv("ADDRESS"); ok {
		host = r
	}

	if r, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		res, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			panic(err)
		}
		reportInterval = res
	}

	if r, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		res, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			panic(err)
		}
		pollInterval = res
	}

	return host, time.Duration(reportInterval) * time.Second, time.Duration(pollInterval)
}
