package ipam

import (
	"fmt"
	"log"
	"net"

	"github.com/lugnut42/openipam/internal/config"
)

// CreateSubnetFromPattern creates a new subnet from a pattern
func CreateSubnetFromPattern(cfg *config.Config, patternName, fileKey string) error {
	log.Printf("Creating subnet from pattern: patternName=%s, fileKey=%s", patternName, fileKey)

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

	// Check for available space in the block
	availableCIDRs := calculateAvailableCIDRs(block)
	log.Printf("Available CIDRs in block %s: %v", block.CIDR, availableCIDRs)
	if len(availableCIDRs) == 0 {
		return fmt.Errorf("no available CIDR found in block %s", block.CIDR)
	}

	// Find the next available subnet
	var newSubnetCIDR string
	_, blockNet, _ := net.ParseCIDR(block.CIDR)
	for ip := blockNet.IP.Mask(blockNet.Mask); blockNet.Contains(ip); incrementIP(ip) {
		subnet := fmt.Sprintf("%s/%d", ip.String(), pattern.CIDRSize)
		_, subnetNet, _ := net.ParseCIDR(subnet)
		if !isSubnetOverlapping(block.Subnets, subnetNet) {
			newSubnetCIDR = subnet
			break
		}
	}

	if newSubnetCIDR == "" {
		return fmt.Errorf("no available subnet found in block %s", block.CIDR)
	}

	newSubnet := Subnet{
		CIDR:   newSubnetCIDR,
		Name:   patternName,
		Region: pattern.Region,
	}

	block.Subnets = append(block.Subnets, newSubnet)

	newYamlData, err := marshalBlocks(blocks)
	if err != nil {
		return fmt.Errorf("error marshalling blocks: %w", err)
	}

	err = writeYAMLFile(blockFile, newYamlData)
	if err != nil {
		return fmt.Errorf("error writing YAML file: %w", err)
	}

	log.Printf("Subnet created successfully from pattern: %s", newSubnetCIDR)
	return nil
}
