package config

import (
	"errors"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	Delays      []time.Duration
}

func (r RetryConfig) Validate() error {
	if len(r.Delays) == 0 {
		return errors.New("не заданы интервалы повторов")
	}

	if r.MaxAttempts <= 0 {
		return errors.New("MaxAttempts должно быть больше 0")
	}

	if r.MaxAttempts != len(r.Delays) {
		return errors.New("количество интервалов и максимальное количество повторов не совпадает")
	}
	return nil
}
