package main

import (
	"github.com/go-errors/errors"
	"github.com/segmentio/go-prompt"
	"net"
)

func getExternalIPAddr() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	ipAddrs := []string{}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ipAddrs = append(ipAddrs, ip.String())
		}
	}
	if len(ipAddrs) == 0 {
		return "", errors.New("Unable to determine external interface address. Are you connected to the network?")
	}

	i := prompt.Choose("Select External IP Address", ipAddrs)
	return ipAddrs[i], nil
}
