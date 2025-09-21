package main

import (
	"flag"
	"github.com/ValentinaKh/go-metrics/internal/utils"
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

	host = utils.LoadEnvVar("ADDRESS", host, func(s string) (string, error) { return s, nil })
	reportInterval = utils.LoadEnvVar("REPORT_INTERVAL", reportInterval, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
	pollInterval = utils.LoadEnvVar("POLL_INTERVAL", pollInterval, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })

	return host, time.Duration(reportInterval) * time.Second, time.Duration(pollInterval) * time.Second
}
