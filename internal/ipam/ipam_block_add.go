package ipam

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/lugnut42/openipam/internal/config"
)

func AddBlock(cfg *config.Config, cidr, description, fileKey string) error {
	log.Printf("DEBUG: AddBlock called with CIDR=%s, description=%s, fileKey=%s", cidr, description, fileKey)
	log.Printf("DEBUG: Config contents: %+v", cfg)

	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		log.Printf("ERROR: Block file not found for key '%s'. Available keys: %v", fileKey, cfg.BlockFiles)
		return fmt.Errorf("block file for key %s not found", fileKey)
	}
	log.Printf("DEBUG: Using block file: %s", blockFile)

	// Check if block file exists
	if _, err := os.Stat(blockFile); err != nil {
		log.Printf("ERROR: Block file does not exist: %v", err)
		return fmt.Errorf("block file %s does not exist: %w", blockFile, err)
	}

	// Validate CIDR
	_, newBlockNet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Printf("ERROR: Invalid CIDR format: %v", err)
		return fmt.Errorf("invalid CIDR: %w", err)
	}
	log.Printf("DEBUG: CIDR validation passed for %s", cidr)

	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		log.Printf("ERROR: Failed to read YAML file %s: %v", blockFile, err)
		return fmt.Errorf("error reading YAML file: %w", err)
	}
	log.Printf("DEBUG: Read YAML data from file (length: %d): %s", len(yamlData), string(yamlData))

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		log.Printf("ERROR: Failed to unmarshal blocks: %v", err)
		return fmt.Errorf("error unmarshalling YAML data: %w", err)
	}
	log.Printf("DEBUG: Unmarshalled %d existing blocks", len(blocks))
	for i, block := range blocks {
		log.Printf("DEBUG: Existing block %d: %+v", i, block)
	}

	// Check if the block already exists or overlaps with an existing block
	for _, b := range blocks {
		_, existingBlockNet, err := net.ParseCIDR(b.CIDR)
		if err != nil {
			log.Printf("ERROR: Failed to parse existing block CIDR %s: %v", b.CIDR, err)
			return fmt.Errorf("error parsing existing block CIDR: %w", err)
		}

		if newBlockNet.Contains(existingBlockNet.IP) || existingBlockNet.Contains(newBlockNet.IP) {
			log.Printf("ERROR: CIDR overlap detected between %s and %s", cidr, b.CIDR)
			return fmt.Errorf("block with CIDR %s overlaps with existing block %s", cidr, b.CIDR)
		}
	}
	log.Printf("DEBUG: No overlapping blocks found")

	// Add new block
	newBlock := Block{
		CIDR:        cidr,
		Description: description,
	}
	blocks = append(blocks, newBlock)
	log.Printf("DEBUG: Added new block: %+v", newBlock)

	newYamlData, err := marshalBlocks(blocks)
	if err != nil {
		log.Printf("ERROR: Failed to marshal blocks: %v", err)
		return fmt.Errorf("error marshalling blocks: %w", err)
	}
	log.Printf("DEBUG: Marshalled new YAML data (length: %d): %s", len(newYamlData), string(newYamlData))

	err = writeYAMLFile(blockFile, newYamlData)
	if err != nil {
		log.Printf("ERROR: Failed to write YAML file: %v", err)
		return fmt.Errorf("error writing YAML file: %w", err)
	}
	log.Printf("DEBUG: Successfully wrote updated blocks to file")

	return nil
}
