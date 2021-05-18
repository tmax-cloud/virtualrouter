package rulemanager_test

import (
	"fmt"
	"testing"

	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	"github.com/cho4036/virtualrouter/ruleManager/ruleManagerFactory"
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRuleManager(t *testing.T) {
	f := ruleManagerFactory.New().NatRuleManager()
	err := f.Add(v1.NATRule{
		ObjectMeta: k8sv1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.NATRuleSpec{
			Rules: []v1.Rules{
				v1.Rules{
					Match: v1.Match{
						SrcIP:   "1",
						DstIP:   "2",
						Protocl: "tcp",
					},
					Action: v1.Action{
						SrcIP: "3",
						DstIP: "4",
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range f.List() {
		fmt.Println(val)
	}

	err = f.Update(v1.NATRule{
		ObjectMeta: k8sv1.ObjectMeta{
			Name: "test",
		},
		Spec: v1.NATRuleSpec{
			Rules: []v1.Rules{
				v1.Rules{
					Match: v1.Match{
						SrcIP:   "2",
						DstIP:   "2",
						Protocl: "tcp",
					},
					Action: v1.Action{
						SrcIP: "3",
						DstIP: "4",
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, val := range f.List() {
		fmt.Println(val)
	}
}
