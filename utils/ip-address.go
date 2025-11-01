package utils

import (
	"fmt"
	"net"
)

// GetIPFromDomain returns all IP addresses for a given domain as strings
func GetIPFromDomain(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("could not get IPs for %s: %v", domain, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no IP addresses found for %s", domain)
	}

	// Convert net.IP to string
	ipStrings := make([]string, 0, len(ips))
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	return ipStrings, nil
}
