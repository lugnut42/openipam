package ipam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

// ShowSubnet displays the details of a specific subnet
func ShowSubnet(cfg *config.Config, subnetCIDR string) error {
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
			for _, subnet := range block.Subnets {
				if subnet.CIDR == subnetCIDR {
					// Found the subnet
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "Block CIDR:\t", block.CIDR)
					fmt.Fprintln(w, "Subnet CIDR:\t", subnet.CIDR)
					fmt.Fprintln(w, "Name:\t", subnet.Name)
					fmt.Fprintln(w, "Region:\t", subnet.Region) // Include the Region

					w.Flush()

					return nil
				}
			}
		}
	}

	return fmt.Errorf("subnet with CIDR %s not found", subnetCIDR)
}
