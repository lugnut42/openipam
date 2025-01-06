package config

import (
	"fmt"
	"os"

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
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	cfg.ConfigFile = configFile
	return &cfg, nil
}

func WriteConfig(cfg *Config) error {
	if cfg.ConfigFile == "" {
		return fmt.Errorf("config file path not set")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	err = os.WriteFile(cfg.ConfigFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
