package ipam

import (
	"testing"

	"github.com/lugnut42/openipam/internal/config"
)

func TestCreateSubnet_NoAvailableCIDR(t *testing.T) {
	cfg := &config.Config{
		BlockFiles: map[string]string{"default": "test_block.yaml"},
	}

	// Create a block with no available CIDR
	block := Block{
		CIDR: "10.0.0.0/24",
		Subnets: []Subnet{
			{CIDR: "10.0.0.0/26"},
			{CIDR: "10.0.0.64/26"},
			{CIDR: "10.0.0.128/26"},
			{CIDR: "10.0.0.192/26"},
		},
	}

	// Write the block to a test YAML file
	yamlData, err := marshalBlocks([]Block{block})
	if err != nil {
		t.Fatalf("Failed to marshal blocks: %v", err)
	}
	err = writeYAMLFile("test_block.yaml", yamlData)
	if err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	// Attempt to create a new subnet, which should fail
	err = CreateSubnet(cfg, "10.0.0.0/24", "10.0.0.256/26", "test-subnet", "us-west")
	if err == nil {
		t.Fatalf("Expected error due to no available CIDR, but got nil")
	}
}

func TestCreateSubnetFromPattern_NoAvailableCIDR(t *testing.T) {
	cfg := &config.Config{
		BlockFiles: map[string]string{"default": "test_block.yaml"},
		Patterns: map[string]map[string]config.Pattern{
			"default": {
				"dev-gke-uswest": {
					Block:    "10.0.0.0/24",
					CIDRSize: 26,
					Region:   "us-west",
				},
			},
		},
	}

	// Create a block with no available CIDR
	block := Block{
		CIDR: "10.0.0.0/24",
		Subnets: []Subnet{
			{CIDR: "10.0.0.0/26"},
			{CIDR: "10.0.0.64/26"},
			{CIDR: "10.0.0.128/26"},
			{CIDR: "10.0.0.192/26"},
		},
	}

	// Write the block to a test YAML file
	yamlData, err := marshalBlocks([]Block{block})
	if err != nil {
		t.Fatalf("Failed to marshal blocks: %v", err)
	}
	err = writeYAMLFile("test_block.yaml", yamlData)
	if err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	// Attempt to create a new subnet from pattern, which should fail
	err = CreateSubnetFromPattern(cfg, "dev-gke-uswest", "default")
	if err == nil {
		t.Fatalf("Expected error due to no available CIDR, but got nil")
	}
}
