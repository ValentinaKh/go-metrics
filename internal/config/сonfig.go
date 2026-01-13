// Package config - конфигурация приложения
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ValentinaKh/go-metrics/internal/logger"
	"github.com/ValentinaKh/go-metrics/internal/utils"
	"go.uber.org/zap"
	"os"
	"strconv"
)

type Basic interface {
	string | bool | int | uint64
}

// AgentArg  - agent config
type AgentArg struct {
	CommonArgs
	ReportInterval uint64 `json:"report_interval"`
	PollInterval   uint64 `json:"poll_interval"`
	RateLimit      uint64
}

// ServerArg - server config
type ServerArg struct {
	CommonArgs
	Interval       uint64 `json:"store_interval"`
	File           string `json:"store_file"`
	Restore        bool   `json:"restore"`
	ConnStr        string `json:"database_dsn"`
	AuditFile      string
	AuditURL       string
	ProfilePort    string
	AuditQueueSize uint64
}

type CommonArgs struct {
	Host      string `json:"address"`
	Key       string `json:"key"`
	CryptoKey string `json:"crypto_key"`
}

func registerCommonFlags(cfg *CommonArgs) {
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "address for endpoint")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "")
}

func getCommonEnvVars(cfg *CommonArgs) {
	cfg.Host = utils.LoadEnvVar("ADDRESS", cfg.Host, func(s string) (string, error) { return s, nil })
	cfg.Key = utils.LoadEnvVar("KEY", cfg.Key, func(s string) (string, error) { return s, nil })
	cfg.CryptoKey = utils.LoadEnvVar("CRYPTO_KEY", cfg.CryptoKey, func(s string) (string, error) { return s, nil })
}

func MustParseAgentArgs() *AgentArg {
	var cfg AgentArg
	path := getConfigPath()
	if path != "" {
		cfg = loadConfigFile[AgentArg](path)
	}

	registerCommonFlags(&cfg.CommonArgs)
	flag.Uint64Var(&cfg.ReportInterval, "r", configOrDefault(cfg.ReportInterval, 10), "reportInterval")
	flag.Uint64Var(&cfg.PollInterval, "p", configOrDefault(cfg.PollInterval, 2), "pollInterval")
	flag.Uint64Var(&cfg.RateLimit, "l", configOrDefault(cfg.RateLimit, 2), "rateLimit")

	flag.Parse()

	getCommonEnvVars(&cfg.CommonArgs)
	cfg.ReportInterval = utils.LoadEnvVar("REPORT_INTERVAL", cfg.ReportInterval, uintParser)
	cfg.PollInterval = utils.LoadEnvVar("POLL_INTERVAL", cfg.PollInterval, uintParser)
	cfg.RateLimit = utils.LoadEnvVar("RATE_LIMIT", cfg.RateLimit, uintParser)

	return &cfg
}

func MustParseServerArgs() *ServerArg {
	var cfg ServerArg
	path := getConfigPath()
	if path != "" {
		cfg = loadConfigFile[ServerArg](path)
	}

	registerCommonFlags(&cfg.CommonArgs)
	flag.StringVar(&cfg.ConnStr, "d", cfg.ConnStr, "key")
	flag.Uint64Var(&cfg.Interval, "i", configOrDefault(cfg.Interval, 300), "store interval")
	flag.Uint64Var(&cfg.AuditQueueSize, "b", 300, "audit queue size")
	flag.StringVar(&cfg.File, "f", configOrDefault(cfg.File, "metrics.json"), "file name")
	flag.StringVar(&cfg.AuditFile, "audit-file", "audit.json", "file name")
	flag.StringVar(&cfg.AuditURL, "audit-url", "http://localhost:8080", "url")
	flag.BoolVar(&cfg.Restore, "r", configOrDefault(cfg.Restore, true), "load history")

	flag.Parse()

	getCommonEnvVars(&cfg.CommonArgs)
	cfg.ProfilePort = ":6060"
	cfg.ConnStr = utils.LoadEnvVar("DATABASE_DSN", cfg.ConnStr, strParser)
	cfg.File = utils.LoadEnvVar("FILE_STORAGE_PATH", cfg.File, strParser)
	cfg.AuditFile = utils.LoadEnvVar("AUDIT_FILE", cfg.AuditFile, strParser)
	cfg.AuditURL = utils.LoadEnvVar("AUDIT_URL", cfg.AuditURL, strParser)
	cfg.AuditURL = utils.LoadEnvVar("PROFILE_PORT", cfg.ProfilePort, strParser)
	cfg.Interval = utils.LoadEnvVar("STORE_INTERVAL", cfg.Interval, uintParser)
	cfg.Restore = utils.LoadEnvVar("RESTORE", cfg.Restore, boolParser)

	return &cfg
}

func strParser(s string) (string, error)  { return s, nil }
func uintParser(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) }
func boolParser(s string) (bool, error)   { return strconv.ParseBool(s) }

func getConfigPath() string {
	var path string
	for i, arg := range os.Args {
		if arg == "-c" || arg == "-config" {
			if i+1 < len(os.Args) {
				path = os.Args[i+1]
				break
			}
		}
	}
	return utils.LoadEnvVar("CONFIG", path, strParser)
}

func loadConfigFile[T any](path string) T {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("can't open config file: %w", err))
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logger.Log.Error("can't close config file", zap.Error(err))
		}
	}(f)

	decoder := json.NewDecoder(f)

	var cfg T
	if err := decoder.Decode(&cfg); err != nil {
		panic(fmt.Errorf("can't parse config file: %w", err))
	}
	return cfg
}

func configOrDefault[T Basic](configVal T, defaultVal T) T {
	var zero T
	if configVal != zero {
		return configVal
	}
	return defaultVal
}
