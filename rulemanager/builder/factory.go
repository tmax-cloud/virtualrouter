package builder

import (
	"reflect"
	"sync"

	"github.com/cho4036/virtualrouter/nfv/natrulemanager"
	"github.com/cho4036/virtualrouter/rulemanager"
)

type RuleManagerFactory interface {
	Start(stopch <-chan struct{})
	ImportManager(rulemanager.Interface)
	Nat(rulemanager.RuleManagerConfig) rulemanager.Interface
}

type ruleManagerFactory struct {
	lock           sync.Mutex
	managers       map[reflect.Type]rulemanager.Interface
	startedManager map[reflect.Type]bool
}

func NewRuleManagerFactory() RuleManagerFactory {
	return &ruleManagerFactory{}
}

func (r *ruleManagerFactory) Start(stopch <-chan struct{}) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for managerType, manager := range r.managers {
		if !r.startedManager[managerType] {
			go manager.Run(stopch)
			r.startedManager[managerType] = true
		}
	}
}

func (r *ruleManagerFactory) ImportManager(manager rulemanager.Interface) {
	r.lock.Lock()
	defer r.lock.Unlock()

	managerType := reflect.TypeOf(manager)
	r.managers[managerType] = manager
}

func (r *ruleManagerFactory) Nat(cfg rulemanager.RuleManagerConfig) rulemanager.Interface {
	return natrulemanager.NewRuleManager(cfg)
}
