package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/lugnut42/openipam/internal/logger"

	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config
var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "ipam",
	Short: "IP Address Management tool",
	Long: `IP Address Management tool for managing IP blocks and subnets.

To get started, you need to initialize the configuration file using one of the following methods:

1. Using the IPAM_CONFIG_PATH environment variable:
  export IPAM_CONFIG_PATH=/path/to/ipam-config.yaml
  ipam config init

2. Using the --config flag:
  ipam config init --config /path/to/ipam-config.yaml

You can then use the following commands to manage IP blocks and subnets:
  ipam block create --cidr <CIDR> --name <n>
  ipam subnet create --block <block CIDR> --cidr <CIDR> --name <n>`,
	Example: `  ipam config init --config /path/to/ipam-config.yaml
  ipam block create --cidr 10.0.0.0/16 --name main-datacenter
  ipam subnet create --block 10.0.0.0/16 --cidr 10.0.1.0/24 --name app-tier`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set debug mode in logger package
		logger.SetDebugMode(debugMode)

		logger.Debug("PersistentPreRunE called for command: %s", cmd.Name())
		logger.Debug("Current cfgFile value: %s", cfgFile)
		logger.Debug("Current cfg value: %+v", cfg)

		// Skip configuration check for "config init" command
		if cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			logger.Debug("Skipping config check for config init command")
			return nil
		}

		// Skip configuration check for "config add-block" command
		if cmd.Name() == "add-block" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			logger.Debug("Skipping config check for config add-block command")
			return nil
		}

		// Check for --config flag
		if cfgFile == "" {
			// Check for environment variable
			envConfigPath := os.Getenv("IPAM_CONFIG_PATH")
			logger.Debug("Environment IPAM_CONFIG_PATH: %s", envConfigPath)
			if envConfigPath == "" {
				return fmt.Errorf("no configuration file specified. Please set the IPAM_CONFIG_PATH environment variable or use the --config flag")
			}
			// Construct config file path from environment variable
			cfgFile = filepath.Join(envConfigPath, "ipam-config.yaml")
		}

		logger.Debug("Using config file: %s", cfgFile)

		// Check if the configuration file exists
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Printf("ERROR: Config file not found at %s", cfgFile)
			return fmt.Errorf("configuration file not found at %s", cfgFile)
		}

		// Load the configuration
		logger.Debug("Loading configuration from %s", cfgFile)
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			log.Printf("ERROR: Failed to load config: %v", err)
			return fmt.Errorf("error loading config file: %v", err)
		}
		logger.Debug("Loaded config: %+v", cfg)

		// Set the configuration in the ipam package
		logger.Debug("Setting config in ipam package")
		ipam.SetConfig(cfg)

		return nil
	},
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		log.Printf("ERROR: Command execution failed: %v", err)
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func init() {
	logger.Debug("Initializing root command")
	cfg = &config.Config{}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug logging")

	// Add a direct command to check block file integrity
	validateFilesCmd := &cobra.Command{
		Use:   "check-files [file-key]",
		Short: "Check configuration files for integrity",
		Long:  `Check block files and configuration for integrity and consistency.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				fileKey := args[0]
				results, err := ipam.ValidateBlockFile(cfg, fileKey)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error:", err)
					os.Exit(1)
				}

				fmt.Printf("=== Block File: %s ===\n", fileKey)
				if err := ipam.PrintValidationResults(results); err != nil {
					fmt.Fprintf(os.Stderr, "Error printing validation results: %v\n", err)
					os.Exit(1)
				}

				if results.ErrorCount > 0 {
					os.Exit(1)
				}
			} else {
				err := ipam.ValidateAllBlockFiles(cfg)
				if err != nil {
					os.Exit(1)
				}
			}
		},
	}

	validateFilesCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	rootCmd.AddCommand(validateFilesCmd)
	logger.Debug("Root command initialized with empty config: %+v", cfg)
}
