package ipam

import (
	"fmt"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
)

func CreatePattern(cfg *config.Config, name string, cidrSize int, environment, region, block, fileKey string) error {
	logger.Debug("Creating pattern: %s", name)
	if cfg.Patterns == nil {
		cfg.Patterns = make(map[string]map[string]config.Pattern)
	}
	patterns, ok := cfg.Patterns[fileKey]
	if !ok {
		patterns = make(map[string]config.Pattern)
		cfg.Patterns[fileKey] = patterns
	}

	if _, exists := patterns[name]; exists {
		return fmt.Errorf("pattern %s already exists", name)
	}

	// Validate CIDR size
	if cidrSize < 0 || cidrSize > 32 {
		return fmt.Errorf("invalid CIDR size: %d", cidrSize)
	}

	// Ensure the block exists
	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return fmt.Errorf("block file for key %s not found", fileKey)
	}

	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		return fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	blockExists := false
	for _, b := range blocks {
		if b.CIDR == block {
			blockExists = true
			break
		}
	}

	if !blockExists {
		return fmt.Errorf("block %s not found", block)
	}

	pattern := config.Pattern{
		CIDRSize:    cidrSize,
		Environment: environment,
		Region:      region,
		Block:       block,
	}

	patterns[name] = pattern
	logger.Debug("Pattern created: %+v", pattern)
	return config.WriteConfig(cfg)
}

func ListPatterns(cfg *config.Config, fileKey string) error {
	logger.Debug("Listing patterns for file key: %s", fileKey)
	patterns, ok := cfg.Patterns[fileKey]
	if !ok {
		return fmt.Errorf("no patterns found for file key %s", fileKey)
	}

	for name, pattern := range patterns {
		fmt.Printf("Name: %s, CIDR Size: %d, Environment: %s, Region: %s, Block: %s\n",
			name, pattern.CIDRSize, pattern.Environment, pattern.Region, pattern.Block)
	}
	return nil
}

func ShowPattern(cfg *config.Config, name, fileKey string) error {
	logger.Debug("Showing pattern: %s", name)
	patterns, ok := cfg.Patterns[fileKey]
	if !ok {
		return fmt.Errorf("no patterns found for file key %s", fileKey)
	}

	pattern, ok := patterns[name]
	if !ok {
		return fmt.Errorf("pattern %s not found", name)
	}

	fmt.Printf("Name: %s, CIDR Size: %d, Environment: %s, Region: %s, Block: %s\n",
		name, pattern.CIDRSize, pattern.Environment, pattern.Region, pattern.Block)
	return nil
}

func DeletePattern(cfg *config.Config, name, fileKey string) error {
	logger.Debug("Deleting pattern: %s", name)
	patterns, ok := cfg.Patterns[fileKey]
	if !ok {
		return fmt.Errorf("no patterns found for file key %s", fileKey)
	}

	if _, exists := patterns[name]; !exists {
		return fmt.Errorf("pattern %s not found", name)
	}

	delete(patterns, name)
	logger.Debug("Pattern deleted: %s", name)
	return config.WriteConfig(cfg)
}
