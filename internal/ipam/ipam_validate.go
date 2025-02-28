package ipam

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/lugnut42/openipam/internal/logger"
	"gopkg.in/yaml.v3"
)

// ValidationResult represents a validation error or warning
type ValidationResult struct {
	Type        string // "error" or "warning"
	File        string // File where the issue was detected
	Category    string // Category of the validation (e.g., "structure", "cidr", "reference")
	Description string // Description of the issue
	Location    string // Location in the file (e.g., "blocks.10.0.0.0/16.subnets.0")
}

// ValidationResults holds all validation results for a file
type ValidationResults struct {
	Filename     string
	ErrorCount   int
	WarningCount int
	Results      []ValidationResult
}

// ValidateBlockFile performs comprehensive validation on a block file
func ValidateBlockFile(cfg *config.Config, fileKey string) (*ValidationResults, error) {
	filepath, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return nil, fmt.Errorf("block file for key %s not found", fileKey)
	}

	// Initialize results
	results := &ValidationResults{
		Filename: filepath,
		Results:  []ValidationResult{},
	}

	// Read the YAML file
	yamlData, err := readYAMLFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %w", err)
	}

	// Validate YAML structure
	validateYAMLStructure(yamlData, fileKey, results)

	// If we can parse the blocks, perform additional validations
	blocks, err := unmarshalBlocks(yamlData)
	if err == nil {
		validateBlocks(blocks, fileKey, results)
		validateSubnets(blocks, fileKey, results)
		validateCrossReferences(blocks, cfg, fileKey, results)
	}

	// Count errors and warnings
	results.ErrorCount = 0
	results.WarningCount = 0
	for _, r := range results.Results {
		if r.Type == "error" {
			results.ErrorCount++
		} else if r.Type == "warning" {
			results.WarningCount++
		}
	}

	return results, nil
}

// validateYAMLStructure checks if the YAML has the expected structure
// This function is now compatible with both formats:
// 1. A list of blocks (application format)
// 2. A map with "blocks" key (validation format)
func validateYAMLStructure(yamlData []byte, fileKey string, results *ValidationResults) {
	// First try to unmarshal as a list of blocks (application format)
	var blocksList []interface{}
	err := yaml.Unmarshal(yamlData, &blocksList)
	
	// If successful and it's a non-empty list, validate as list format
	if err == nil && len(blocksList) > 0 {
		// It's a list format, which is valid for the application
		// Check each block in the list
		for i, blockData := range blocksList {
			blockMap, ok := blockData.(map[string]interface{})
			if !ok {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "structure",
					Description: "Block data is not a map",
					Location:    fmt.Sprintf("blocks[%d]", i),
				})
				continue
			}
			
			// Check if cidr exists and is valid
			cidrValue, ok := blockMap["cidr"]
			if !ok {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "structure",
					Description: "Block is missing required 'cidr' field",
					Location:    fmt.Sprintf("blocks[%d]", i),
				})
				continue
			}
			
			cidr, ok := cidrValue.(string)
			if !ok {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "structure",
					Description: "Block 'cidr' field is not a string",
					Location:    fmt.Sprintf("blocks[%d].cidr", i),
				})
				continue
			}
			
			// Validate CIDR format
			_, _, err := net.ParseCIDR(cidr)
			if err != nil {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "cidr",
					Description: fmt.Sprintf("Invalid CIDR format: %s", err),
					Location:    fmt.Sprintf("blocks[%d].cidr", i),
				})
			}
		}
		
		// Successfully validated list format
		return
	}
	
	// If list format didn't work, try map format (older validation format)
	var yamlMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &yamlMap)
	if err != nil {
		results.Results = append(results.Results, ValidationResult{
			Type:        "error",
			File:        fileKey,
			Category:    "structure",
			Description: fmt.Sprintf("File is not a valid YAML: %s", err),
			Location:    "root",
		})
		return
	}

	// Check if the "blocks" key exists and is a map
	blocks, ok := yamlMap["blocks"]
	if !ok {
		results.Results = append(results.Results, ValidationResult{
			Type:        "error",
			File:        fileKey,
			Category:    "structure",
			Description: "File does not contain a 'blocks' key",
			Location:    "root",
		})
		return
	}

	blocksMap, ok := blocks.(map[string]interface{})
	if !ok {
		results.Results = append(results.Results, ValidationResult{
			Type:        "error",
			File:        fileKey,
			Category:    "structure",
			Description: "'blocks' is not a map of CIDR to block details",
			Location:    "blocks",
		})
		return
	}

	// Check each block
	for cidr, blockData := range blocksMap {
		// Validate the CIDR format
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			results.Results = append(results.Results, ValidationResult{
				Type:        "error",
				File:        fileKey,
				Category:    "cidr",
				Description: fmt.Sprintf("Invalid CIDR format: %s", err),
				Location:    fmt.Sprintf("blocks.%s", cidr),
			})
		}

		// Check block structure
		blockMap, ok := blockData.(map[string]interface{})
		if !ok {
			results.Results = append(results.Results, ValidationResult{
				Type:        "error",
				File:        fileKey,
				Category:    "structure",
				Description: "Block data is not a map",
				Location:    fmt.Sprintf("blocks.%s", cidr),
			})
			continue
		}

		// Check if description exists (optional but recommended)
		if _, ok := blockMap["description"]; !ok {
			results.Results = append(results.Results, ValidationResult{
				Type:        "warning",
				File:        fileKey,
				Category:    "metadata",
				Description: "Block has no description",
				Location:    fmt.Sprintf("blocks.%s", cidr),
			})
		}

		// Check subnets if they exist
		if subnetsData, ok := blockMap["subnets"]; ok {
			subnetsMap, ok := subnetsData.(map[string]interface{})
			if !ok {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "structure",
					Description: "Subnets data is not a map",
					Location:    fmt.Sprintf("blocks.%s.subnets", cidr),
				})
				continue
			}

			// Check each subnet
			for subnetCIDR, subnetData := range subnetsMap {
				// Validate the subnet CIDR format
				_, _, err := net.ParseCIDR(subnetCIDR)
				if err != nil {
					results.Results = append(results.Results, ValidationResult{
						Type:        "error",
						File:        fileKey,
						Category:    "cidr",
						Description: fmt.Sprintf("Invalid subnet CIDR format: %s", err),
						Location:    fmt.Sprintf("blocks.%s.subnets.%s", cidr, subnetCIDR),
					})
				}

				// Check subnet structure
				subnetMap, ok := subnetData.(map[string]interface{})
				if !ok {
					results.Results = append(results.Results, ValidationResult{
						Type:        "error",
						File:        fileKey,
						Category:    "structure",
						Description: "Subnet data is not a map",
						Location:    fmt.Sprintf("blocks.%s.subnets.%s", cidr, subnetCIDR),
					})
					continue
				}

				// Check required subnet fields
				for _, field := range []string{"name", "region"} {
					if _, ok := subnetMap[field]; !ok {
						results.Results = append(results.Results, ValidationResult{
							Type:        "error",
							File:        fileKey,
							Category:    "metadata",
							Description: fmt.Sprintf("Subnet missing required field: %s", field),
							Location:    fmt.Sprintf("blocks.%s.subnets.%s", cidr, subnetCIDR),
						})
					}
				}
			}
		}
	}
}

// validateBlocks performs validations on block data
func validateBlocks(blocks []Block, fileKey string, results *ValidationResults) {
	// Check for duplicate CIDRs
	seenCIDRs := make(map[string]bool)
	for _, block := range blocks {
		if seenCIDRs[block.CIDR] {
			results.Results = append(results.Results, ValidationResult{
				Type:        "error",
				File:        fileKey,
				Category:    "duplicate",
				Description: fmt.Sprintf("Duplicate block CIDR: %s", block.CIDR),
				Location:    fmt.Sprintf("blocks.%s", block.CIDR),
			})
		}
		seenCIDRs[block.CIDR] = true

		// Validate CIDR format
		_, blockNet, err := net.ParseCIDR(block.CIDR)
		if err != nil {
			results.Results = append(results.Results, ValidationResult{
				Type:        "error",
				File:        fileKey,
				Category:    "cidr",
				Description: fmt.Sprintf("Invalid CIDR format: %s", err),
				Location:    fmt.Sprintf("blocks.%s", block.CIDR),
			})
			continue
		}

		// Check if block has a description
		if block.Description == "" {
			results.Results = append(results.Results, ValidationResult{
				Type:        "warning",
				File:        fileKey,
				Category:    "metadata",
				Description: "Block has no description",
				Location:    fmt.Sprintf("blocks.%s", block.CIDR),
			})
		}

		// Check for overlapping blocks within this file
		for _, otherBlock := range blocks {
			if block.CIDR == otherBlock.CIDR {
				continue // Skip self-comparison
			}

			_, otherBlockNet, err := net.ParseCIDR(otherBlock.CIDR)
			if err != nil {
				continue // Skip invalid CIDRs, they are reported elsewhere
			}

			if checkCIDROverlap(blockNet, otherBlockNet) {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "overlap",
					Description: fmt.Sprintf("Block %s overlaps with block %s", block.CIDR, otherBlock.CIDR),
					Location:    fmt.Sprintf("blocks.%s", block.CIDR),
				})
			}
		}
	}
}

// validateSubnets performs validations on subnet data
func validateSubnets(blocks []Block, fileKey string, results *ValidationResults) {
	for _, block := range blocks {
		_, blockNet, err := net.ParseCIDR(block.CIDR)
		if err != nil {
			continue // Skip invalid blocks, they are reported elsewhere
		}

		// Check for duplicate subnet CIDRs within the block
		seenSubnetCIDRs := make(map[string]bool)
		seenSubnetNames := make(map[string]bool)

		for i, subnet := range block.Subnets {
			location := fmt.Sprintf("blocks.%s.subnets[%d]", block.CIDR, i)

			// Check for duplicate CIDRs
			if seenSubnetCIDRs[subnet.CIDR] {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "duplicate",
					Description: fmt.Sprintf("Duplicate subnet CIDR: %s", subnet.CIDR),
					Location:    location,
				})
			}
			seenSubnetCIDRs[subnet.CIDR] = true

			// Check for duplicate names (should be unique within a block)
			if seenSubnetNames[subnet.Name] {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "duplicate",
					Description: fmt.Sprintf("Duplicate subnet name: %s", subnet.Name),
					Location:    location,
				})
			}
			seenSubnetNames[subnet.Name] = true

			// Validate subnet CIDR format
			_, subnetNet, err := net.ParseCIDR(subnet.CIDR)
			if err != nil {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "cidr",
					Description: fmt.Sprintf("Invalid subnet CIDR format: %s", err),
					Location:    location,
				})
				continue
			}

			// Check if subnet is within its parent block
			if !blockNet.Contains(subnetNet.IP) {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "containment",
					Description: fmt.Sprintf("Subnet %s is not contained within its parent block %s", subnet.CIDR, block.CIDR),
					Location:    location,
				})
			}

			// Check for required fields
			if subnet.Name == "" {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "metadata",
					Description: "Subnet missing required field: name",
					Location:    location,
				})
			}

			if subnet.Region == "" {
				results.Results = append(results.Results, ValidationResult{
					Type:        "error",
					File:        fileKey,
					Category:    "metadata",
					Description: "Subnet missing required field: region",
					Location:    location,
				})
			}

			// Check for overlapping subnets within this block
			for j, otherSubnet := range block.Subnets {
				if i == j {
					continue // Skip self-comparison
				}

				_, otherSubnetNet, err := net.ParseCIDR(otherSubnet.CIDR)
				if err != nil {
					continue // Skip invalid CIDRs, they are reported elsewhere
				}

				if checkCIDROverlap(subnetNet, otherSubnetNet) {
					results.Results = append(results.Results, ValidationResult{
						Type:        "error",
						File:        fileKey,
						Category:    "overlap",
						Description: fmt.Sprintf("Subnet %s overlaps with subnet %s", subnet.CIDR, otherSubnet.CIDR),
						Location:    location,
					})
				}
			}
		}
	}
}

// validateCrossReferences checks references between different parts of the configuration
func validateCrossReferences(blocks []Block, cfg *config.Config, fileKey string, results *ValidationResults) {
	// Check patterns that reference blocks in this file
	for patternName, pattern := range cfg.Patterns {
		for _, p := range pattern {
			// Check if the pattern references a block in this file
			if _, ok := cfg.BlockFiles[fileKey]; ok && p.Block != "" {
				// Check if the referenced block exists
				blockExists := false
				for _, block := range blocks {
					if block.CIDR == p.Block {
						blockExists = true
						break
					}
				}

				if !blockExists {
					results.Results = append(results.Results, ValidationResult{
						Type:        "error",
						File:        fileKey,
						Category:    "reference",
						Description: fmt.Sprintf("Pattern '%s' references non-existent block: %s", patternName, p.Block),
						Location:    fmt.Sprintf("patterns.%s", patternName),
					})
				}
			}
		}
	}
}

// PrintValidationResults displays validation results in a formatted table
func PrintValidationResults(results *ValidationResults) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Validation Results for: %s\n", results.Filename)
	fmt.Fprintf(w, "Errors: %d  Warnings: %d\n\n", results.ErrorCount, results.WarningCount)

	if len(results.Results) == 0 {
		fmt.Fprintln(w, "No issues found. Configuration is valid.")
		if err := w.Flush(); err != nil {
			return err
		}
		return nil
	}

	fmt.Fprintln(w, "Type\tCategory\tLocation\tDescription")
	fmt.Fprintln(w, "----\t--------\t--------\t-----------")

	for _, result := range results.Results {
		var typeStr string
		if result.Type == "error" {
			typeStr = "ERROR"
		} else {
			typeStr = "WARNING"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			typeStr,
			result.Category,
			result.Location,
			result.Description)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

// ValidateAllBlockFiles validates all block files in the configuration
func ValidateAllBlockFiles(cfg *config.Config) error {
	if len(cfg.BlockFiles) == 0 {
		fmt.Println("No block files configured.")
		return nil
	}

	totalErrors := 0
	totalWarnings := 0

	for fileKey := range cfg.BlockFiles {
		results, err := ValidateBlockFile(cfg, fileKey)
		if err != nil {
			logger.Debug("Error validating block file %s: %v", fileKey, err)
			fmt.Printf("Error validating block file %s: %v\n", fileKey, err)
			continue
		}

		fmt.Printf("\n=== Block File: %s ===\n", fileKey)
		if err := PrintValidationResults(results); err != nil {
			return fmt.Errorf("error printing validation results: %w", err)
		}

		totalErrors += results.ErrorCount
		totalWarnings += results.WarningCount
	}

	fmt.Printf("\nValidation Summary\n")
	fmt.Printf("Total Errors: %d  Total Warnings: %d\n", totalErrors, totalWarnings)

	if totalErrors > 0 {
		return fmt.Errorf("validation found %d errors across all block files", totalErrors)
	}

	return nil
}
