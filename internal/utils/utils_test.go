package utils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestParseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected []string
	}{
		{
			name:     "valid gauge update",
			url:      "/update/gauge/SomeMetric/123.45",
			expected: []string{"/update/gauge/SomeMetric/123.45", "gauge", "SomeMetric", "123.45"},
		},
		{
			name:     "valid counter update",
			url:      "/update/counter/TotalHits/42",
			expected: []string{"/update/counter/TotalHits/42", "counter", "TotalHits", "42"},
		},
		{
			name:     "metric name with dots and dashes",
			url:      "/update/gauge/my.metric-name/100.0",
			expected: []string{"/update/gauge/my.metric-name/100.0", "gauge", "my.metric-name", "100.0"},
		},
		{
			name:     " wrong prefix",
			url:      "/api/update/gauge/m/1",
			expected: nil,
		},
		{
			name:     "no match",
			url:      "/update/gauge/m",
			expected: nil,
		},
		{
			name:     "empty value",
			url:      "/update/gauge/m/",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadEnvVar(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		envValue    string
		hasEnv      bool
		flagValue   int
		setter      func(string) (int, error)
		expected    int
		expectPanic bool
	}{
		{
			name:      "env exists and valid",
			key:       "TEST_INT",
			envValue:  "42",
			hasEnv:    true,
			flagValue: 10,
			setter:    func(s string) (int, error) { return atoi(s) },
			expected:  42,
		},
		{
			name:      "env not set - use flag value",
			key:       "TEST_NOT_SET",
			hasEnv:    false,
			flagValue: 99,
			setter:    func(s string) (int, error) { return atoi(s) },
			expected:  99,
		},
		{
			name:        "env exists but invalid format",
			key:         "TEST_INVALID",
			envValue:    "not-a-number",
			hasEnv:      true,
			flagValue:   0,
			setter:      func(s string) (int, error) { return atoi(s) },
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasEnv {
				require.NoError(t, os.Setenv(tt.key, tt.envValue))
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			if tt.expectPanic {
				assert.Panics(t, func() {
					_ = LoadEnvVar(tt.key, tt.flagValue, tt.setter)
				})
			} else {
				result := LoadEnvVar(tt.key, tt.flagValue, tt.setter)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Вспомогательная функция для парсинга int
func atoi(s string) (int, error) {
	var i int
	_, err := fmt.Sscan(s, &i)
	return i, err
}

// ———————————————————————————————
// Тесты для ToString
// ———————————————————————————————

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string pointer",
			input:    strPtr("hello"),
			expected: "hello",
		},
		{
			name:     "int pointer",
			input:    intPtr(42),
			expected: "42",
		},
		{
			name:     "nil pointer",
			input:    (*string)(nil),
			expected: "nil",
		},
		{
			name:     "float64 pointer",
			input:    func() *float64 { f := 3.14; return &f }(),
			expected: "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Так как ToString — generic, вызываем через рефлексию или явно
			// Но проще — написать обёртки или проверить вручную
			// Здесь используем assert напрямую для каждого типа
		})
	}

	// Явные вызовы (так как generic в Go не позволяет легко делать table-driven для разных типов)
	assert.Equal(t, "hello", ToString(strPtr("hello")))
	assert.Equal(t, "42", ToString(intPtr(42)))
	assert.Equal(t, "nil", ToString((*string)(nil)))
	assert.Equal(t, "3.14", ToString(func() *float64 { f := 3.14; return &f }()))
}
