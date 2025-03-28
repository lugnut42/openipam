package ipam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

// ListSubnets lists all subnets within a block
func ListSubnets(cfg *config.Config, blockCIDR, region string) error {
	// Create a single tabwriter for all results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Block CIDR\tSubnet CIDR\tName\tRegion") // Table header
	
	// Track if we found any subnets
	foundSubnets := false
	
	// Iterate through all block files
	for _, blockFile := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return err
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return err
		}

		for _, block := range blocks {
			// Optionally filter by blockCIDR
			if blockCIDR != "" && block.CIDR != blockCIDR {
				continue // Skip blocks that don't match the filter
			}

			for _, subnet := range block.Subnets {
				// Optionally filter by region
				if region != "" && region != subnet.Region {
					continue // Skip subnets that don't match the region
				}

				fmt.Fprintln(w, block.CIDR+"\t"+subnet.CIDR+"\t"+subnet.Name+"\t"+subnet.Region)
				foundSubnets = true
			}
		}
	}
	
	// Only flush if we found subnets
	if foundSubnets {
		if err := w.Flush(); err != nil {
			return fmt.Errorf("error flushing writer: %w", err)
		}
	} else {
		fmt.Println("No subnets found.")
	}
	
	return nil
}
