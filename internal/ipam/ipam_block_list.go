package ipam

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

func ListBlocks(cfg *config.Config, fileKey ...string) error {
	// Get all block files or a specific one
	blockFiles := make(map[string]string)

	if len(fileKey) > 0 && fileKey[0] != "" {
		// Use a specific block file
		specificFile, ok := cfg.BlockFiles[fileKey[0]]
		if !ok {
			return fmt.Errorf("block file for key %s not found", fileKey[0])
		}
		blockFiles[fileKey[0]] = specificFile
	} else {
		// Use all block files
		for key, file := range cfg.BlockFiles {
			blockFiles[key] = file
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Block CIDR\tSubnet CIDR\tDescription")

	for _, blockFile := range blockFiles {
		yamlData, err := readYAMLFile(blockFile)
		if err != nil {
			return fmt.Errorf("error reading block file: %w", err)
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return fmt.Errorf("error parsing blocks: %w", err)
		}

		for _, block := range blocks {
			if len(block.Subnets) > 0 {
				for _, subnet := range block.Subnets {
					fmt.Fprintln(w, block.CIDR+"\t"+subnet.CIDR+"\t"+block.Description)
				}
			} else {
				fmt.Fprintln(w, block.CIDR+"\t\t"+block.Description)
			}
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}
	return nil
}
