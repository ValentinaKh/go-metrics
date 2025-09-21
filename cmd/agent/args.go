package main

import (
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/alexflint/go-arg"
)

func parseArgs() *config.AgentArg {
	var sArg config.AgentArg
	arg.MustParse(&sArg)
	return &sArg
}
