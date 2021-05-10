package natrulemanager

import (
	"reflect"

	"github.com/cho4036/virtualrouter/nfv/executor"
	"github.com/cho4036/virtualrouter/rulemanager"
)

type natRuleManager struct {
	executor    executor.Executor
	ruleChannel interface{}
}

func (n *natRuleManager) Run(stopch <-chan struct{}) error {
	if reflect.ValueOf(n.executor).Kind() == reflect.TypeOf(&executor.IptablesExecutor{}).Kind() {
		return n.runIptables(stopch)
	}
	return nil
}

func (n *natRuleManager) SetExecutor() {
}

func NewRuleManager(cfg rulemanager.RuleManagerConfig) *natRuleManager {
	rulemanager := &natRuleManager{
		executor:    executor.NewExecutor(cfg.Executor),
		ruleChannel: cfg.Rulechannel,
	}

	// switch reflect.TypeOf(cfg.Rulechannel).Kind() {
	// case reflect.Chan:
	// 	n.ruleChannel = reflect.ValueOf(cfg.Rulechannel)
	// }
	// switch cfg.Executor {
	// case "iptables":
	// 	// setExecutor(nil)
	// case "":
	// 	// return fmt.Errorf("Executor is not specified")
	// }
	return rulemanager
}

func (n *natRuleManager) runIptables(stopch <-chan struct{}) error {
	for !(len(stopch) > 0) {

	}
	return nil
}
