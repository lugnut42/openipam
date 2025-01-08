package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/username/openipam/internal/ipam"

	"github.com/username/openipam/internal/config"

	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config

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
		log.Printf("DEBUG: PersistentPreRunE called for command: %s", cmd.Name())
		log.Printf("DEBUG: Current cfgFile value: %s", cfgFile)
		log.Printf("DEBUG: Current cfg value: %+v", cfg)

		// Skip configuration check for "config init" command
		if cmd.Name() == "init" && cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			log.Printf("DEBUG: Skipping config check for config init command")
			return nil
		}

		// Check for --config flag
		if cfgFile == "" {
			// Check for environment variable
			envConfigPath := os.Getenv("IPAM_CONFIG_PATH")
			log.Printf("DEBUG: Environment IPAM_CONFIG_PATH: %s", envConfigPath)
			if envConfigPath == "" {
				return fmt.Errorf("no configuration file specified. Please set the IPAM_CONFIG_PATH environment variable or use the --config flag")
			}
			cfgFile = envConfigPath
		}

		log.Printf("DEBUG: Using config file: %s", cfgFile)

		// Check if the configuration file exists
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			log.Printf("ERROR: Config file not found at %s", cfgFile)
			return fmt.Errorf("configuration file not found at %s", cfgFile)
		}

		// Load the configuration
		log.Printf("DEBUG: Loading configuration from %s", cfgFile)
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			log.Printf("ERROR: Failed to load config: %v", err)
			return fmt.Errorf("error loading config file: %v", err)
		}
		log.Printf("DEBUG: Loaded config: %+v", cfg)

		// Set the configuration in the ipam package
		log.Printf("DEBUG: Setting config in ipam package")
		ipam.SetConfig(cfg)

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Printf("ERROR: Command execution failed: %v", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	log.Printf("DEBUG: Initializing root command")
	cfg = &config.Config{}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
	log.Printf("DEBUG: Root command initialized with empty config: %+v", cfg)
}
