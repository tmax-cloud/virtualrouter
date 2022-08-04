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

		return fmt.Errorf("FW Rule is invalid") //
	}

	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	key := firewallrule.GetNamespace() + firewallrule.GetName()
	// h := sha256.New() //Bang: new Hash instance
	var newFwHashRule FwRuleStruct
	newFwHashRule.ruleHashMap = make(map[v1.Match]v1.Rules)

	for _, rule := range firewallrule.Spec.Rules {
		var buffer []string
		n.appendRule(&rule, "", &buffer)
		// klog.Infoln("Input rule in the buffer:", buffer)
		buffer = strings.Split(buffer[0], "\n")
		args := strings.Split(buffer[0], " ")
		args = args[1:] //to remove the symbol at the very beginning
		if _, err := n.iptables.EnsureRule(iptables.Append, iptables.TableFilter, filterForwardfwruleChain, args...); err != nil {
			klog.Errorln(err, "Table: ", string(iptables.TableFilter), "chain: ", string(filterForwardfwruleChain), "args: ", args)
			return err
		}
		klog.Infoln("Rule added", "Table: ", string(iptables.TableFilter), "chain:", string(filterForwardfwruleChain), "args: ", args)

		// h.Reset()                                      // reset the hash instance
		// h.Write([]byte(fmt.Sprintf("%v", rule.Match))) // add the data to be hashed
		// ruleIndex := fmt.Sprintf("%x", h.Sum(nil))     // yield the hash of the match
		newFwHashRule.ruleHashMap[rule.Match] = rule

	}

	klog.Infoln("Map added\n", n.firewallRuleMap)
	n.iptablesdata.Reset()
	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	klog.Infoln("Deploying Rules:", "\n", n.iptablesdata.String())

	n.firewallRuleMap[key] = newFwHashRule

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
	val, ok := n.firewallRuleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnFirewallDelete Event", key)
		return nil
	}

	var chainName string
	for _, rule := range val.ruleHashMap {
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
	delete(n.firewallRuleMap, key)
	return nil
}

func (n *Iptablescontroller) OnFirewallUpdate(newfwRule *v1.FireWallRule) error {
	if fwruleValidationCheck(newfwRule) == FWRULE_INVALIDE {
		return fmt.Errorf("FW Rule is invalid")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	key := newfwRule.GetNamespace() + newfwRule.GetName()
	klog.Infof("onUpdate Called!: %s", key)
	oldRules, exist := n.firewallRuleMap[key]
	if !exist {
		n.firewallruleSynced = false
		return fmt.Errorf("Updating empty value on key(%s) detected During OnFirewallUpdate Event", key)
	}

	n.iptablesdata.Reset()
	n.iptables.SaveInto(iptables.TableFilter, n.iptablesdata)
	// klog.Infof("Original iptable contents: %s", n.iptablesdata.String())
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	// for k, v := range n.natruleMap {
	// 	klog.Infof("key: %s, value: %+v", k, v)
	// }

	var newFwHashRule FwRuleStruct
	newFwHashRule.ruleHashMap = make(map[v1.Match]v1.Rules)
	chainName := string(filterForwardfwruleChain)
	// h := sha256.New() //Bang: new Hash instance

	for _, rule := range newfwRule.Spec.Rules {
		// h.Reset()                                      // reset the hash instance
		// h.Write([]byte(fmt.Sprintf("%v", rule.Match))) // add the data to be hashed
		// ruleIndex := fmt.Sprintf("%x", h.Sum(nil))     // yield the hash of the match

		oldRule, exist := oldRules.ruleHashMap[rule.Match]
		if exist { //If match exist in the old rule
			klog.Infoln("old-rule exists")
			if oldRule.Action == rule.Action { //If even the action statement in the old and the new rule is same
				// Do nothing
				klog.Infoln("match and action is same")

			} else { //Otherwise
				//update the rule in the namespace (lines) since the new rule has to replace the action of the old action
				klog.Infoln("action is different. Need  rule modification")
				var buffer []string
				n.appendRule(&oldRule, chainName, &buffer) // buffer[0] <- iptable command for the oldrule
				n.appendRule(&rule, chainName, &buffer)    // buffer[1] <- iptable command for the updated rule
				oldCommand := strings.ReplaceAll(buffer[0], "\n", "")
				newCommand := strings.ReplaceAll(buffer[1], "\n", "")

				for i, line := range lines {
					// klog.Infoln("oldrule:", oldCommand)
					// klog.Infoln("line:", line)
					if line == oldCommand {
						lines[i] = newCommand
						// klog.Infoln("Gotcha!!")
						break
					}
				}
			}
			// Since the old map is addressed, it should never be considered again in the for loop
			// So we delete that entry in the old map entry
			delete(oldRules.ruleHashMap, rule.Match)
		} else { //If the new rule's match is cannot be found in the old rule,
			//the new rule should be added to the namespace (lines)
			klog.Infoln("Updating a new rule")
			n.appendRule(&rule, chainName, &lines)
		}
		//add the entry in the internal rule memory (newHashRuleMap)
		newFwHashRule.ruleHashMap[rule.Match] = rule
	}
	//delete the remaining old rules (within n.firewallRuleMap[key]) in the namespace (lines)
	//Since the new rule do not want them to stay in the namespace
	for _, rule := range oldRules.ruleHashMap {
		n.removeRule(&rule, chainName, &lines)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)

	klog.Infof("Deploying rules\n : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("filter", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.firewallRuleMap[key] = newFwHashRule
	klog.Infoln("Map status\n", n.firewallRuleMap)
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
