package database

import (
	"fmt"
	"net"
)

func IPv4ToInt(ipStr string) (int64, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid ip: %s", ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("not ipv4: %s", ipStr)
	}
	return ipv4ToInt(ip), nil
}

func CIDRRange(cidr string) (start, end int64, err error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, 0, err
	}

	ip := ipNet.IP.To4()
	if ip == nil {
		return 0, 0, fmt.Errorf("not ipv4 cidr: %s", cidr)
	}

	start = ipv4ToInt(ip)
	mask := ipNet.Mask
	broadcast := make(net.IP, 4)
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}
	end = ipv4ToInt(broadcast)
	return start, end, nil
}

func ipv4ToInt(ip net.IP) int64 {
	return int64(uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3]))
}
