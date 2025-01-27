package ipam

import (
	"errors"
	"fmt"
	"net"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
)

// CreateSubnet creates a new subnet within a block
func CreateSubnet(cfg *config.Config, blockCIDR, subnetCIDR, name, region string) error {
	logger.Debug("Creating subnet: blockCIDR=%s, subnetCIDR=%s, name=%s, region=%s", blockCIDR, subnetCIDR, name, region)

	for _, blockFile := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return err
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return err
		}

		// Validate subnet CIDR (ensure it's a valid CIDR and within the block)
		_, subnetNet, err := net.ParseCIDR(subnetCIDR)
		if err != nil {
			return fmt.Errorf("invalid subnet CIDR: %w", err)
		}

		_, blockNet, err := net.ParseCIDR(blockCIDR)
		if err != nil {
			return fmt.Errorf("invalid block CIDR: %w", err)
		}

		if !blockNet.Contains(subnetNet.IP) {
			return errors.New("subnet is not within the specified block")
		}

		newSubnet := Subnet{
			CIDR:   subnetCIDR,
			Name:   name,
			Region: region,
		}

		// Find the block and add the subnet. If the block does not exist, return an error.
		found := false

		for i, block := range blocks {
			if block.CIDR == blockCIDR {
				// Check for available space in the block
				availableCIDRs := calculateAvailableCIDRs(&block)
				logger.Debug("Available CIDRs in block %s: %v", blockCIDR, availableCIDRs)
				if len(availableCIDRs) == 0 {
					return fmt.Errorf("no available CIDR found in block %s", block.CIDR)
				}

				// Check for overlapping subnets
				for _, existingSubnet := range block.Subnets {
					_, existingSubnetNet, err := net.ParseCIDR(existingSubnet.CIDR)
					if err != nil {
						return fmt.Errorf("error parsing existing subnet CIDR: %w", err)
					}

					if subnetNet.Contains(existingSubnetNet.IP) || existingSubnetNet.Contains(subnetNet.IP) {
						return fmt.Errorf("subnet with CIDR %s overlaps with existing subnet %s", subnetCIDR, existingSubnet.CIDR)
					}
				}

				blocks[i].Subnets = append(blocks[i].Subnets, newSubnet)
				found = true
				break // Exit the loop once the block is found
			}
		}

		if found {
			newYamlData, err := marshalBlocks(blocks)
			if err != nil {
				return err
			}
			err = writeYAMLFile(blockFile, newYamlData)
			if err != nil {
				return err
			}

			logger.Debug("Subnet created successfully: %s", subnetCIDR)
			return nil
		}
	}

	return fmt.Errorf("block with CIDR %s not found", blockCIDR)
}
