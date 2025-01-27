package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Initialize the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		blockYAMLFile, _ := cmd.Flags().GetString("block-yaml-file")

		logger.Debug("Config init called with config=%s, block-yaml=%s", configFile, blockYAMLFile)

		if configFile == "" {
			return fmt.Errorf("configuration file path is required")
		}

		if blockYAMLFile == "" {
			return fmt.Errorf("block YAML file path is required")
		}

		// Create the configuration directory if it doesn't exist
		configDir := filepath.Dir(configFile)
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating configuration directory: %w", err)
		}

		// Check if configuration file already exists
		if _, err := os.Stat(configFile); err == nil {
			return fmt.Errorf("configuration file already exists at %s", configFile)
		}

		// Create an empty block YAML file if it doesn't exist
		if _, err := os.Stat(blockYAMLFile); os.IsNotExist(err) {
			err = os.WriteFile(blockYAMLFile, []byte("[]"), 0644)
			if err != nil {
				return fmt.Errorf("error creating block YAML file: %w", err)
			}
		}

		// Write the configuration file
		cfg = &config.Config{
			BlockFiles: map[string]string{"default": blockYAMLFile},
			ConfigFile: configFile,
			Patterns:   make(map[string]map[string]config.Pattern),
		}

		err = config.WriteConfig(cfg)
		if err != nil {
			return fmt.Errorf("error writing configuration file: %w", err)
		}

		// Set the config for use by other commands
		ipam.SetConfig(cfg)

		fmt.Printf("Configuration file created successfully at %s\n", configFile)
		fmt.Printf("Block YAML file created successfully at %s\n", blockYAMLFile)

		logger.Debug("Config initialization complete. Config=%+v", cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configInitCmd.Flags().StringP("config", "c", "", "Path to configuration file")
	configInitCmd.Flags().StringP("block-yaml-file", "b", "", "Path to block YAML file")
}
