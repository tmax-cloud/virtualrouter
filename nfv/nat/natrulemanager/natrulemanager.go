package natrulemanager

import (
	"github.com/cho4036/virtualrouter/component/rulemanager"
	"github.com/cho4036/virtualrouter/component/rulemanager/parser"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule/rulelister"
)

type NatRuleManager interface {
}

type natRuleManager struct {
	rulemanager.Rulemanager
}

func (n *natRuleManager) Lister() rulelister.RuleLister {
	return n.List
}

func New() rulemanager.Interface {
	return &natRuleManager{
		Rulemanager: rulemanager.Rulemanager{
			List:  rulelister.New(),
			Parse: parser.New(),
		},
	}
}
