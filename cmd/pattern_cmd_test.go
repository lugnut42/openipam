package cmd

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/lugnut42/openipam/internal/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPatternCommands(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "ipam-config.yaml")
	blockPath := filepath.Join(tempDir, "ip-blocks.yaml")

	debugLog("Test setup - Config path: %s", configPath)
	debugLog("Test setup - Block path: %s", blockPath)

	// Create test command
	rootCmd := &cobra.Command{Use: "ipam"}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip configuration check for "config init" command
		if cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			debugLog("Skipping config check for config init command")
			return nil
		}

		debugLog("PreRun hook - Loading config from: %s", cfgFile)
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			log.Printf("ERROR: Failed to load config in PreRun: %v", err)
			return err
		}
		debugLog("PreRun hook - Loaded config: %+v", cfg)
		return nil
	}

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(patternCmd)

	// Helper function to execute commands
	executeCommand := func(args ...string) error {
		debugLog("Executing command: %v", args)
		rootCmd.SetArgs(args)
		return rootCmd.Execute()
	}

	// Initialize config first
	debugLog("Initializing configuration")
	err := executeCommand("config", "init", "--config", configPath, "--block-yaml-file", blockPath)
	assert.NoError(t, err)

	// Verify config file exists and is valid
	debugLog("Verifying config file was created")
	testCfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("ERROR: Failed to load config for verification: %v", err)
		t.Fatal(err)
	}
	assert.NotNil(t, testCfg)
	assert.Contains(t, testCfg.BlockFiles, "default")
	debugLog("Initial config verification passed: %+v", testCfg)

	// Set the config file for subsequent commands
	cfgFile = configPath

	// Create the block that the pattern will reference
	debugLog("Creating prerequisite block")
	err = executeCommand("block", "add", "--cidr", "10.0.0.0/8", "--description", "Test Block", "--file", "default")
	if err != nil {
		log.Printf("ERROR: Failed to add block: %v", err)
		if cfg != nil {
			debugLog("Current config state: %+v", cfg)
		}
		t.Fatal(err)
	}
	debugLog("Block created successfully")

	// Create pattern
	debugLog("Creating pattern")
	err = executeCommand("pattern", "create",
		"--name", "dev-gke-uswest",
		"--cidr-size", "26",
		"--environment", "dev",
		"--region", "us-west1",
		"--block", "10.0.0.0/8",
		"--file", "default")
	if err != nil {
		log.Printf("ERROR: Failed to create pattern: %v", err)
		t.Fatal(err)
	}
	debugLog("Pattern created successfully")

	// List patterns
	debugLog("Listing patterns")
	err = executeCommand("pattern", "list", "--file", "default")
	assert.NoError(t, err)

	// Show pattern
	debugLog("Showing pattern details")
	err = executeCommand("pattern", "show", "--name", "dev-gke-uswest", "--file", "default")
	assert.NoError(t, err)

	// Delete pattern
	debugLog("Deleting pattern")
	err = executeCommand("pattern", "delete", "--name", "dev-gke-uswest", "--file", "default")
	assert.NoError(t, err)
}
