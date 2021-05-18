package k8sNatRule

import (
	"reflect"

	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	"github.com/cho4036/virtualrouter/rule/natrule"
)

var InputType = reflect.TypeOf(v1.NATRule{})

//ToDo: Memory Optimizaion needed
var ParseFunc = func(i interface{}) (string, []natrule.NatRule) {
	var rules []natrule.NatRule
	rules = make([]natrule.NatRule, 0)

	item := i.(v1.NATRule)
	key := item.GetNamespace() + item.GetName()

	var rule natrule.NatRule
	for _, val := range item.Spec.Rules {
		rule.Match.SrcIP = val.Match.SrcIP
		rule.Match.DstIP = val.Match.DstIP
		rule.Match.Protocl = val.Match.Protocl

		rule.Action.SrcIP = val.Action.SrcIP
		rule.Action.DstIP = val.Action.DstIP
		rules = append(rules, rule)
	}

	return key, rules
}
