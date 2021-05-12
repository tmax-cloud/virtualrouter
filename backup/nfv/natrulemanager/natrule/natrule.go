package natrule

import "github.com/cho4036/virtualrouter/component/rulemanager/rule"

type NatRule struct {
	Metadata rule.RuleMata

	Match  Match
	Action string
}

type Match struct {
	SrcIP    string
	DstIP    string
	Protocol string
}

func New() rule.Rule {
	return &NatRule{}
}

func (n *NatRule) GetMetadata() rule.RuleMata {
	return n.Metadata
}
