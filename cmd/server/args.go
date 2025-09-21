package main

import (
	"flag"
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"github.com/alexflint/go-arg"
	"strconv"
)

func parseArgs() *config.ServerArg {
	var sArg config.ServerArg
	arg.MustParse(&sArg)
	return &sArg
}

func mustParseArgs() *config.ServerArg {
	var cfg config.ServerArg

	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address for endpoint")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.StringVar(&cfg.ConnStr, "d", "", "key")
	flag.Uint64Var(&cfg.Interval, "i", 300, "store interval")
	flag.StringVar(&cfg.File, "f", "metrics.json", "file name")
	flag.BoolVar(&cfg.Restore, "r", true, "load history")

	flag.Parse()

	cfg.Host = utils.LoadEnvVar("ADDRESS", cfg.Host, func(s string) (string, error) { return s, nil })
	cfg.Key = utils.LoadEnvVar("KEY", cfg.Key, func(s string) (string, error) { return s, nil })
	cfg.ConnStr = utils.LoadEnvVar("DATABASE_DSN", cfg.ConnStr, func(s string) (string, error) { return s, nil })
	cfg.File = utils.LoadEnvVar("FILE_STORAGE_PATH", cfg.File, func(s string) (string, error) { return s, nil })
	cfg.Interval = utils.LoadEnvVar("STORE_INTERVAL", cfg.Interval, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
	cfg.Restore = utils.LoadEnvVar("RESTORE", cfg.Restore, func(s string) (bool, error) { return strconv.ParseBool(s) })

	return &cfg
}
