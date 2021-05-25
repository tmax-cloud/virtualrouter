package rulecontroller

import (
	"github.com/cho4036/virtualrouter/executor"
	rulemanager "github.com/cho4036/virtualrouter/ruleManager"
)

type Controller interface {
	RuleManager() rulemanager.Interface
	Executor() executor.Interface
}

type controller struct {
	ruleManager rulemanager.Interface
	executor    executor.Interface
}

func New() Controller {
	return &controller{
		ruleManager: rulemanager.New(),
		executor:    executor.New(),
	}
}

func (c *controller) RuleManager() rulemanager.Interface {
	return c.ruleManager
}

func (c *controller) Executor() executor.Interface {
	return c.executor
}
