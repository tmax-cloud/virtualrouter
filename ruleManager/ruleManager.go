package rulemanager

import (
	"github.com/cho4036/virtualrouter/ruleManager/natRuleManager"
)

type Interface interface {
	NAT() natRuleManager.NatRuleManager
	AddRuleMager(interface{})
}

type rulemanager struct {
	nat natRuleManager.NatRuleManager
}

func New() Interface {
	return &rulemanager{}
}

func (n *rulemanager) NAT() natRuleManager.NatRuleManager {
	return n.nat
}

func (n *rulemanager) AddRuleMager(manager interface{}) {
	switch manager.(type) {
	case natRuleManager.NatRuleManager:
		n.nat = manager.(natRuleManager.NatRuleManager)
	default:
	}
}
