package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	cfg := &Config{
		BlockFiles: map[string]string{"default": "/path/to/block/file.yaml"},
		Patterns: map[string]map[string]Pattern{
			"default": {
				"dev-gke-uswest": {
					CIDRSize:    26,
					Environment: "dev",
					Region:      "us-west1",
					Block:       "10.0.0.0/8",
				},
			},
		},
		ConfigFile: tmpfile.Name(),
	}

	// Test WriteConfig
	err = WriteConfig(cfg)
	assert.NoError(t, err)

	// Test LoadConfig
	loadedCfg, err := LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, cfg.BlockFiles["default"], loadedCfg.BlockFiles["default"])
	assert.Equal(t, cfg.Patterns["default"]["dev-gke-uswest"].CIDRSize, loadedCfg.Patterns["default"]["dev-gke-uswest"].CIDRSize)
}
