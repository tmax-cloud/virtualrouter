package executor

import (
	"k8s.io/kubernetes/pkg/util/iptables"
)

type ExecuteProgram string

const (
	IPtables  ExecuteProgram = "iptables"
	IPtables6 ExecuteProgram = "iptables6"
)

type Executor interface {
	IPtables() iptables.Interface
	IPtablesV6() iptables.Interface
}

type executor struct {
}

func NewExecutor(executor ExecuteProgram) Executor {
	if executor == IPtables {
		return &IptablesExecutor{}
	} else if executor == IPtables6 {
		return &IptablesExecutor{}
	}
	return nil
}

// func (executor) IPtables() executors.Iptables {
// 	return &IptablesExecutor{}
// }

type IptablesExecutor struct{}

func (IptablesExecutor) IPtables() iptables.Interface

type IptablesV6Executor struct{}

func (IptablesExecutor) IPtablesV6() iptables.Interface

type Cmd interface{}
