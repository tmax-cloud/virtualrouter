package k8s2iptables

import (
	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
)

type Interface interface {
	rule.Rule

	SetSpec(interface{})
	SetStatus(interface{})
}

type Object struct {
	Spec   interface{}
	Status interface{}
}

func (o *Object) GetRule() rule.Rule {
	return nil
}

func (o *Object) SetSpec(spec interface{}) {
	o.Spec = spec
}

func (o *Object) SetStatus(status interface{}) {
	o.Status = status
}
