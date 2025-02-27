package ipam

import (
	"net"
	"testing"

	"github.com/lugnut42/openipam/internal/config"
)

func TestPartialOverlap(t *testing.T) {
	// Test cases with network ranges that should overlap
	testCases := []struct{
		name string
		cidr1 string
		cidr2 string
		shouldOverlap bool
	}{
		{
			name: "Partial overlap case 1",
			cidr1: "172.16.0.0/16",  // 172.16.0.0 - 172.16.255.255
			cidr2: "172.16.128.0/17", // 172.16.128.0 - 172.16.255.255
			shouldOverlap: true,
		},
		{
			name: "Partial overlap case 2",
			cidr1: "10.0.0.0/8",     // 10.0.0.0 - 10.255.255.255
			cidr2: "10.10.0.0/16",   // 10.10.0.0 - 10.10.255.255
			shouldOverlap: true,
		},
		{
			name: "Complete containment",
			cidr1: "192.168.0.0/16", // 192.168.0.0 - 192.168.255.255
			cidr2: "192.168.0.0/24", // 192.168.0.0 - 192.168.0.255
			shouldOverlap: true,
		},
		{
			name: "Non-overlapping",
			cidr1: "172.16.0.0/16",  // 172.16.0.0 - 172.16.255.255
			cidr2: "172.17.0.0/16",  // 172.17.0.0 - 172.17.255.255
			shouldOverlap: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, cidr1, err := net.ParseCIDR(tc.cidr1)
			if err != nil {
				t.Fatalf("Failed to parse CIDR1 %s: %v", tc.cidr1, err)
			}
	
			_, cidr2, err := net.ParseCIDR(tc.cidr2)
			if err != nil {
				t.Fatalf("Failed to parse CIDR2 %s: %v", tc.cidr2, err)
			}
	
			overlaps := checkCIDROverlap(cidr1, cidr2)
			
			if overlaps != tc.shouldOverlap {
				t.Errorf("Expected overlap=%v, got %v", tc.shouldOverlap, overlaps)
				
				// Debug info
				cidr1Start := cidr1.IP
				cidr1End := lastIP(cidr1)
				cidr2Start := cidr2.IP
				cidr2End := lastIP(cidr2)
	
				t.Logf("CIDR1 range: %s - %s", cidr1Start, cidr1End)
				t.Logf("CIDR2 range: %s - %s", cidr2Start, cidr2End)
			}
		})
	}
}

func TestBlockDeletion(t *testing.T) {
	// Create a temporary file for testing
	tempFile := t.TempDir() + "/test_blocks.yaml"
	
	// Create an empty block file
	err := writeYAMLFile(tempFile, []byte("[]"))
	if err != nil {
		t.Fatalf("Failed to create test block file: %v", err)
	}
	
	// Setup config
	cfg := &config.Config{
		BlockFiles: map[string]string{
			"default": tempFile,
		},
	}

	// Try to delete non-existent block
	err = DeleteBlock(cfg, "192.168.1.0/24", true)
	if err == nil {
		t.Error("DeleteBlock should fail when block doesn't exist")
	}

	// Add a block
	err = AddBlock(cfg, "10.0.0.0/8", "test block", "default")
	if err != nil {
		t.Fatalf("Failed to add test block: %v", err)
	}

	// Try to delete it
	err = DeleteBlock(cfg, "10.0.0.0/8", true)
	if err != nil {
		t.Errorf("Failed to delete existing block: %v", err)
	}

	// Try to delete it again (should fail)
	err = DeleteBlock(cfg, "10.0.0.0/8", true)
	if err == nil {
		t.Error("DeleteBlock should fail when trying to delete non-existent block")
	}
}
