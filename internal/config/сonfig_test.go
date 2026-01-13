package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

type TestConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func TestConfigOrDefault(t *testing.T) {
	assert.Equal(t, 42, configOrDefault(42, 0))
	assert.Equal(t, 41, configOrDefault(0, 41))
	assert.Equal(t, "test", configOrDefault("", "test"))
	assert.Equal(t, "test", configOrDefault("test", ""))
	assert.Equal(t, true, configOrDefault(true, false))
}

func TestLoadConfigFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configContent := `{"host": "localhost", "port": 8080}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg := loadConfigFile[TestConfig](configPath)

	assert.Equal(t, TestConfig{Host: "localhost", Port: 8080}, cfg)
}

func TestLoadConfigFile_FileNotFound(t *testing.T) {
	require.Panics(t, func() { loadConfigFile[TestConfig]("/test/path.json") })
}
