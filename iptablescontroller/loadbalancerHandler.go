package iptablescontroller

import (
	"fmt"
	"strings"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

func (n *Iptablescontroller) OnLoadbalanceAdd(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()
	var chainName string
	oldLBRules, ok := n.loadbalancerruleMap[key]
	if ok {
		//n.natruleSynced = false
		oldRules := transLBRule2Rule(oldLBRules)
		klog.Warningf("Duplicated key(%s) detected During LoadbalanceAdd Event. Going to overwrite rule : %+v to %+v", key, oldRules, loadbalancerrule)
		for _, rule := range oldRules {
			chainName = string(natPreroutingLoadBalanceChain)
			n.removeRule(rule, chainName, &lines, rule.Args...)
			if err := delRouteForProxyARP(rule.Match.DstIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		}
	}

	rules := transLBRule2Rule(*loadbalancerrule)
	for _, rule := range rules {
		chainName = string(natPreroutingLoadBalanceChain)
		n.appendRule(rule, chainName, &lines, rule.Args...)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	for i := range rules {
		if err := setRouteForProxyARP(rules[i].Match.DstIP); err != nil {
			klog.ErrorS(err, "setRouteForProxyARP")
			return err
		}
	}

	n.loadbalancerruleMap[key] = *loadbalancerrule
	return nil
}

func (n *Iptablescontroller) OnLoadbalanceDelete(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onDelete Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()
	val, ok := n.loadbalancerruleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnLoadbalanceDelete Event", key)
		return nil
	}

	var chainName string
	oldRules := transLBRule2Rule(val)
	for _, rule := range oldRules {
		chainName = string(natPreroutingLoadBalanceChain)
		n.removeRule(rule, chainName, &lines, rule.Args...)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	for _, rule := range oldRules {
		if err := delRouteForProxyARP(rule.Match.DstIP); err != nil {
			klog.ErrorS(err, "delRouteForProxyARP")
			return err
		}
	}

	delete(n.loadbalancerruleMap, key)
	return nil
}

func (n *Iptablescontroller) OnLoadbalanceUpdate(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onUpdate Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()
	// for k, v := range n.natruleMap {
	// 	klog.Infof("key: %s, value: %+v", k, v)
	// }

	var chainName string
	var oldRules []*v1.Rules
	val, ok := n.loadbalancerruleMap[key]
	if !ok {
		n.loadbalancerruleSynced = false
		klog.Warningf("Updating empty value on key(%s) detected During OnLoadbalanceUpdate Event", key)
	} else {
		oldRules := transLBRule2Rule(val)
		for _, rule := range oldRules {
			chainName = string(natPreroutingLoadBalanceChain)
			n.removeRule(rule, chainName, &lines, rule.Args...)
		}
		delete(n.loadbalancerruleMap, key)
	}

	rules := transLBRule2Rule(*loadbalancerrule)
	for _, rule := range rules {
		chainName = string(natPreroutingLoadBalanceChain)
		n.appendRule(rule, chainName, &lines, rule.Args...)
	}

	for _, rule := range oldRules {
		if err := delRouteForProxyARP(rule.Match.DstIP); err != nil {
			klog.ErrorS(err, "delRouteForProxyARP")
			return err
		}
	}
	for _, rule := range rules {
		if err := setRouteForProxyARP(rule.Match.DstIP); err != nil {
			klog.ErrorS(err, "delRouteForProxyARP")
			return err
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

	n.loadbalancerruleMap[key] = *loadbalancerrule
	return nil
}

func transLBRule2Rule(lbrule v1.LoadBalancerRule) []*v1.Rules {
	var rules []*v1.Rules

	for i := range lbrule.Spec.Rules {
		for j := range lbrule.Spec.Rules[i].BackendIPs {
			weight := float64(lbrule.Spec.Rules[i].BackendIPs[j].Weight) * 0.01
			rule := &v1.Rules{
				Match: v1.Match{
					DstIP: lbrule.Spec.Rules[i].LoadBalancerIP,
				},
				Action: v1.Action{
					DstIP: lbrule.Spec.Rules[i].BackendIPs[j].BackendIP,
				},
				Args: []string{
					// "-m statistic --mode random --probability " + strconv.FormatFloat(weight, 'e', -1, 64),
					"-m statistic --mode random --probability " + fmt.Sprintf("%.2f", weight),
				},
			}
			rules = append(rules, rule)
		}
	}

	return rules
}
