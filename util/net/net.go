package net

import "net"

func IPValidation(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	return true
}
