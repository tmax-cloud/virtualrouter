package natrule

import "github.com/cho4036/virtualrouter/component/rulemanager/rule"

type NatRule struct {
	Metadata rule.RuleMeta
	Item     []MatchNaction
}

type Match struct {
	SrcIP    string
	DstIP    string
	Protocol string
}

type MatchNaction struct {
	Match
	Action string
}

func New() rule.Rule {
	return &NatRule{}
}

func (n *NatRule) GetMetadata() rule.RuleMeta {
	return n.Metadata
}

func (n *NatRule) GetItems() interface{} {
	return n.Item
}
