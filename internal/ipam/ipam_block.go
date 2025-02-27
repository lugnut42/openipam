package ipam

import (
	"net"
)

// Block represents an IP block
type Block struct {
	CIDR        string   `yaml:"cidr"`
	Description string   `yaml:"description"`
	Subnets     []Subnet `yaml:"subnets"`
	
	// Stats are calculated at runtime, not stored in YAML
	Stats *UtilizationStats `yaml:"-"`
}

// UtilizationStats represents runtime utilization statistics
type UtilizationStats struct {
	TotalIPs     uint64
	AllocatedIPs uint64
	AvailableIPs uint64
	Utilization  float64
}

// Subnet represents a subnet within a block
type Subnet struct {
	CIDR   string `yaml:"cidr"`
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
}

// Helper functions
func nextIP(ip net.IP, mask net.IPMask) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}
	return next.Mask(mask)
}

func lastIP(network *net.IPNet) net.IP {
	ip := make(net.IP, len(network.IP))
	copy(ip, network.IP)
	for i := range ip {
		ip[i] |= ^network.Mask[i]
	}
	return ip
}
