package executor

import "github.com/cho4036/virtualrouter/executor/iptables"

type Interface interface {
	IPTABLESV4() iptables.IPV4Interface
	IPTABLESV6() iptables.IPV6Interface
	AddExecutor(interface{})
}

type executor struct {
	iptablesV4 iptables.IPV4Interface
	iptablesV6 iptables.IPV6Interface
}

func New() Interface {
	return &executor{}
}

func (e *executor) IPTABLESV4() iptables.IPV4Interface {
	return e.iptablesV4
}

func (e *executor) IPTABLESV6() iptables.IPV6Interface {
	return e.iptablesV6
}

func (e *executor) AddExecutor(runner interface{}) {
	switch runner.(type) {
	case iptables.IPV4Interface:
		e.iptablesV4 = runner.(iptables.IPV4Interface)
	case iptables.IPV6Interface:
		e.iptablesV6 = runner.(iptables.IPV6Interface)
	}
}
