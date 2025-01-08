package ipam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

func ListBlocks(cfg *config.Config) error {
	for _, blockFile := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return err // Return the wrapped error
		}

		blocks, err := unmarshalBlocks(yamlData) // Use the helper function
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Block CIDR\tSubnet CIDR\tDescription")

		for _, block := range blocks {
			if len(block.Subnets) > 0 {
				for _, subnet := range block.Subnets {
					fmt.Fprintln(w, block.CIDR+"\t"+subnet.CIDR+"\t"+block.Description)
				}

			} else {
				fmt.Fprintln(w, block.CIDR+"\t\t"+block.Description)
			}
		}

		w.Flush()
	}
	return nil
}
