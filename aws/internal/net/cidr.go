package net

import (
	"net"
)

// CIDRBlocksEqual returns whether or not two CIDR blocks are equal:
// - Both CIDR blocks parse to an IP address and network
// - The string representation of the IP addresses are equal
// - The string representation of the networks are equal
// This function is especially useful for IPv6 CIDR blocks which have multiple valid representations.
func CIDRBlocksEqual(cidr1, cidr2 string) bool {
	ip1, ipnet1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return false
	}
	ip2, ipnet2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return false
	}

	return ip2.String() == ip1.String() && ipnet2.String() == ipnet1.String()
}

// CanonicalCIDRBlock returns the canonical representation of a CIDR block.
// This function is especially useful for hash functions for sets which include IPv6 CIDR blocks.
func CanonicalCIDRBlock(cidr string) string {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return cidr
	}

	return ipnet.String()
}
