package natrulemanager_test

import (
	"fmt"
	"testing"

	"github.com/cho4036/virtualrouter/backup/nfv/natrulemanager"
	"github.com/cho4036/virtualrouter/backup/nfv/natrulemanager/natrule"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
)

func TestCheck(t *testing.T) {
	n := natrulemanager.New()
	n.Lister().Add(&natrule.NatRule{
		Metadata: rule.RuleMata{
			Name: "testRule",
		},
		Match: natrule.Match{
			SrcIP:    "1",
			DstIP:    "2",
			Protocol: "tcp",
		},
		Action: "NAT",
	}, false)
	fmt.Println("start")
	fmt.Println(n.Lister().List()[0].GetMetadata().Name)
}
