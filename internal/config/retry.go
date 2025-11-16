package config

import (
	"errors"
	"time"
)

// RetryConfig - конфигурация повторов
type RetryConfig struct {
	MaxAttempts int
	Delays      []time.Duration
}

// Validate - проверка конфигурации
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
