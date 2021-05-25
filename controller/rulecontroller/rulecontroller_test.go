package rulecontroller_test

import (
	"testing"

	"github.com/cho4036/virtualrouter/controller/rulecontroller"
	"github.com/cho4036/virtualrouter/executor/executorFactory"
	"github.com/cho4036/virtualrouter/ruleManager/ruleManagerFactory"
)

func Test(t *testing.T) {
	c := rulecontroller.New()
	rm_Factory := ruleManagerFactory.New()
	executorFactory := executorFactory.New()
	c.RuleManager().AddRuleMager(rm_Factory.NatRuleManager())
	c.Executor().AddExecutor(executorFactory.IPTABLESV4())
	// fmt.Println(c.Executor().IPTABLESV4().List())
}
