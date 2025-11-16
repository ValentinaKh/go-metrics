package config

import (
	"flag"
	"strconv"

	"github.com/ValentinaKh/go-metrics/internal/utils"
)

type AgentArg struct {
	CommonArgs
	ReportInterval uint64
	PollInterval   uint64
	RateLimit      uint64
}

type ServerArg struct {
	CommonArgs
	Interval  uint64
	File      string
	Restore   bool
	ConnStr   string
	AuditFile string
	AuditURL  string
}

type CommonArgs struct {
	Host string
	Key  string
}

func registerCommonFlags(cfg *CommonArgs) {
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address for endpoint")
	flag.StringVar(&cfg.Key, "k", "", "key")
}

func getCommonEnvVars(cfg *CommonArgs) {
	cfg.Host = utils.LoadEnvVar("ADDRESS", cfg.Host, func(s string) (string, error) { return s, nil })
	cfg.Key = utils.LoadEnvVar("KEY", cfg.Key, func(s string) (string, error) { return s, nil })
}

func MustParseAgentArgs() *AgentArg {
	var cfg AgentArg

	registerCommonFlags(&cfg.CommonArgs)
	flag.Uint64Var(&cfg.ReportInterval, "r", 10, "reportInterval")
	flag.Uint64Var(&cfg.PollInterval, "p", 2, "pollInterval")
	flag.Uint64Var(&cfg.RateLimit, "l", 2, "rateLimit")

	flag.Parse()

	getCommonEnvVars(&cfg.CommonArgs)
	cfg.ReportInterval = utils.LoadEnvVar("REPORT_INTERVAL", cfg.ReportInterval, uintParser)
	cfg.PollInterval = utils.LoadEnvVar("POLL_INTERVAL", cfg.PollInterval, uintParser)
	cfg.RateLimit = utils.LoadEnvVar("RATE_LIMIT", cfg.RateLimit, uintParser)

	return &cfg
}

func MustParseServerArgs() *ServerArg {
	var cfg ServerArg

	registerCommonFlags(&cfg.CommonArgs)
	flag.StringVar(&cfg.ConnStr, "d", "", "key")
	flag.Uint64Var(&cfg.Interval, "i", 300, "store interval")
	flag.StringVar(&cfg.File, "f", "metrics.json", "file name")
	flag.StringVar(&cfg.AuditFile, "audit-file", "metrics1.json", "file name")
	flag.StringVar(&cfg.AuditURL, "audit-url", "http://localhost:8080", "url")
	flag.BoolVar(&cfg.Restore, "r", true, "load history")

	flag.Parse()

	getCommonEnvVars(&cfg.CommonArgs)
	cfg.ConnStr = utils.LoadEnvVar("DATABASE_DSN", cfg.ConnStr, strParser)
	cfg.File = utils.LoadEnvVar("FILE_STORAGE_PATH", cfg.File, strParser)
	cfg.AuditFile = utils.LoadEnvVar("AUDIT_FILE", cfg.AuditFile, strParser)
	cfg.AuditURL = utils.LoadEnvVar("AUDIT_URL", cfg.AuditURL, strParser)
	cfg.Interval = utils.LoadEnvVar("STORE_INTERVAL", cfg.Interval, uintParser)
	cfg.Restore = utils.LoadEnvVar("RESTORE", cfg.Restore, boolParser)

	return &cfg
}

func strParser(s string) (string, error)  { return s, nil }
func uintParser(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) }
func boolParser(s string) (bool, error)   { return strconv.ParseBool(s) }
