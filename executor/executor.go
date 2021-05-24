package executor

import "github.com/cho4036/virtualrouter/executor/iptables"

type Interface interface {
	IPTABLESV4() iptables.Interface
	IPTABLESV6() iptables.Interface
}

type executor struct {
	iptablesV4 iptables.Interface
	iptablesV6 iptables.Interface
}

func New() Interface {
	return &executor{
		iptablesV4: iptables.New("IPv4"),
		iptablesV6: iptables.New("IPv6"),
	}
}

func (e *executor) IPTABLESV4() iptables.Interface {
	return e.iptablesV4
}

func (e *executor) IPTABLESV6() iptables.Interface {
	return e.iptablesV6
}
