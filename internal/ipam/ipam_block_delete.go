package ipam

import (
	"fmt"

	"github.com/lugnut42/openipam/internal/config"
)

func DeleteBlock(cfg *config.Config, cidr string, force bool) error {
	for _, blockFile := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return err
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return err
		}

		if !force {
			// Prompt for confirmation if --force is not set
			var confirmation string
			fmt.Printf("Are you sure you want to delete block %s? (yes/no): ", cidr)
			if _, err := fmt.Scanln(&confirmation); err != nil {
				return err
			}

			if confirmation != "yes" {
				return fmt.Errorf("deletion cancelled")
			}
		}

		newBlocks := []Block{} // Create a new slice to store the remaining blocks

		for _, block := range blocks {
			if block.CIDR != cidr {
				newBlocks = append(newBlocks, block)
			}
		}

		if len(blocks) == len(newBlocks) {
			continue // If no block was removed, continue to the next file
		}

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

	return fmt.Errorf("block with CIDR %s not found", cidr) // Handle if block isn't found in any file
}
