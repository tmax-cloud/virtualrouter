package executorFactory

import (
	"sync"

	"github.com/cho4036/virtualrouter/executor/iptables"
)

type ExecutorFactory interface {
	IPTABLESV4() iptables.IPV4Interface
	IPTABLESV6() iptables.IPV6Interface
}

type executorFactory struct {
}

var factory *executorFactory
var once sync.Once

func New() ExecutorFactory {
	once.Do(func() {
		factory = &executorFactory{}
	})
	return factory
}

func (e *executorFactory) IPTABLESV4() iptables.IPV4Interface {
	return iptables.NewIPV4()
}

func (e *executorFactory) IPTABLESV6() iptables.IPV6Interface {
	return iptables.NewIPV6()
}
