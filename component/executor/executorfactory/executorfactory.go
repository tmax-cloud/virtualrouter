package executor

import (
	"github.com/cho4036/virtualrouter/component/rulemanager/rule/rulelister"
	"github.com/cho4036/virtualrouter/executor/iptables"
)

type ExecutorFactory interface {
	IPtables(ruleList rulelister.RuleLister) iptables.Interface
}

type executorFactory struct{}

func NewExecutor() ExecutorFactory {
	return &executorFactory{}
}

func (e *executorFactory) IPtables(ruleList rulelister.RuleLister) iptables.Interface {
	return iptables.New(ruleList)
}
