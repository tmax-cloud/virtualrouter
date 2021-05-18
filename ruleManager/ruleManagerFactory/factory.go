package ruleManagerFactory

import (
	"sync"

	"github.com/cho4036/virtualrouter/ruleManager/natRuleManager"
)

type RuleManagerFactory interface {
	NatRuleManager() natRuleManager.NatRuleManager
}

type ruleManagerFactory struct {
}

var factory *ruleManagerFactory
var once sync.Once

func New() RuleManagerFactory {
	once.Do(func() {
		factory = &ruleManagerFactory{}
	})
	return factory
}

func (r *ruleManagerFactory) NatRuleManager() natRuleManager.NatRuleManager {
	return natRuleManager.New()
}
