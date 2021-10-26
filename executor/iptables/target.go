package iptables

import (
	"net"
	"regexp"
	"strings"

	utilNet "github.com/tmax-cloud/virtualrouter/util/net"
)

type TargetCommand string

const (
	BALANCE    TargetCommand = "BALANCE"
	CONNMARK   TargetCommand = "CONNMARK"
	DNAT       TargetCommand = "DNAT"
	MARK       TargetCommand = "MARK"
	MASQUERADE TargetCommand = "MASQUERADE"
	SNAT       TargetCommand = "SNAT"
)

type Target struct {
	Balance    Balance    `json:"BALANCE,omitempty"`
	ConnMark   ConnMark   `json:"CONNMARK,omitempty"`
	Dnat       Dnat       `json:"DNAT,omitempty"`
	Mark       Mark       `json:"MARK,omitempty"`
	Masquerade Masquerade `json:"MASQUERADE,omitempty"`
	Snat       Snat       `json:"SNAT,omitempty"`
}

type Balance struct {
	Destination net.IP `json:"--to-destination,omitempty"`
}

type ConnMark struct {
	// set-mark mark[/mask]. Set connection mark. If a mask is specified then only those bits set in the mask is modified.
	SetMark string `json:"--set-mark,omitempty"`
	// save-mark [--mask mask]. Copy the netfilter packet mark value to the connection mark. If a mask is specified then only those bits are copied.
	SaveMark string `json:"--save-mark,omitempty"`
	// restore-mark [--mask mask]. Copy the connection mark value to the packet. If a mask is specified then only those bits are copied. This is only valid in the mangle table.
	RestoreMark string `json:"--restore-mark,omitempty"`
}

type Dnat struct {
	// to-destination ipaddr[-ipaddr][:port-port]. which can specify a single new destination IP address, an inclusive range of IP addresses, and optionally, a port range (which is only valid if the rule also specifies -p tcp or -p udp). If no port range is specified, then the destination port will never be modified.
	Destination string `json:"--to-destination,omitempty"`
}

type Mark struct {
	// set-mark mark. This is used to set the netfilter mark value associated with the packet. Range: 2^32-1
	MarkValue uint32 `json:"--set-mark,omitempty"`
}

type Masquerade struct {
	//	to-ports port[-port]. This specifies a range of source ports to use, overriding the default SNAT source port-selection heuristics (see above). This is only valid if the rule also specifies -p tcp or -p udp
	Port string `json:"--to-ports,omitempty"`
}

func (m *Masquerade) EnsurePort() string {
	ports := strings.Split(m.Port, "-")
	re := regexp.MustCompile("^([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$")
	for _, val := range ports {
		if match := re.MatchString(val); !match {
			return ""
		}
	}
	return m.Port
}

type Snat struct {
	//to-source ipaddr[-ipaddr][:port-port]. port range (which is only valid if the rule also specifies -p tcp or -p udp). If no port range is specified, then source ports below 512 will be mapped to other ports below 512: those between 512 and 1023 inclusive will be mapped to ports below 1024, and other ports will be mapped to 1024 or above. Where possible, no port alteration will occur.
	Source string `json:"--to-source,omitempty"`
}

func (s *Snat) EnsureSource() string {
	addr := strings.Split(s.Source, ":")
	if len(addr) == 2 {
		ips := strings.Split(addr[0], "-")
		for _, val := range ips {
			if !utilNet.IPValidation(val) {
				return ""
			}
		}
		ports := strings.Split(addr[1], "-")
		for _, val := range ports {
			if !utilNet.PortValidation(val) {
				return ""
			}
		}
	} else if len(addr) == 1 {
		ips := strings.Split(addr[0], "-")
		for _, val := range ips {
			if !utilNet.IPValidation(val) {
				return ""
			}
		}
	} else {
		return ""
	}
	return s.Source
}
