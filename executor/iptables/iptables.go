package iptables

import (
	"time"

	"github.com/cho4036/virtualrouter/component/executor"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule/rulelister"

	"k8s.io/apimachinery/pkg/util/wait"
	k8sIptables "k8s.io/kubernetes/pkg/util/iptables"
)

var _ = k8sIptables.ChainInput

type Interface interface {
	executor.Interface
	Sync()
}

type iptables struct {
	ruleList rulelister.RuleLister
}

func (i *iptables) Start(stopCh <-chan struct{}) error {

	go wait.Until(i.Sync, time.Second, stopCh)
	<-stopCh
	return nil
}

func (i *iptables) Sync() {
	// ruleList := i.ruleList.List()
	// ruleList
}
func New(ruleList rulelister.RuleLister) Interface {
	// if protocol == string(k8sIptables.ProtocolIPv4) {
	// 	return k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv4)
	// } else if protocol == string(k8sIptables.ProtocolIPv6) {
	// 	return k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv6)
	// }
	return &iptables{ruleList: ruleList}
}
