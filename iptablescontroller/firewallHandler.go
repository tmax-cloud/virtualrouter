package iptablescontroller

import (
	"fmt"
	"strings"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

type FWRULEVALIDATION bool

const (
	FWRULE_INVALIDE FWRULEVALIDATION = false
	FWRULE_VALID    FWRULEVALIDATION = true
)

func (n *Iptablescontroller) OnFirewallAdd(firewallrule *v1.FireWallRule) error {
	if fwruleValidationCheck(firewallrule) == FWRULE_INVALIDE {
		return fmt.Errorf("FW Rule is invalid")
	}

	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	for _, rule := range firewallrule.Spec.Rules {
		var lines []string
		n.appendRule(&rule, string(filterForwardfwruleChain), &lines)
		lines = strings.Split(lines[0], "\n")
		args := strings.Split(lines[0], " ")
		args = args[2:]
		if _, err := n.iptables.EnsureRule(iptables.Append, iptables.TableFilter, filterForwardfwruleChain, args...); err != nil {
			klog.ErrorS(err, "Table: ", string(iptables.TableFilter), "chain: ", string(filterForwardfwruleChain), "args: ", args)
			return err
		}
		klog.InfoS("Rule added", "Table: ", string(iptables.TableFilter), "chain:", string(filterForwardfwruleChain), "args: ", args)

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
	if fwruleValidationCheck(firewallrule) == FWRULE_INVALIDE {
		return fmt.Errorf("FW Rule is invalid")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	klog.Infof("onUpdate Called!: %s", key)
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

func fwruleValidationCheck(fwRule *v1.FireWallRule) FWRULEVALIDATION {
	// validate natRule instance
	for _, rule := range fwRule.Spec.Rules {
		if rule.Match.DstPort != 0 || rule.Match.SrcPort != 0 {
			if rule.Match.Protocol == "" || rule.Match.Protocol == PROTOCOL_ALL {
				return FWRULE_INVALIDE
			}
		}
		if rule.Match.DstPort < 0 || rule.Match.DstPort > 65535 || rule.Match.SrcPort < 0 || rule.Match.SrcPort > 65535 {
			return FWRULE_INVALIDE
		}
	}

	return FWRULE_VALID
}
