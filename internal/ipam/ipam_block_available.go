package ipam

import (
	"bytes"
	"fmt"
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

func calculateAvailableCIDRs(block *Block) []string {
	var availableCIDRs []string
	_, blockNet, _ := net.ParseCIDR(block.CIDR)

	// Sort subnets by starting IP
	subnets := block.Subnets
	sort.Slice(subnets, func(i, j int) bool {
		_, subnetNetI, _ := net.ParseCIDR(subnets[i].CIDR)
		_, subnetNetJ, _ := net.ParseCIDR(subnets[j].CIDR)
		return bytes.Compare(subnetNetI.IP, subnetNetJ.IP) < 0
	})

	// Find gaps between subnets
	prevIP := blockNet.IP
	for _, subnet := range subnets {
		_, subnetNet, _ := net.ParseCIDR(subnet.CIDR)
		if bytes.Compare(prevIP, subnetNet.IP) < 0 {
			availableCIDRs = append(availableCIDRs, fmt.Sprintf("%s - %s", prevIP, subnetNet.IP))
		}
		prevIP = nextIP(lastIP(subnetNet), subnetNet.Mask)
	}

	// Check for space after the last subnet
	if blockNet.Contains(prevIP) {
		availableCIDRs = append(availableCIDRs, fmt.Sprintf("%s - %s", prevIP, lastIP(blockNet)))
	}

	return availableCIDRs
}
