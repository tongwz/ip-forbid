package test

import (
	"fmt"
	"net"
	"testing"
)

// go test -v -run TestIpIn test/ip_test.go -count=1
func TestIpIn(t *testing.T) {
	ipStr := "103.31.3.1"

	ipRanges := []string{
		"173.245.48.0/20",
		"103.21.244.0/22",
		"103.22.200.0/22",
		"103.31.4.0/22",
		"141.101.64.0/18",
		"108.162.192.0/18",
		"190.93.240.0/20",
		"188.114.96.0/20",
		"197.234.240.0/22",
		"198.41.128.0/17",
		"162.158.0.0/15",
		"104.16.0.0/13",
		"104.24.0.0/14",
		"172.64.0.0/13",
		"131.0.72.0/22",
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Printf("Invalid IP address: %s\n", ipStr)
		return
	}

	for _, rangeStr := range ipRanges {
		_, ipNet, err := net.ParseCIDR(rangeStr)
		if err != nil {
			fmt.Printf("Invalid CIDR notation: %s\n", rangeStr)
			continue
		}
		if ipNet.Contains(ip) {
			fmt.Printf("%s is in the range %s.\n", ipStr, rangeStr)
			return
		}
	}

	fmt.Printf("%s is not in any of the specified ranges.\n", ipStr)
}
