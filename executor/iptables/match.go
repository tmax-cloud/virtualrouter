package iptables

import "net"

type Match struct {
	Source       net.IP `json:"source,omitempty"`
	Destination  net.IP `json:"destination,omitempty"`
	InInterface  string `json:"in-interface,omitempty"`
	OutInterface string `json:"out-interface,omitempty"`
	extensions
}

type extensions struct {
	Conntrack conntrack
	Tcp       tcp    `json:"--protocol tcp,omitempty"`
	Udp       udp    `json:"--protocol udp,omitempty"`
	Mark      string `json:"--mark,omitempty"`
}

type CTSTATE string

const (
	CTSTATE_NEW         CTSTATE = "NEW"
	CTSTATE_ESTABLISHED CTSTATE = "ESTABLISHED"
	CTSTATE_RELATED     CTSTATE = "RELATED"
)

type conntrack struct {
	State CTSTATE `json:"--ctstate,omitempty"`
}

type tcp struct {
	Source      uint16   `json:"--source-port,omitempty"`
	Destination uint16   `json:"--destination-port,omitempty"`
	Flags       []string `json:"--tcp-flags,omitempty"`
}

type udp struct {
	Source      uint16 `json:"--source-port,omitempty"`
	Destination uint16 `json:"--destination-port,omitempty"`
}
