package rulemanager_test

import (
	"testing"

	"github.com/cho4036/virtualrouter/component/rulemanager"
)

type testObject struct {
	Spec
}

type NATRuleSpec struct {
	SrcIP string
	DstIP string
}

type NATRuleStatus struct {
	Deployed string
}

type NatRuleLister struct {
}

func TestRulemanager(t *testing.T) {
	ruleManager := rulemanager.New()
	ruleList, err := ruleManager.Lister().List()
	if err != nil {
		t.Error("Error occured when Getting ruleList")
	}
	_ = ruleList
}
