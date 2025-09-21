package main

import (
	"flag"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"strconv"
)

func mustParseArgs() *config.AgentArg {
	var cfg config.AgentArg

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address for endpoint")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.Uint64Var(&cfg.ReportInterval, "r", 10, "reportInterval")
	flag.Uint64Var(&cfg.PollInterval, "p", 2, "pollInterval")

	flag.Parse()

	cfg.Host = utils.LoadEnvVar("ADDRESS", cfg.Host, func(s string) (string, error) { return s, nil })
	cfg.Key = utils.LoadEnvVar("KEY", cfg.Key, func(s string) (string, error) { return s, nil })
	cfg.ReportInterval = utils.LoadEnvVar("REPORT_INTERVAL", cfg.ReportInterval, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
	cfg.PollInterval = utils.LoadEnvVar("POLL_INTERVAL", cfg.PollInterval, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })

	return &cfg
}
