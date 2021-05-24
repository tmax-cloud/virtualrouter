package rulemanager

import "github.com/cho4036/virtualrouter/ruleManager/natRuleManager"

type Interface interface {
	NAT() natRuleManager.NatRuleManager
}

type rulemanager struct {
	nat natRuleManager.NatRuleManager
}

func New() Interface {
	return &rulemanager{
		nat: natRuleManager.New(),
	}
}

func (n *rulemanager) NAT() natRuleManager.NatRuleManager {
	return n.nat
}
