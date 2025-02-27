package ipam

import (
	"fmt"
	"math/big"
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

	w.Flush()
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
	w.Flush()

	return nil
}

// calculateIPCount calculates the number of IP addresses in a subnet
func calculateIPCount(ipNet *net.IPNet) uint64 {
	// For IPv4
	ones, bits := ipNet.Mask.Size()
	if bits == 32 { // IPv4
		// Special case for /31 and /32 networks
		if ones >= 31 {
			return uint64(1) << uint64(32-ones)
		}
		// Account for network and broadcast addresses
		return uint64(1)<<uint64(32-ones) - 2
	} else { // IPv6
		// Convert to big integers for IPv6
		maskLen := bits - ones
		return networkSize(maskLen)
	}
}

// networkSize returns the size of a network with the given mask length
func networkSize(maskLen int) uint64 {
	if maskLen >= 64 {
		return 1 << uint(maskLen)
	}

	// For larger networks, use big.Int
	size := new(big.Int).Lsh(big.NewInt(1), uint(maskLen))
	// If the result fits in uint64, return it
	if size.IsUint64() {
		return size.Uint64()
	}
	// Otherwise, return max uint64 (this is a limitation, but IPv6 networks
	// can be extremely large)
	return ^uint64(0)
}
