package cmd

import (
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lugnut42/openipam/internal/ipam"

	"github.com/lugnut42/openipam/internal/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSubnetCommands(t *testing.T) {
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
		ipam.SetConfig(cfg)
		return nil
	}

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(blockCmd)
	rootCmd.AddCommand(subnetCmd)

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

	// Set the config file for subsequent commands
	cfgFile = configPath

	// Create the block first that we'll create subnets in
	debugLog("Creating prerequisite block")
	err = executeCommand("block", "add", "--cidr", "10.0.0.0/16", "--description", "Test Block", "--file", "default")
	if err != nil {
		log.Printf("ERROR: Failed to add block: %v", err)
		if cfg != nil {
			debugLog("Current config state: %+v", cfg)
		}
		t.Fatal(err)
	}
	debugLog("Block created successfully")

	t.Run("subnet create", func(t *testing.T) {
		debugLog("Creating subnet")
		err := executeCommand("subnet", "create",
			"--block", "10.0.0.0/16",
			"--cidr", "10.0.1.0/24",
			"--name", "Test Subnet",
			"--region", "us-east1")
		if err != nil {
			log.Printf("ERROR: Failed to create subnet: %v", err)
		}
		assert.NoError(t, err)
	})

	t.Run("subnet create invalid CIDR", func(t *testing.T) {
		debugLog("Testing invalid CIDR creation")
		err := executeCommand("subnet", "create",
			"--block", "10.0.0.0/16",
			"--cidr", "10.0.1.0/33",
			"--name", "Invalid CIDR Subnet",
			"--region", "us-east1")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "invalid CIDR"),
			"Expected error about invalid CIDR, got: %v", err)
	})

	t.Run("subnet create overlapping", func(t *testing.T) {
		debugLog("Testing overlapping subnet creation")
		err := executeCommand("subnet", "create",
			"--block", "10.0.0.0/16",
			"--cidr", "10.0.1.0/24",
			"--name", "Overlapping Subnet",
			"--region", "us-east1")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "overlaps"),
			"Expected error about overlapping subnets, got: %v", err)
	})

	t.Run("subnet create outside block", func(t *testing.T) {
		debugLog("Testing out-of-block subnet creation")
		err := executeCommand("subnet", "create",
			"--block", "10.0.0.0/16",
			"--cidr", "11.0.0.0/24",
			"--name", "Outside Block Subnet",
			"--region", "us-east1")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not within"),
			"Expected error about subnet not being within block, got: %v", err)
	})

	t.Run("subnet list", func(t *testing.T) {
		debugLog("Listing subnets")
		err := executeCommand("subnet", "list", "--block", "10.0.0.0/16")
		assert.NoError(t, err)
	})

	t.Run("subnet show", func(t *testing.T) {
		debugLog("Showing subnet details")
		err := executeCommand("subnet", "show", "--cidr", "10.0.1.0/24")
		assert.NoError(t, err)
	})

	t.Run("subnet show non-existent", func(t *testing.T) {
		debugLog("Testing show of non-existent subnet")
		err := executeCommand("subnet", "show", "--cidr", "192.168.1.0/24")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not found"),
			"Expected error about subnet not found, got: %v", err)
	})

	t.Run("subnet delete", func(t *testing.T) {
		debugLog("Deleting subnet")
		err := executeCommand("subnet", "delete", "--cidr", "10.0.1.0/24", "--force")
		assert.NoError(t, err)
	})

	t.Run("subnet delete non-existent", func(t *testing.T) {
		debugLog("Testing deletion of non-existent subnet")
		err := executeCommand("subnet", "delete", "--cidr", "192.168.1.0/24", "--force")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not found"),
			"Expected error about subnet not found, got: %v", err)
	})
}
