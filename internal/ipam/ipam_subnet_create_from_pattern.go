package ipam

import (
	"fmt"
	"net"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
)

// CreateSubnetFromPattern creates a new subnet from a pattern
func CreateSubnetFromPattern(cfg *config.Config, patternName, fileKey string) error {
	logger.Debug("Creating subnet from pattern: patternName=%s, fileKey=%s", patternName, fileKey)

	patterns, ok := cfg.Patterns[fileKey]
	if !ok {
		return fmt.Errorf("patterns for file key %s not found", fileKey)
	}

	pattern, ok := patterns[patternName]
	if !ok {
		return fmt.Errorf("pattern %s not found", patternName)
	}

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

	var block *Block
	for i, b := range blocks {
		if b.CIDR == pattern.Block {
			block = &blocks[i]
			break
		}
	}

	if block == nil {
		return fmt.Errorf("block %s not found", pattern.Block)
	}

	// Get available CIDRs
	availableCIDRs := calculateAvailableCIDRs(block)
	logger.Debug("Available CIDRs in block %s: %v", block.CIDR, availableCIDRs)
	if len(availableCIDRs) == 0 {
		return fmt.Errorf("no available CIDR found in block %s", block.CIDR)
	}

	// Find an available CIDR that can accommodate our requested size
	var selectedCIDR string
	for _, availableCIDR := range availableCIDRs {
		_, availNet, err := net.ParseCIDR(availableCIDR)
		if err != nil {
			logger.Debug("Error parsing CIDR %s: %v", availableCIDR, err)
			continue
		}

		ones, _ := availNet.Mask.Size()
		if ones <= pattern.CIDRSize {
			// This CIDR is big enough to accommodate our requested size
			selectedCIDR = availableCIDR
			break
		}
	}

	if selectedCIDR == "" {
		return fmt.Errorf("no available CIDR found that can accommodate /%d subnet", pattern.CIDRSize)
	}

	// Calculate the specific subnet within the selected CIDR
	_, selectedNet, _ := net.ParseCIDR(selectedCIDR)
	newSubnetIP := selectedNet.IP
	newSubnetCIDR := fmt.Sprintf("%s/%d", newSubnetIP.String(), pattern.CIDRSize)

	// Verify the new subnet doesn't overlap with existing ones
	_, newSubnetNet, _ := net.ParseCIDR(newSubnetCIDR)
	if isSubnetOverlapping(block.Subnets, newSubnetNet) {
		return fmt.Errorf("calculated subnet %s overlaps with existing subnets", newSubnetCIDR)
	}

	// Create the new subnet
	newSubnet := Subnet{
		CIDR:   newSubnetCIDR,
		Name:   fmt.Sprintf("%s-%s", patternName, newSubnetIP.String()),
		Region: pattern.Region,
	}

	block.Subnets = append(block.Subnets, newSubnet)

	// Save the updated block configuration
	newYamlData, err := marshalBlocks(blocks)
	if err != nil {
		return fmt.Errorf("error marshalling blocks: %w", err)
	}

	err = writeYAMLFile(blockFile, newYamlData)
	if err != nil {
		return fmt.Errorf("error writing YAML file: %w", err)
	}

	logger.Debug("Subnet created successfully from pattern: %s", newSubnetCIDR)
	return nil
}