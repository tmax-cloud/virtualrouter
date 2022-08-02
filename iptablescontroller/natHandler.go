package iptablescontroller

import (
	"fmt"
	"strings"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

type NATRULEVALIDATION bool

const (
	NATRULE_INVALIDE NATRULEVALIDATION = false
	NATRULE_VALID    NATRULEVALIDATION = true
)

func (n *Iptablescontroller) OnNATAdd(natrule *v1.NATRule) error {
	if natruleValidationCheck(natrule) == NATRULE_INVALIDE {
		return fmt.Errorf("NAT Rule is invalid")
	}

	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	key := natrule.GetNamespace() + natrule.GetName()
	_, ok := n.natruleMap[key]
	if ok {
		return fmt.Errorf("The NAT Rule CR namespace-name exists: %s", key)
	}

	for _, rule := range natrule.Spec.Rules {
		var lines []string
		n.appendRule(&rule, "", &lines) // Dummy chain name
		lines = strings.Split(lines[0], "\n")
		args := strings.Split(lines[0], " ")
		args = args[1:]
		wrongFormat := false
		var chainName iptables.ChainName

		// Setting the appropriate chain name
		if rule.Action.SrcIP == "0.0.0.0" { //In case CR is for Masquerade
			chainName = natPostroutingSNATChain
		} else if rule.Action.SrcIP != "" { //In case CR is for static SNAT
			chainName = natPostroutingStaticNATChain
			if err := setRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "setRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" { //In case CR is for static DNAT
			chainName = natPreroutingStaticNATChain
		} else {
			klog.Errorln("Wrong Format of rules")
			wrongFormat = true
		}
		if !wrongFormat {
			// Apply the rule through the iptables command
			if _, err := n.iptables.EnsureRule(iptables.Append, iptables.TableNAT, chainName, args...); err != nil {
				klog.ErrorS(err, "Failed to  add NAT rule : chain:", string(chainName), args)
			}
			klog.InfoS("NAT rule added. chain: ", string(chainName), args)
		}

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
	if natruleValidationCheck(natrule) == NATRULE_INVALIDE {
		return fmt.Errorf("NAT Rule is invalid")
	}

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

func natruleValidationCheck(natRule *v1.NATRule) NATRULEVALIDATION {
	// validate natRule instance
	for _, rule := range natRule.Spec.Rules {
		if rule.Match.DstPort != 0 || rule.Match.SrcPort != 0 {
			if rule.Match.Protocol == "" {
				return NATRULE_INVALIDE
			}
		}

		if rule.Match.DstPort < 0 || rule.Match.DstPort > 65535 || rule.Match.SrcPort < 0 || rule.Match.SrcPort > 65535 {
			return NATRULE_INVALIDE
		}

		if rule.Action.DstPort < 0 || rule.Action.DstPort > 65535 || rule.Action.SrcPort < 0 || rule.Action.SrcPort > 65535 {
			return NATRULE_INVALIDE
		}
	}

	return NATRULE_VALID
}
