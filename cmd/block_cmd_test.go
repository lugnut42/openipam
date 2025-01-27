package cmd

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/lugnut42/openipam/internal/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBlockCommands(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

	tempDir := t.TempDir()
	debugLog("Created temp directory at: %s", tempDir)

	configFilePath := filepath.Join(tempDir, "ipam-config.yaml")
	blockFilePath := filepath.Join(tempDir, "ip-blocks.yaml")
	debugLog("Config file path: %s", configFilePath)
	debugLog("Block file path: %s", blockFilePath)

	initConfig := func() {
		debugLog("Starting initConfig()")

		// Create a new config instance
		cfg = &config.Config{
			BlockFiles: map[string]string{"default": blockFilePath},
			ConfigFile: configFilePath,
		}

		rootCmd := &cobra.Command{Use: "ipam"}
		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
		rootCmd.AddCommand(configCmd)

		debugLog("About to execute config init command")
		initArgs := []string{"config", "init", "--config", configFilePath, "--block-yaml-file", blockFilePath}
		debugLog("Config init args: %v", initArgs)
		rootCmd.SetArgs(initArgs)

		err := rootCmd.Execute()
		if err != nil {
			log.Printf("ERROR: Failed to execute config init command: %v", err)
		}
		assert.NoError(t, err)

		// Create a block YAML file with valid initial structure
		debugLog("Writing initial block file content")
		initialBlockContent := `[]`
		err = os.WriteFile(blockFilePath, []byte(initialBlockContent), 0644)
		if err != nil {
			log.Printf("ERROR: Failed to write initial block file: %v", err)
		}
		assert.NoError(t, err)

		// Debug prints
		configContent, err := os.ReadFile(configFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to read config file: %v", err)
		} else {
			debugLog("Config file contents:\n%s", string(configContent))
		}
		assert.NoError(t, err)

		blockContent, err := os.ReadFile(blockFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to read block file: %v", err)
		} else {
			debugLog("Block file contents:\n%s", string(blockContent))
		}
		assert.NoError(t, err)

		debugLog("Completed initConfig()")
	}

	initConfig()

	executeCommand := func(args ...string) error {
		debugLog("Executing command with args: %v", args)
		rootCmd := &cobra.Command{Use: "ipam"}

		// Important: Set up the persistent pre-run hook
		rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			debugLog("PreRun hook - Loading config from: %s", configFilePath)
			var err error
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				log.Printf("ERROR: Failed to load config in PreRun: %v", err)
				return err
			}
			debugLog("PreRun hook - Loaded config: %+v", cfg)
			return nil
		}

		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFilePath, "Path to configuration file")
		rootCmd.AddCommand(blockCmd)
		rootCmd.SetArgs(args)

		err := rootCmd.Execute()
		if err != nil {
			log.Printf("ERROR: Command execution failed: %v", err)
		} else {
			debugLog("Command executed successfully")
		}
		return err
	}

	t.Run("block add", func(t *testing.T) {
		debugLog("Starting 'block add' test")

		// Add additional validation
		_, err := os.Stat(configFilePath)
		assert.NoError(t, err, "Config file should exist")

		_, err = os.Stat(blockFilePath)
		assert.NoError(t, err, "Block file should exist")

		configContent, err := os.ReadFile(configFilePath)
		assert.NoError(t, err)
		debugLog("Config file contents before block add:\n%s", string(configContent))

		err = executeCommand("block", "add", "--cidr", "10.0.0.0/16", "--description", "Test Block", "--file", "default")
		assert.NoError(t, err)

		// Verify block was added
		blockContent, err := os.ReadFile(blockFilePath)
		assert.NoError(t, err)
		debugLog("Block file contents after add:\n%s", string(blockContent))
	})

	// Rest of the tests...
}
