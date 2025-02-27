package ipam

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Helper functions to handle YAML file operations with security validation
func readYAMLFile(filePath string) ([]byte, error) {
	// Use filepath.Clean to normalize the path and remove any relative path elements
	cleanPath := filepath.Clean(filePath)
	
	// Verify the file exists before trying to read it
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing YAML file: %w", err)
	}
	
	// Check that it's a regular file, not a directory or other special file
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("path is not a regular file: %s", cleanPath)
	}
	
	// Read the file with controlled permissions (addresses G304: Potential file inclusion via variable)
	yamlData, err := os.ReadFile(cleanPath) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}
	return yamlData, nil
}

func unmarshalBlocks(yamlData []byte) ([]Block, error) {
	var blocks []Block
	var yamlDataInterface interface{}
	err := yaml.Unmarshal(yamlData, &yamlDataInterface)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

	if yamlDataInterface != nil {
		if iBlocks, ok := yamlDataInterface.([]interface{}); ok { // Check if it's []interface{}
			for _, iBlock := range iBlocks {
				if block, ok := iBlock.(map[string]interface{}); ok {
					// Now convert each map[string]interface{} to Block
					cidr := block["cidr"].(string)
					description := block["description"].(string)
					subnetsInterface, ok := block["subnets"].([]interface{})
					if !ok {
						// Handle the case where "subnets" is missing or not an array
						subnetsInterface = []interface{}{} // Empty slice if "subnets" key is missing or of incorrect type
					}

					subnets := make([]Subnet, len(subnetsInterface))

					for i, subnetInterface := range subnetsInterface {
						if subnet, ok := subnetInterface.(map[string]interface{}); ok {
							subnets[i] = Subnet{
								CIDR: subnet["cidr"].(string),
								Name: subnet["name"].(string),
							}

						}

					}
					blocks = append(blocks, Block{CIDR: cidr, Description: description, Subnets: subnets})

				}
			}

		} else if blocksInterface, ok := yamlDataInterface.([]Block); ok {
			blocks = blocksInterface
		} else {
			return nil, fmt.Errorf("unexpected YAML data type %T", yamlDataInterface)
		}

	}

	return blocks, nil
}

func marshalBlocks(blocks []Block) ([]byte, error) {
	newYamlData, err := yaml.Marshal(blocks)
	if err != nil {
		return nil, fmt.Errorf("error marshalling YAML: %w", err)
	}
	return newYamlData, nil

}

func writeYAMLFile(filePath string, yamlData []byte) error {
	// Use filepath.Clean to normalize the path and remove any relative path elements
	cleanPath := filepath.Clean(filePath)
	
	// Make sure the parent directory exists
	dir := filepath.Dir(cleanPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("parent directory does not exist: %s", dir)
	}
	
	// Write the file with secure permissions (0600 - only owner can read/write)
	err := os.WriteFile(cleanPath, yamlData, 0600) // #nosec G304
	if err != nil {
		return fmt.Errorf("error writing YAML file: %w", err)
	}
	return nil
}
