package main

import "github.com/cho4036/virtualrouter/rule-manager/ruleManagerFactory"

func main() {
	f := ruleManagerFactory.New()
	f.NatRuleManager()
}
