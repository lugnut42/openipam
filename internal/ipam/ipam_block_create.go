package ipam

import (
	"fmt"
	"net"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
)

func AddBlock(cfg *config.Config, cidr, description, fileKey string) error {
	logger.Debug("AddBlock called with CIDR=%s, description=%s, fileKey=%s", cidr, description, fileKey)

	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return fmt.Errorf("block file for key %s not found", fileKey)
	}

	// Validate CIDR
	_, newBlockNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	// Check for overlaps across all block files
	for bfKey, blockFilePath := range cfg.BlockFiles {
		yamlData, err := readYAMLFile(blockFilePath)
		if err != nil {
			return fmt.Errorf("error reading block file %s: %w", bfKey, err)
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			return fmt.Errorf("error parsing block file %s: %w", bfKey, err)
		}

		for _, b := range blocks {
			_, existingBlockNet, err := net.ParseCIDR(b.CIDR)
			if err != nil {
				return fmt.Errorf("error parsing existing block CIDR %s: %w", b.CIDR, err)
			}

			if checkCIDROverlap(newBlockNet, existingBlockNet) {
				return fmt.Errorf("block with CIDR %s overlaps with existing block %s in file %s", cidr, b.CIDR, bfKey)
			}
		}
	}

	// Now add the block to the specified file
	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		return fmt.Errorf("error reading block file: %w", err)
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		return fmt.Errorf("error parsing blocks: %w", err)
	}

	blocks = append(blocks, Block{
		CIDR:        cidr,
		Description: description,
	})

	newYamlData, err := marshalBlocks(blocks)
	if err != nil {
		return fmt.Errorf("error marshalling blocks: %w", err)
	}

	if err = writeYAMLFile(blockFile, newYamlData); err != nil {
		return fmt.Errorf("error writing block file: %w", err)
	}

	return nil
}
