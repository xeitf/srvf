package net

import (
	"net"
)

// LookupIP
func LookupIP(host string) (IPs []string) {
	vv, _ := net.LookupIP(host)
	if len(IPs) <= 0 {
		return
	}
	for _, v := range vv {
		IPs = append(IPs, v.String())
	}
	return
}

// LookupFirstIP
func LookupFirstIP(host string) (IP string) {
	if IPs := LookupIP(host); len(IPs) <= 0 {
		return
	} else {
		return IPs[0]
	}
}
