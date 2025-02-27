package ipam

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

func ListAvailableCIDRs(cfg *config.Config, blockCIDR, fileKey string) error {
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

	var block *Block
	for _, b := range blocks {
		if b.CIDR == blockCIDR {
			block = &b
			break
		}
	}

	if block == nil {
		return fmt.Errorf("block %s not found", blockCIDR)
	}

	availableCIDRs := calculateAvailableCIDRs(block)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Available CIDR Ranges")
	for _, cidr := range availableCIDRs {
		fmt.Fprintln(w, cidr)
	}
	w.Flush()

	return nil
}

// calculateAvailableCIDRs returns a list of available CIDR blocks in the block
func calculateAvailableCIDRs(block *Block) []string {
	var availableCIDRs []string
	_, blockNet, _ := net.ParseCIDR(block.CIDR)
	blockSize, _ := blockNet.Mask.Size()

	// Sort subnets by starting IP
	subnets := block.Subnets
	sort.Slice(subnets, func(i, j int) bool {
		_, subnetNetI, _ := net.ParseCIDR(subnets[i].CIDR)
		_, subnetNetJ, _ := net.ParseCIDR(subnets[j].CIDR)
		return bytes.Compare(subnetNetI.IP, subnetNetJ.IP) < 0
	})

	// Start with the block's first IP
	currentIP := blockNet.IP

	// For each subnet, find the gap before it
	for _, subnet := range subnets {
		_, subnetNet, _ := net.ParseCIDR(subnet.CIDR)

		// If there's space before the current subnet
		if bytes.Compare(currentIP, subnetNet.IP) < 0 {
			// Add available CIDRs in the gap
			gapCIDRs := calculateCIDRsInRange(currentIP, subnetNet.IP, blockSize)
			availableCIDRs = append(availableCIDRs, gapCIDRs...)
		}

		// Move current pointer to after this subnet
		currentIP = nextIP(lastIP(subnetNet), subnetNet.Mask)
	}

	// Check for space after the last subnet
	if bytes.Compare(currentIP, lastIP(blockNet)) < 0 {
		gapCIDRs := calculateCIDRsInRange(currentIP, nextIP(lastIP(blockNet), blockNet.Mask), blockSize)
		availableCIDRs = append(availableCIDRs, gapCIDRs...)
	}

	return availableCIDRs
}

// calculateCIDRsInRange calculates the largest possible CIDR blocks in the given IP range
func calculateCIDRsInRange(start, end net.IP, maxPrefix int) []string {
	var cidrs []string
	for bytes.Compare(start, end) < 0 {
		maxSize := maxCIDRSize(start, end, maxPrefix)
		cidr := fmt.Sprintf("%s/%d", start.String(), maxSize)
		cidrs = append(cidrs, cidr)

		// Move to the next IP block
		ones := math.Pow(2, float64(32-maxSize))
		start = nextIPWithStep(start, int(ones))
	}
	return cidrs
}

// maxCIDRSize calculates the maximum CIDR size that can be allocated starting at the given IP
func maxCIDRSize(start, end net.IP, maxPrefix int) int {
	size := 32
	for size > maxPrefix {
		maskLen := 32 - size
		ones := math.Pow(2, float64(maskLen))
		mask := net.CIDRMask(size, 32)

		// Check if the IP is aligned for this mask size
		if !isIPAligned(start, mask) {
			break
		}

		// Check if this size fits within our range
		endIP := nextIPWithStep(start, int(ones)-1)
		if bytes.Compare(endIP, end) >= 0 {
			break
		}

		size--
	}
	return size
}

// isIPAligned checks if an IP address is aligned for the given mask
func isIPAligned(ip net.IP, mask net.IPMask) bool {
	masked := make(net.IP, len(ip))
	copy(masked, ip)
	masked = masked.Mask(mask)
	return ip.Equal(masked)
}

// nextIPWithStep returns the next IP address with a given step size
func nextIPWithStep(ip net.IP, step int) net.IP {
	newIP := make(net.IP, len(ip))
	copy(newIP, ip)

	for i := len(newIP) - 1; i >= 0; i-- {
		sum := int(newIP[i]) + (step % 256)
		newIP[i] = byte(sum % 256)
		step = (step / 256) + (sum / 256)
		if step == 0 {
			break
		}
	}
	return newIP
}
