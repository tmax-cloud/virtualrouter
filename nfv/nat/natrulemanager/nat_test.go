package natrulemanager_test

import (
	"fmt"
	"testing"

	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
	"github.com/cho4036/virtualrouter/nfv/nat/natrule"
	"github.com/cho4036/virtualrouter/nfv/nat/natrulemanager"
)

func TestCheck(t *testing.T) {
	n := natrulemanager.New()
	n.Lister().Add(&natrule.NatRule{
		Metadata: rule.RuleMeta{
			Name: "testRule",
		},
		Item: []natrule.MatchNaction{
			natrule.MatchNaction{
				Match: natrule.Match{
					SrcIP:    "1",
					DstIP:    "2",
					Protocol: "tcp",
				},
				Action: "NAT",
			},
			natrule.MatchNaction{
				Match: natrule.Match{
					SrcIP:    "1",
					DstIP:    "2",
					Protocol: "tcp",
				},
				Action: "NAT",
			},
		},
	}, false)
	n.Lister().Add(&natrule.NatRule{
		Metadata: rule.RuleMeta{
			Name: "testRule2",
		},
		Item: []natrule.MatchNaction{
			natrule.MatchNaction{
				Match: natrule.Match{
					SrcIP:    "1",
					DstIP:    "2",
					Protocol: "tcp",
				},
				Action: "NAT",
			},
			natrule.MatchNaction{
				Match: natrule.Match{
					SrcIP:    "1",
					DstIP:    "2",
					Protocol: "tcp",
				},
				Action: "NAT",
			},
		},
	}, false)
	fmt.Println("start")
	for _, val := range n.Lister().List() {
		fmt.Println(val)
	}
}
