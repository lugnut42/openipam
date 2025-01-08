package ipam

import (
	"net"
)

// Helper functions
func isSubnetOverlapping(subnets []Subnet, newSubnet *net.IPNet) bool {
	for _, subnet := range subnets {
		_, existingSubnetNet, _ := net.ParseCIDR(subnet.CIDR)
		if newSubnet.Contains(existingSubnetNet.IP) || existingSubnetNet.Contains(newSubnet.IP) {
			return true
		}
	}
	return false
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
