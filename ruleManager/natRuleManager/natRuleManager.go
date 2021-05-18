package natRuleManager

import (
	"fmt"

	"github.com/cho4036/virtualrouter/rule/natrule"
	"github.com/cho4036/virtualrouter/ruleParser/natRuleParser"
)

type NatRuleManager interface {
	Add(obj interface{}) error
	Delete(obj interface{}) error
	DeleteByName(name string) error
	Update(obj interface{}) error
	List() []natrule.NatRule
	GetByName(name string) []natrule.NatRule
}

type natRuleManager struct {
	NatRuleList   map[string][]natrule.NatRule
	NatRuleParser natRuleParser.NatRuleParser
}

func New() NatRuleManager {
	return &natRuleManager{
		NatRuleList:   make(map[string][]natrule.NatRule),
		NatRuleParser: natRuleParser.New(),
	}
}

func (n *natRuleManager) Add(item interface{}) error {
	key, natRules, err := n.NatRuleParser.Parse(item)
	if err != nil {
		return err
	}
	if _, exist := n.NatRuleList[key]; exist {
		return fmt.Errorf("duplicated key %s. the key is already exist", key)
	}
	n.NatRuleList[key] = append(n.NatRuleList[key], natRules...)
	return nil
}

func (n *natRuleManager) Delete(item interface{}) error {
	key, _, err := n.NatRuleParser.Parse(item)
	if err != nil {
		return err
	}
	delete(n.NatRuleList, key)
	return nil
}

func (n *natRuleManager) DeleteByName(key string) error {
	if _, exist := n.NatRuleList[key]; !exist {
		return fmt.Errorf("there are no rules matching with the key:%s", key)
	}
	delete(n.NatRuleList, key)
	return nil
}

func (n *natRuleManager) Update(item interface{}) error {
	key, natRules, err := n.NatRuleParser.Parse(item)
	if err != nil {
		return err
	}
	n.NatRuleList[key] = natRules
	return nil
}

func (n *natRuleManager) List() []natrule.NatRule {
	var natRules []natrule.NatRule

	for _, val := range n.NatRuleList {
		natRules = append(natRules, val...)
	}
	return natRules
}

func (n *natRuleManager) GetByName(name string) []natrule.NatRule {
	return n.NatRuleList[name]
}

// func (n *natRuleManager) Parser() natRuleParser.NatRuleParser {
// 	return n.NatRuleParser
// }
