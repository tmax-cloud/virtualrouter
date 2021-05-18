package natRuleParser

import (
	"fmt"
	"reflect"

	"github.com/cho4036/virtualrouter/rule/natrule"
	"github.com/cho4036/virtualrouter/ruleParser/natRuleParser/k8sNatRule"
)

type NatRuleParser interface {
	Parse(item interface{}) (key string, rules []natrule.NatRule, err error)
	// RegisterParser(inputType string, parseFunction func(item interface{}) *natrule.NatRule)
}

type natRuleParser struct {
	parsingFunc map[string]func(item interface{}) (key string, rules []natrule.NatRule)
}

func New() NatRuleParser {
	n := &natRuleParser{
		parsingFunc: make(map[string]func(item interface{}) (key string, rules []natrule.NatRule)),
	}
	n.registerParser(k8sNatRule.InputType.Name(), k8sNatRule.ParseFunc)
	return n
}

func (n *natRuleParser) registerParser(inputType string, parseFunction func(item interface{}) (key string, rules []natrule.NatRule)) {
	n.parsingFunc[inputType] = parseFunction
}

func (n *natRuleParser) Parse(item interface{}) (key string, natRules []natrule.NatRule, _ error) {
	parseFunc, exist := n.parsingFunc[reflect.TypeOf(item).Name()]
	if !exist {
		return "", nil, fmt.Errorf("%s item parser is not registered but used", reflect.TypeOf(item).Name())
	}
	key, natRules = parseFunc(item)
	if key == "" {
		return "", nil, fmt.Errorf("failed to retrieve key of object %+v", item)
	}
	return key, natRules, nil
}
