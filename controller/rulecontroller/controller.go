package rulecontroller

import (
	"github.com/cho4036/virtualrouter/executor"
	rulemanager "github.com/cho4036/virtualrouter/ruleManager"
)

type Controller interface {
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
