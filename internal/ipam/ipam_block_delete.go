package ipam

import (
	"fmt"
	"strings"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
)

func DeleteBlock(cfg *config.Config, cidr string, force bool, fileKey ...string) error {
	logger.Debug("DeleteBlock called with CIDR=%s, force=%v, fileKey=%v", cidr, force, fileKey)

	blockFound := false
	var blockFile string
	var blockFileKey string

	// Determine which block files to search
	blockFiles := make(map[string]string)
	if len(fileKey) > 0 && fileKey[0] != "" {
		// Use specified block file
		if file, ok := cfg.BlockFiles[fileKey[0]]; ok {
			blockFiles[fileKey[0]] = file
		} else {
			return fmt.Errorf("block file for key %s not found", fileKey[0])
		}
	} else {
		// Use all block files
		blockFiles = cfg.BlockFiles
	}

	// First pass: verify block exists
	for bfKey, bf := range blockFiles {
		logger.Debug("Checking file %s for block %s", bfKey, cidr)

		yamlData, err := readYAMLFile(bf)
		if err != nil {
			logger.Debug("Error reading block file %s: %v", bf, err)
			return fmt.Errorf("error reading block file %s: %w", bf, err)
		}

		blocks, err := unmarshalBlocks(yamlData)
		if err != nil {
			logger.Debug("Error parsing blocks from %s: %v", bf, err)
			return fmt.Errorf("error parsing blocks from %s: %w", bf, err)
		}

		logger.Debug("Found %d blocks in file %s", len(blocks), bfKey)
		for _, block := range blocks {
			logger.Debug("Comparing block CIDR %s with target %s", block.CIDR, cidr)
			if strings.TrimSpace(block.CIDR) == strings.TrimSpace(cidr) {
				blockFound = true
				blockFile = bf
				blockFileKey = bfKey
				logger.Debug("Found matching block in file %s", bfKey)
				break
			}
		}
		if blockFound {
			break
		}
	}

	if !blockFound {
		logger.Debug("Block %s not found in any file", cidr)
		return fmt.Errorf("block with CIDR %s not found", cidr)
	}

	logger.Debug("Found block %s in file %s (%s)", cidr, blockFile, blockFileKey)
	
	// Now remove the block
	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		logger.Debug("Error reading block file %s: %v", blockFile, err)
		return fmt.Errorf("error reading block file %s: %w", blockFile, err)
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		logger.Debug("Error parsing blocks from %s: %v", blockFile, err)
		return fmt.Errorf("error parsing blocks from %s: %w", blockFile, err)
	}

	// Find and remove the block
	var updatedBlocks []Block
	for _, block := range blocks {
		if strings.TrimSpace(block.CIDR) != strings.TrimSpace(cidr) {
			updatedBlocks = append(updatedBlocks, block)
		}
	}

	// Check if we actually removed a block
	if len(updatedBlocks) == len(blocks) {
		logger.Debug("Block %s not found in file %s (should not happen)", cidr, blockFile)
		return fmt.Errorf("block with CIDR %s not found in file %s", cidr, blockFile)
	}

	// Write the updated blocks back to the file
	newYamlData, err := marshalBlocks(updatedBlocks)
	if err != nil {
		logger.Debug("Error marshalling blocks: %v", err)
		return fmt.Errorf("error marshalling blocks: %w", err)
	}

	if err = writeYAMLFile(blockFile, newYamlData); err != nil {
		logger.Debug("Error writing block file %s: %v", blockFile, err)
		return fmt.Errorf("error writing block file %s: %w", blockFile, err)
	}

	logger.Debug("Successfully deleted block %s from file %s", cidr, blockFileKey)
	return nil
}