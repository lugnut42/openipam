package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/lugnut42/openipam/internal/logger"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Initialize and manage the configuration file.`,
}

// validateBlockName checks if the block name is valid
func validateBlockName(name string) error {
	// Only allow alphanumeric characters, hyphens, and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9-_]+$", name)
	if err != nil {
		return fmt.Errorf("error validating block name: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid block name '%s': only alphanumeric characters, hyphens, and underscores are allowed", name)
	}
	return nil
}

var configInitCmd = &cobra.Command{
	Use:   "init [block-name]",
	Short: "Initialize configuration",
	Long:  `Initialize the configuration file with a named block file.`,
	Args:  cobra.ExactArgs(1), // Require the block name argument
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := os.Getenv("IPAM_CONFIG_PATH")
		if configDir == "" {
			return fmt.Errorf("IPAM_CONFIG_PATH environment variable is required")
		}

		blockName := args[0]
		if err := validateBlockName(blockName); err != nil {
			return err
		}

		logger.Debug("Config init called with config directory=%s, block name=%s", configDir, blockName)

		// Create the configuration directory if it doesn't exist
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating configuration directory: %w", err)
		}

		// Create blocks subdirectory
		blocksDir := filepath.Join(configDir, "blocks")
		err = os.MkdirAll(blocksDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating blocks directory: %w", err)
		}

		// Construct the config file path
		configFile := filepath.Join(configDir, "ipam-config.yaml")

		// Check if configuration file already exists
		if _, err := os.Stat(configFile); err == nil {
			return fmt.Errorf("configuration file already exists at %s", configFile)
		}

		// Create named block file in blocks directory
		blockFile := filepath.Join(blocksDir, fmt.Sprintf("%s.yaml", blockName))
		if _, err := os.Stat(blockFile); os.IsNotExist(err) {
			err = os.WriteFile(blockFile, []byte("[]"), 0644)
			if err != nil {
				return fmt.Errorf("error creating block file: %w", err)
			}
		}

		// Write the configuration file
		cfg = &config.Config{
			BlockFiles: map[string]string{blockName: blockFile},
			ConfigFile: configFile,
			Patterns:   make(map[string]map[string]config.Pattern),
		}

		err = config.WriteConfig(cfg)
		if err != nil {
			return fmt.Errorf("error writing configuration file: %w", err)
		}

		// Set the config for use by other commands
		ipam.SetConfig(cfg)

		fmt.Printf("Configuration initialized successfully:\n")
		fmt.Printf("  Config file: %s\n", configFile)
		fmt.Printf("  Block file: %s (%s)\n", blockFile, blockName)

		return nil
	},
}

var configAddBlockCmd = &cobra.Command{
	Use:   "add-block [block-name]",
	Short: "Add a new block file",
	Long:  `Add a new named block file to the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		blockName := args[0]
		if err := validateBlockName(blockName); err != nil {
			return err
		}

		configDir := os.Getenv("IPAM_CONFIG_PATH")
		if configDir == "" {
			return fmt.Errorf("IPAM_CONFIG_PATH environment variable is required")
		}

		// Ensure we're using the full path to the config file
		configFile := filepath.Join(configDir, "ipam-config.yaml")
		logger.Debug("Loading config from: %s", configFile)

		// Load existing config
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("error loading configuration from %s: %w", configFile, err)
		}

		// Check if block name already exists
		if _, exists := cfg.BlockFiles[blockName]; exists {
			return fmt.Errorf("block file with name '%s' already exists", blockName)
		}

		// Create the new block file
		blockFile := filepath.Join(configDir, "blocks", fmt.Sprintf("%s.yaml", blockName))
		if err := os.MkdirAll(filepath.Dir(blockFile), 0755); err != nil {
			return fmt.Errorf("error creating blocks directory: %w", err)
		}

		if _, err := os.Stat(blockFile); os.IsNotExist(err) {
			err = os.WriteFile(blockFile, []byte("[]"), 0644)
			if err != nil {
				return fmt.Errorf("error creating block file: %w", err)
			}
		}

		// Add to config and save
		cfg.BlockFiles[blockName] = blockFile
		err = config.WriteConfig(cfg)
		if err != nil {
			return fmt.Errorf("error updating configuration: %w", err)
		}

		fmt.Printf("Added new block file:\n")
		fmt.Printf("  Name: %s\n", blockName)
		fmt.Printf("  Path: %s\n", blockFile)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configAddBlockCmd)
}