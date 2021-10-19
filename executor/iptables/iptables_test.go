package iptables_test

import (
	"bytes"
	"testing"

	"github.com/cho4036/virtualrouter/executor/iptables"
	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
)

func TestIPTables(t *testing.T) {
	i := iptables.NewIPV4()
	buffer := new(bytes.Buffer)
	i.SaveInto("nat", buffer)

	t.Log(buffer.String())
}

func TestIPTablesAddRuleWithoutFlush(t *testing.T) {
	i := iptables.NewIPV4()
	buffer := new(bytes.Buffer)
	exist, err := i.EnsureChain("nat", "test")
	if err != nil {
		t.Log(err)
	}
	if exist {
		// i.SaveInto("nat", buffer)
	}
	i.SaveInto("nat", buffer)

	// if testSNATRule.Spec.Rules[0].Action.DstIP != "" {
	// 	_, _ = i.EnsureChain("nat", "testprerouting")
	// 	var match string
	// 	if testSNATRule.Spec.Rules[0].Match.DstIP != "" {
	// 		match = ""
	// 	}
	// 	writeLine(buffer, "-A", "testprerouting", "-")
	// }

	// t.Log(buffer.String())
	iptables.NF_ADD(testSNATRule.Spec.Rules[0].Match, testSNATRule.Spec.Rules[0].Action, "test", buffer)
	t.Log(buffer.String())
}

func TestIPTablesAddRuleWithFlush(t *testing.T) {
	i := iptables.NewIPV4()
	buffer := new(bytes.Buffer)
	exist, err := i.EnsureChain("nat", "test")
	if err != nil {
		t.Log(err)
	}
	if exist {
		i.FlushChain("nat", "test")
	}
	i.SaveInto("nat", buffer)

	t.Log(buffer.String())

}

var testSNATRule *v1.NATRule = &v1.NATRule{
	Spec: v1.NATRuleSpec{
		Rules: []v1.Rules{
			v1.Rules{
				Match: v1.Match{
					SrcIP:    "5.5.5.5",
					DstIP:    "4.4.4.4",
					Protocol: "tcp",
				},
				Action: v1.Action{
					SrcIP: "3.3.3.3",
				},
			},
		},
	},
}

var testDNATRule *v1.NATRule = &v1.NATRule{
	Spec: v1.NATRuleSpec{
		Rules: []v1.Rules{
			v1.Rules{
				Match: v1.Match{
					SrcIP:    "5.5.5.5",
					DstIP:    "4.4.4.4",
					Protocol: "tcp",
				},
				Action: v1.Action{
					DstIP: "3.3.3.3",
				},
			},
		},
	},
}

func writeLine(buf *bytes.Buffer, words ...string) {
	// We avoid strings.Join for performance reasons.
	for i := range words {
		buf.WriteString(words[i])
		if i < len(words)-1 {
			buf.WriteByte(' ')
		} else {
			buf.WriteByte('\n')
		}
	}
}
