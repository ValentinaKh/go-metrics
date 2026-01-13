package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryConfig_Validate_TableDriven(t *testing.T) {
	testCases := []struct {
		name        string
		config      RetryConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with 3 attempts",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			},
			expectError: false,
		},
		{
			name: "Delays is nil",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      nil,
			},
			expectError: true,
			errorMsg:    "не заданы интервалы повторов",
		},
		{
			name: "Delays is empty slice",
			config: RetryConfig{
				MaxAttempts: 3,
				Delays:      []time.Duration{},
			},
			expectError: true,
			errorMsg:    "не заданы интервалы повторов",
		},
		{
			name: "MaxAttempts is zero",
			config: RetryConfig{
				MaxAttempts: 0,
				Delays:      []time.Duration{100 * time.Millisecond},
			},
			expectError: true,
			errorMsg:    "MaxAttempts должно быть больше 0",
		},
		{
			name: "MaxAttempts is negative",
			config: RetryConfig{
				MaxAttempts: -1,
				Delays:      []time.Duration{100 * time.Millisecond},
			},
			expectError: true,
			errorMsg:    "MaxAttempts должно быть больше 0",
		},
		{
			name: "MaxAttempts less than Delays length",
			config: RetryConfig{
				MaxAttempts: 2,
				Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond},
			},
			expectError: true,
			errorMsg:    "количество интервалов и максимальное количество повторов не совпадает",
		},
		{
			name: "MaxAttempts greater than Delays length",
			config: RetryConfig{
				MaxAttempts: 4,
				Delays:      []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
			},
			expectError: true,
			errorMsg:    "количество интервалов и максимальное количество повторов не совпадает",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
