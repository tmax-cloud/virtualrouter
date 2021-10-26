package iptablescontroller

import (
	"strings"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

func (n *Iptablescontroller) OnFirewallAdd(firewallrule *v1.FireWallRule) error {
	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	var chainName string
	oldRules, ok := n.firewallruleMap[key]
	if ok {
		//n.natruleSynced = false
		klog.Warningf("Duplicated key(%s) detected During OnFirewallAdd Event. Going to overwrite rule : %+v to %+v", key, oldRules, firewallrule)
		for _, rule := range oldRules.Spec.Rules {
			chainName = string(filterForwardfwruleChain)
			n.removeRule(&rule, chainName, &lines)
		}
	}

	for _, rule := range firewallrule.Spec.Rules {
		chainName = string(filterForwardfwruleChain)
		n.appendRule(&rule, chainName, &lines)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("filter", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.firewallruleMap[key] = *firewallrule
	return nil
}

func (n *Iptablescontroller) OnFirewallDelete(firewallrule *v1.FireWallRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onDelete Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	val, ok := n.firewallruleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnFirewallDelete Event", key)
		return nil
	}

	var chainName string
	for _, rule := range val.Spec.Rules {
		chainName = string(filterForwardfwruleChain)
		n.removeRule(&rule, chainName, &lines)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("filter", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}
	delete(n.firewallruleMap, key)
	return nil
}

func (n *Iptablescontroller) OnFirewallUpdate(firewallrule *v1.FireWallRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onUpdate Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	// for k, v := range n.natruleMap {
	// 	klog.Infof("key: %s, value: %+v", k, v)
	// }

	var chainName string
	val, ok := n.firewallruleMap[key]
	if !ok {
		n.firewallruleSynced = false
		klog.Warningf("Updating empty value on key(%s) detected During OnFirewallUpdate Event", key)
	} else {
		for _, rule := range val.Spec.Rules {
			chainName = string(filterForwardfwruleChain)
			n.removeRule(&rule, chainName, &lines)
		}
	}

	for _, rule := range firewallrule.Spec.Rules {
		chainName = string(filterForwardfwruleChain)
		n.appendRule(&rule, chainName, &lines)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)

	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("filter", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.firewallruleMap[key] = *firewallrule
	return nil
}
