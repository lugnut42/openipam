package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigCommands(t *testing.T) {
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "ipam-config.yaml")
	blockFilePath := filepath.Join(tempDir, "ip-blocks.yaml")

	executeCommand := func(args ...string) error {
		rootCmd := &cobra.Command{Use: "ipam"}
		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
		rootCmd.AddCommand(configCmd)
		rootCmd.SetArgs(args)
		return rootCmd.Execute()
	}

	t.Run("config init", func(t *testing.T) {
		err := executeCommand("config", "init", "--config", configFilePath, "--block-yaml-file", blockFilePath)
		assert.NoError(t, err)
		_, err = os.Stat(configFilePath)
		assert.NoError(t, err)
	})

	t.Run("config show", func(t *testing.T) {
		cfg := &config.Config{
			BlockFiles: map[string]string{"default": blockFilePath},
			ConfigFile: configFilePath,
			Patterns: map[string]map[string]config.Pattern{
				"default": {
					"dev-gke-uswest": {
						CIDRSize:    26,
						Environment: "dev",
						Region:      "us-west1",
						Block:       "10.0.0.0/8",
					},
				},
			},
		}
		err := config.WriteConfig(cfg)
		assert.NoError(t, err)

		err = executeCommand("config", "show", "--config", configFilePath)
		assert.NoError(t, err)
	})
}

func TestConfig(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	cfg := &config.Config{
		BlockFiles: map[string]string{"default": "/path/to/block/file.yaml"},
		Patterns: map[string]map[string]config.Pattern{
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
	err = config.WriteConfig(cfg)
	assert.NoError(t, err)

	// Test LoadConfig
	loadedCfg, err := config.LoadConfig(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, cfg.BlockFiles["default"], loadedCfg.BlockFiles["default"])
	assert.Equal(t, cfg.Patterns["default"]["dev-gke-uswest"].CIDRSize, loadedCfg.Patterns["default"]["dev-gke-uswest"].CIDRSize)
}
