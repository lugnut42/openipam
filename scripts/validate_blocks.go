package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/ipam"
)

func main() {
	// Set up the config file path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(homeDir, ".openipam", "ipam-config.yaml")

	// Override with IPAM_CONFIG_PATH environment variable if set
	if envPath := os.Getenv("IPAM_CONFIG_PATH"); envPath != "" {
		configPath = filepath.Join(envPath, "ipam-config.yaml")
	}

	// Load the configuration
	fmt.Printf("Loading config from: %s\n", configPath)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set the config in the ipam package
	ipam.SetConfig(cfg)

	// Process command line arguments
	args := os.Args[1:]
	validateAll := false
	fileKey := "default"

	// Simple command line argument parsing
	if len(args) > 0 {
		if args[0] == "--all" || args[0] == "-a" {
			validateAll = true
		} else {
			fileKey = args[0]
		}
	}

	if validateAll {
		fmt.Println("Validating all block files...")
		if err := ipam.ValidateAllBlockFiles(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Validating block file '%s'...\n", fileKey)
		results, err := ipam.ValidateBlockFile(cfg, fileKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := ipam.PrintValidationResults(results); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing validation results: %v\n", err)
			os.Exit(1)
		}
		if results.ErrorCount > 0 {
			os.Exit(1)
		}
	}
}
