package iptables

import (
	k8sIptables "k8s.io/kubernetes/pkg/util/iptables"
	k8sexec "k8s.io/utils/exec"
)

type Interface interface {
	k8sIptables.Interface
}

func New(protocol string) Interface {
	if protocol == string(k8sIptables.ProtocolIPv4) {
		return k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv4)
	} else if protocol == string(k8sIptables.ProtocolIPv6) {
		return k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv6)
	}
	return nil
}
