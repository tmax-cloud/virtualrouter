package rulemanager

import (
	"github.com/cho4036/virtualrouter/component/rulemanager/parser"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule/rulelister"
)

type Interface interface {
	Lister() rulelister.RuleLister
}

type Rulemanager struct {
	List  rulelister.RuleLister
	Parse parser.Interface
}
