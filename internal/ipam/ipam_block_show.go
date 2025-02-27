package ipam

import (
	"fmt"
	"net"
	"os"
	"text/tabwriter"

	"github.com/lugnut42/openipam/internal/config"
)

func ShowBlock(cfg *config.Config, cidr, fileKey string) error {
	blockFile, ok := cfg.BlockFiles[fileKey]
	if !ok {
		return fmt.Errorf("block file for key %s not found", fileKey)
	}

	yamlData, err := readYAMLFile(blockFile)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	blocks, err := unmarshalBlocks(yamlData)
	if err != nil {
		return fmt.Errorf("error unmarshalling YAML data: %w", err)
	}

	for i, block := range blocks {
		if block.CIDR == cidr {
			// Calculate utilization stats
			if block.Stats == nil {
				// Calculate the stats
				_, ipNet, _ := net.ParseCIDR(block.CIDR)
				totalIPs := calculateIPCount(ipNet)
				
				var allocatedIPs uint64 = 0
				for _, subnet := range block.Subnets {
					_, subnetNet, err := net.ParseCIDR(subnet.CIDR)
					if err == nil {
						allocatedIPs += calculateIPCount(subnetNet)
					}
				}
				
				utilization := float64(0)
				if totalIPs > 0 {
					utilization = float64(allocatedIPs) / float64(totalIPs) * 100
				}
				
				blocks[i].Stats = &UtilizationStats{
					TotalIPs:     totalIPs,
					AllocatedIPs: allocatedIPs,
					AvailableIPs: totalIPs - allocatedIPs,
					Utilization:  utilization,
				}
			}
			
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Block CIDR\tDescription")
			fmt.Fprintln(w, block.CIDR+"\t"+block.Description)
			
			// Display utilization
			if block.Stats != nil {
				fmt.Fprintln(w, "\nUtilization:")
				fmt.Fprintf(w, "Total IPs:\t%d\n", block.Stats.TotalIPs)
				fmt.Fprintf(w, "Allocated IPs:\t%d\n", block.Stats.AllocatedIPs)
				fmt.Fprintf(w, "Available IPs:\t%d\n", block.Stats.AvailableIPs)
				fmt.Fprintf(w, "Utilization:\t%.2f%%\n", block.Stats.Utilization)
			}
			
			fmt.Fprintln(w, "\nSubnets:")
			fmt.Fprintln(w, "Subnet CIDR\tName\tRegion")
			for _, subnet := range block.Subnets {
				fmt.Fprintln(w, subnet.CIDR+"\t"+subnet.Name+"\t"+subnet.Region)
			}
			w.Flush()
			return nil
		}
	}

	return fmt.Errorf("block with CIDR %s not found", cidr)
}
