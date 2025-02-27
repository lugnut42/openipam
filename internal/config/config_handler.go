package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BlockFiles map[string]string             `yaml:"block_files"`
	Patterns   map[string]map[string]Pattern `yaml:"patterns"`
	ConfigFile string                        `yaml:"-"`
}

type Pattern struct {
	CIDRSize    int    `yaml:"cidr_size"`
	Environment string `yaml:"environment"`
	Region      string `yaml:"region"`
	Block       string `yaml:"block"`
}

func LoadConfig(configFile string) (*Config, error) {
	// Use filepath.Clean to normalize the path and remove any relative path elements
	cleanPath := filepath.Clean(configFile)
	
	// Verify the file exists before trying to read it
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing config file: %w", err)
	}
	
	// Check that it's a regular file, not a directory or other special file
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", cleanPath)
	}
	
	// Read the file with controlled permissions (addresses G304: Potential file inclusion via variable)
	data, err := os.ReadFile(cleanPath) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	cfg.ConfigFile = cleanPath
	return &cfg, nil
}

func WriteConfig(cfg *Config) error {
	if cfg.ConfigFile == "" {
		return fmt.Errorf("config file path not set")
	}

	// Use filepath.Clean to normalize the path and remove any relative path elements
	cleanPath := filepath.Clean(cfg.ConfigFile)
	
	// Make sure the parent directory exists
	dir := filepath.Dir(cleanPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("parent directory does not exist: %s", dir)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	// Write the file with secure permissions (0600 - only owner can read/write)
	err = os.WriteFile(cleanPath, data, 0600) // #nosec G304
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
