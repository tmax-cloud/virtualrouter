package iptablescontroller

import (
	"strings"

	"github.com/cho4036/virtualrouter/executor/iptables"
	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

func (n *Iptablescontroller) OnNATAdd(natrule *v1.NATRule) error {
	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := natrule.GetNamespace() + natrule.GetName()
	var chainName string
	oldRules, ok := n.natruleMap[key]
	if ok {
		//n.natruleSynced = false
		klog.Warningf("Duplicated key(%s) detected During OnAdd Event. Going to overwrite rule : %+v to %+v", key, oldRules, natrule)
		for _, rule := range oldRules.Spec.Rules {
			if rule.Action.SrcIP == "0.0.0.0" {
				chainName = string(natPostroutingSNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else if rule.Action.SrcIP != "" {
				chainName = string(natPostroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
				if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
					klog.ErrorS(err, "delRouteForProxyARP")
					return err
				}
			} else if rule.Action.DstIP != "" {
				chainName = string(natPreroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else {
				klog.Errorln("Wrong Format of rules")
			}
		}
	}

	for _, rule := range natrule.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
			if err := setRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.natruleMap[key] = *natrule
	return nil
}

func (n *Iptablescontroller) OnNATDelete(natrule *v1.NATRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onDelete Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := natrule.GetNamespace() + natrule.GetName()
	val, ok := n.natruleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnDelete Event", key)
		return nil
	}

	var chainName string
	for _, rule := range val.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.removeRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.removeRule(&rule, chainName, &lines)
			if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.removeRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}
	delete(n.natruleMap, key)
	return nil
}

func (n *Iptablescontroller) OnNATUpdate(natrule *v1.NATRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onUpdate Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := natrule.GetNamespace() + natrule.GetName()
	// for k, v := range n.natruleMap {
	// 	klog.Infof("key: %s, value: %+v", k, v)
	// }

	var chainName string
	val, ok := n.natruleMap[key]
	if !ok {
		n.natruleSynced = false
		klog.Warningf("Updating empty value on key(%s) detected During OnUpdate Event", key)
	} else {
		for _, rule := range val.Spec.Rules {
			if rule.Action.SrcIP == "0.0.0.0" {
				chainName = string(natPostroutingSNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else if rule.Action.SrcIP != "" {
				chainName = string(natPostroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
				if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
					klog.ErrorS(err, "delRouteForProxyARP")
					return err
				}
			} else if rule.Action.DstIP != "" {
				chainName = string(natPreroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else {
				klog.Errorln("Wrong Format of rules")
			}
		}
	}

	for _, rule := range natrule.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
			if err := setRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)

	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.natruleMap[key] = *natrule
	return nil
}
