package rulemanagerfactory

import "github.com/cho4036/virtualrouter/nfv/nat/natrulemanager"

type RuleManagerFactory interface {
	NatRuleManager() natrulemanager.NatRuleManager
}

type ruleManagerFactory struct{}

func NewRuleManager() RuleManagerFactory {
	return &ruleManagerFactory{}
}
