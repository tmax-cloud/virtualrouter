package rulemanager

import (
	"github.com/cho4036/virtualrouter/component/rulemanager/parser.go"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
)

type Interface interface {
	Lister() rule.RuleLister
	Parser() parser.Parser
}

type Rulemanager struct {
	List  rule.RuleLister
	Parse parser.Parser
}

func New() *Rulemanager {
	return &Rulemanager{}
}

func (r *Rulemanager) Lister() rule.RuleLister {
	return nil
}

func (r *Rulemanager) Parser() parser.Parser {
	return nil
}
