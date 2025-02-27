package ipam

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

// Intentionally removed debugIP function as it was unused

// compareIP compares two IP addresses lexicographically
func compareIP(a, b net.IP) int {
	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// Intentionally removed unused IP range functions

// Using lastIP from ipam_block.go

// checkCIDROverlap checks if two CIDR ranges overlap
func checkCIDROverlap(cidr1, cidr2 *net.IPNet) bool {
	// Get the actual network addresses (masked)
	net1 := cidr1.IP.Mask(cidr1.Mask)
	net2 := cidr2.IP.Mask(cidr2.Mask)

	// Get the broadcast addresses
	broadcast1 := lastIP(cidr1)
	broadcast2 := lastIP(cidr2)

	// Simple containment check: if either range contains endpoints of the other
	if cidr1.Contains(net2) || cidr1.Contains(broadcast2) || 
	   cidr2.Contains(net1) || cidr2.Contains(broadcast1) {
		return true
	}

	// Special case check for IPv4 partial overlap scenarios
	// This works by converting IPs to integers and checking for range overlaps
	if len(net1) == 4 && len(net2) == 4 { // IPv4 addresses
		// Convert IPs to integers for easier range comparison
		net1Int := (uint32(net1[0]) << 24) | (uint32(net1[1]) << 16) | (uint32(net1[2]) << 8) | uint32(net1[3])
		net2Int := (uint32(net2[0]) << 24) | (uint32(net2[1]) << 16) | (uint32(net2[2]) << 8) | uint32(net2[3])
		
		broadcast1Int := (uint32(broadcast1[0]) << 24) | (uint32(broadcast1[1]) << 16) | (uint32(broadcast1[2]) << 8) | uint32(broadcast1[3])
		broadcast2Int := (uint32(broadcast2[0]) << 24) | (uint32(broadcast2[1]) << 16) | (uint32(broadcast2[2]) << 8) | uint32(broadcast2[3])
		
		// Check for range overlap using integer comparison
		return (net1Int <= broadcast2Int && broadcast1Int >= net2Int)
	}

	// More general case for IPv6 or other scenarios - check IP ranges for overlap
	return (compareIP(net1, broadcast2) <= 0 && compareIP(broadcast1, net2) >= 0)
}

// UtilizationReport represents the utilization statistics for a block or subnet
type UtilizationReport struct {
	CIDR             string  `json:"cidr"`
	TotalIPs         uint64  `json:"total_ips"`
	AllocatedIPs     uint64  `json:"allocated_ips"`
	AvailableIPs     uint64  `json:"available_ips"`
	UtilizationRatio float64 `json:"utilization_ratio"`
}

// CalculateBlockUtilization calculates the IP address utilization for a specific block
func CalculateBlockUtilization(cfg *config.Config, blockCIDR, fileKey string) (*UtilizationReport, error) {
	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return nil, fmt.Errorf("block file for key %s not found", fileKey)
	}

	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		return nil, err
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		return nil, err
	}

	var block *Block
	for _, b := range blocks {
		if b.CIDR == blockCIDR {
			block = &b
			break
		}
	}

	if block == nil {
		return nil, fmt.Errorf("block %s not found", blockCIDR)
	}

	_, ipNet, err := net.ParseCIDR(block.CIDR)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %s", err)
	}

	// Calculate total IPs in the block
	blockSize := calculateIPCount(ipNet)

	// Calculate allocated IPs (sum of all subnet sizes)
	var allocatedSize uint64 = 0
	for _, subnet := range block.Subnets {
		_, subnetNet, err := net.ParseCIDR(subnet.CIDR)
		if err != nil {
			continue // Skip invalid subnets
		}
		allocatedSize += calculateIPCount(subnetNet)
	}

	// Calculate utilization ratio
	utilizationRatio := float64(0)
	if blockSize > 0 {
		utilizationRatio = float64(allocatedSize) / float64(blockSize)
	}

	return &UtilizationReport{
		CIDR:             block.CIDR,
		TotalIPs:         blockSize,
		AllocatedIPs:     allocatedSize,
		AvailableIPs:     blockSize - allocatedSize,
		UtilizationRatio: utilizationRatio,
	}, nil
}

// PrintBlockUtilization prints the utilization report for a specific block
func PrintBlockUtilization(cfg *config.Config, blockCIDR, fileKey string) error {
	report, err := CalculateBlockUtilization(cfg, blockCIDR, fileKey)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Block Utilization Report")
	fmt.Fprintln(w, "----------------------")
	fmt.Fprintf(w, "CIDR:\t%s\n", report.CIDR)
	fmt.Fprintf(w, "Total IPs:\t%d\n", report.TotalIPs)
	fmt.Fprintf(w, "Allocated IPs:\t%d\n", report.AllocatedIPs)
	fmt.Fprintf(w, "Available IPs:\t%d\n", report.AvailableIPs)
	fmt.Fprintf(w, "Utilization:\t%.2f%%\n", report.UtilizationRatio*100)
	
	// List all subnets with their contribution to utilization
	blockFile, ok := cfg.BlockFiles[fileKey]
	if ok {
		yamlData, err := readYAMLFile(blockFile)
		if err == nil {
			blocks, err := unmarshalBlocks(yamlData)
			if err == nil {
				for _, block := range blocks {
					if block.CIDR == blockCIDR && len(block.Subnets) > 0 {
						fmt.Fprintln(w, "\nSubnets:")
						fmt.Fprintln(w, "CIDR\tName\tRegion\tIP Count\t% of Block")
						fmt.Fprintln(w, "----\t----\t------\t--------\t---------")
						
						// Calculate and print each subnet's contribution
						for _, subnet := range block.Subnets {
							_, subnetNet, err := net.ParseCIDR(subnet.CIDR)
							if err == nil {
								subnetSize := calculateIPCount(subnetNet)
								percentage := float64(0)
								if report.TotalIPs > 0 {
									percentage = float64(subnetSize) / float64(report.TotalIPs) * 100
								}
								fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%.2f%%\n", 
									subnet.CIDR, 
									subnet.Name, 
									subnet.Region, 
									subnetSize,
									percentage)
							}
						}
					}
				}
			}
		}
	}
	
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}
	return nil
}

// PrintAllBlocksUtilization prints utilization reports for all blocks
func PrintAllBlocksUtilization(cfg *config.Config, fileKey string) error {
	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return fmt.Errorf("block file for key %s not found", fileKey)
	}

	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		return err
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		return err
	}

	if len(blocks) == 0 {
		fmt.Println("No blocks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CIDR\tTotal IPs\tAllocated IPs\tAvailable IPs\tUtilization")
	fmt.Fprintln(w, "----\t---------\t-------------\t-------------\t-----------")

	for _, block := range blocks {
		report, err := CalculateBlockUtilization(cfg, block.CIDR, fileKey)
		if err != nil {
			continue // Skip blocks with errors
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%.2f%%\n", 
			report.CIDR, 
			report.TotalIPs, 
			report.AllocatedIPs, 
			report.AvailableIPs, 
			report.UtilizationRatio*100)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

// calculateIPCount calculates the number of IP addresses in a subnet
func calculateIPCount(ipNet *net.IPNet) uint64 {
	// For IPv4
	ones, bits := ipNet.Mask.Size()
	if bits == 32 { // IPv4
		// Special case for /31 and /32 networks
		// Safely handle mask size
		if ones > 32 {
			ones = 32 // Prevent negative result
		}
		
		// Now we know ones is definitely in the range [0,32]
		maskBits := 32 - ones // Safe: will be in range [0,32]
		
		// Handle special cases directly to avoid any shift operations that could trigger overflow warnings
		switch maskBits {
		case 0:
			return 1
		case 1:
			return 2
		case 2:
			return 2
		case 3:
			return 6
		case 4:
			return 14
		case 5:
			return 30
		case 6:
			return 62
		case 7:
			return 126
		case 8:
			return 254
		case 9:
			return 510
		case 10:
			return 1022
		case 11:
			return 2046
		case 12:
			return 4094
		case 13:
			return 8190
		case 14:
			return 16382
		case 15:
			return 32766
		case 16:
			return 65534
		case 17:
			return 131070
		case 18:
			return 262142
		case 19:
			return 524286
		case 20:
			return 1048574
		case 21:
			return 2097150
		case 22:
			return 4194302
		case 23:
			return 8388606
		case 24:
			return 16777214
		case 25:
			return 33554430
		case 26:
			return 67108862
		case 27:
			return 134217726
		case 28:
			return 268435454
		case 29:
			return 536870910
		case 30:
			return 1073741822
		case 31:
			return 2147483646
		case 32:
			return 4294967294
		default:
			// This should never happen given our range check
			return 0
		}
	} else { // IPv6
		// Convert to big integers for IPv6
		maskLen := bits - ones
		return networkSize(maskLen)
	}
}

// networkSize returns the size of a network with the given mask length
func networkSize(maskLen int) uint64 {
	// Ensure maskLen is not negative
	if maskLen < 0 {
		maskLen = 0
	}
	
	// Handle direct cases to avoid bitshift overflow
	if maskLen == 0 {
		return 1
	} else if maskLen >= 64 {
		// For very large mask lengths, we need to be careful with integer overflow
		return ^uint64(0) // max uint64
	}
	
	// For maskLen in range [1,63], use a lookup table to avoid shift warnings
	switch maskLen {
	case 1:
		return 2
	case 2:
		return 4
	case 3:
		return 8
	case 4:
		return 16
	case 5:
		return 32
	case 6:
		return 64
	case 7:
		return 128
	case 8:
		return 256
	case 9:
		return 512
	case 10:
		return 1024
	case 11:
		return 2048
	case 12:
		return 4096
	case 13:
		return 8192
	case 14:
		return 16384
	case 15:
		return 32768
	case 16:
		return 65536
	case 17:
		return 131072
	case 18:
		return 262144
	case 19:
		return 524288
	case 20:
		return 1048576
	case 21:
		return 2097152
	case 22:
		return 4194304
	case 23:
		return 8388608
	case 24:
		return 16777216
	case 25:
		return 33554432
	case 26:
		return 67108864
	case 27:
		return 134217728
	case 28:
		return 268435456
	case 29:
		return 536870912
	case 30:
		return 1073741824
	case 31:
		return 2147483648
	case 32:
		return 4294967296
	case 33:
		return 8589934592
	case 34:
		return 17179869184
	case 35:
		return 34359738368
	case 36:
		return 68719476736
	case 37:
		return 137438953472
	case 38:
		return 274877906944
	case 39:
		return 549755813888
	case 40:
		return 1099511627776
	case 41:
		return 2199023255552
	case 42:
		return 4398046511104
	case 43:
		return 8796093022208
	case 44:
		return 17592186044416
	case 45:
		return 35184372088832
	case 46:
		return 70368744177664
	case 47:
		return 140737488355328
	case 48:
		return 281474976710656
	case 49:
		return 562949953421312
	case 50:
		return 1125899906842624
	case 51:
		return 2251799813685248
	case 52:
		return 4503599627370496
	case 53:
		return 9007199254740992
	case 54:
		return 18014398509481984
	case 55:
		return 36028797018963968
	case 56:
		return 72057594037927936
	case 57:
		return 144115188075855872
	case 58:
		return 288230376151711744
	case 59:
		return 576460752303423488
	case 60:
		return 1152921504606846976
	case 61:
		return 2305843009213693952
	case 62:
		return 4611686018427387904
	case 63:
		return 9223372036854775808
	default:
		return 0
	}
}