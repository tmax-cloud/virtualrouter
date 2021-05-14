package ruleManagerFactory

import "github.com/cho4036/virtualrouter/rule-manager/natRuleManager"

type RuleManagerFactory interface {
	NatRuleManager() natRuleManager.NatRuleManager
}

type ruleManagerFactory struct {
}

func New() RuleManagerFactory {
	return &ruleManagerFactory{}
}

func (r *ruleManagerFactory) NatRuleManager() natRuleManager.NatRuleManager {
	return natRuleManager.New()
}
