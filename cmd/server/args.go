package main

import (
	"flag"
	"os"
	"strconv"
	"time"
)

func parseArgs() (string, time.Duration, string, bool) {
	var host, file string
	var interval uint64
	var restore bool

	flag.StringVar(&host, "a", "localhost:8080", "address for endpoint")
	flag.Uint64Var(&interval, "i", 300, "store interval")
	flag.StringVar(&file, "f", "metrics.json", "file name")
	flag.BoolVar(&restore, "r", true, "load history")

	flag.Parse()

	if r, ok := os.LookupEnv("ADDRESS"); ok {
		host = r
	}
	if r, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		res, err := strconv.ParseUint(r, 10, 64)
		if err != nil {
			panic(err)
		}
		interval = res
	}
	if r, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		file = r
	}
	if r, ok := os.LookupEnv("RESTORE"); ok {
		b, err := strconv.ParseBool(r)
		if err != nil {
			panic(err)
		}
		restore = b
	}

	return host, time.Duration(interval) * time.Second, file, restore
}
