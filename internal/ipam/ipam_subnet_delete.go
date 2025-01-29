package ipam

import (
	"fmt"
	"net"

	"github.com/lugnut42/openipam/internal/config"
)

// DeleteSubnet deletes a subnet from a block
func DeleteSubnet(cfg *config.Config, subnetCIDR string, force bool) error {
	// Check if the force flag is set
	if !force {
		return fmt.Errorf("deletion requires --force flag for confirmation")
	}

	// Add CIDR validation here, before any file operations
	if _, _, err := net.ParseCIDR(subnetCIDR); err != nil {
		return fmt.Errorf("invalid subnet CIDR: %v", err)
	}

	subnetFound := false

	for _, blockFile := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return err
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return err
		}

		newBlocks := []Block{} // Create a new slice to store the remaining blocks

		for _, block := range blocks {
			newSubnets := []Subnet{} // Create a new slice to store the remaining subnets

			for _, subnet := range block.Subnets {
				if subnet.CIDR != subnetCIDR {
					newSubnets = append(newSubnets, subnet)
				} else {
					subnetFound = true
				}
			}

			block.Subnets = newSubnets
			newBlocks = append(newBlocks, block)
		}

		if subnetFound {
			newYamlData, err := marshalBlocks(newBlocks) // Marshal the updated blocks
			if err != nil {
				return err
			}
			err = writeYAMLFile(blockFile, newYamlData)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("subnet with CIDR %s not found", subnetCIDR) // Handle if subnet isn't found in any file
}
