package config

import "time"

type RetryConfig struct {
	MaxAttempts int
	Delays      []time.Duration
}
