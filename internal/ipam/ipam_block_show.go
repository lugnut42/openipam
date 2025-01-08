package ipam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

func ShowBlock(cfg *config.Config, cidr, fileKey string) error {
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

	for _, block := range blocks {
		if block.CIDR == cidr {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Block CIDR\tDescription")
			fmt.Fprintln(w, block.CIDR+"\t"+block.Description)
			fmt.Fprintln(w, "\nSubnets:")
			fmt.Fprintln(w, "Subnet CIDR\tName\tRegion")
			for _, subnet := range block.Subnets {
				fmt.Fprintln(w, subnet.CIDR+"\t"+subnet.Name+"\t"+subnet.Region)
			}
			w.Flush()
			return nil
		}
	}

	return fmt.Errorf("block with CIDR %s not found", cidr)
}
