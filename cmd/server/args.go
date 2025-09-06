package main

import (
	"github.com/ValentinaKh/go-metrics/internal/config"
	"github.com/alexflint/go-arg"
)

func parseArgs() *config.ServerArg {
	var sArg config.ServerArg
	arg.MustParse(&sArg)
	return &sArg
}
