package config

type AgentArg struct {
	Host           string `arg:"-a,env:ADDRESS" help:"адрес сервера" default:"localhost:8080"`
	ReportInterval uint64 `arg:"-r,env:REPORT_INTERVAL" help:"reportInterval" default:"10"`
	PollInterval   uint64 `arg:"-p,env:POLL_INTERVAL" help:"pollInterval" default:"2"`
	Key            string `arg:"-k,env:KEY" help:"key"`
}
