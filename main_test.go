package main_test

import (
	"fmt"
	"testing"

	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
	"github.com/cho4036/virtualrouter/nfv/natrulemanager"
	"github.com/cho4036/virtualrouter/nfv/natrulemanager/natrule"
)

func TestMain(t *testing.T) {
	ruleManager := natrulemanager.New()
	rule := &natrule.NatRule{
		Metadata: rule.RuleMeta{
			Name: "testRule",
		},
		Item: []natrule.MatchNaction{
			natrule.MatchNaction{
				Match: natrule.Match{
					SrcIP:    "1.1.1.1",
					DstIP:    "2.2.2.2",
					Protocol: "tcp",
				},
				Action: "NAT",
			},
		},
	}

	ruleManager.Lister().Add(rule, true)
	fmt.Println(ruleManager.Lister().List()[0])
}
