package rulelister_test

import (
	"fmt"
	"testing"

	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
	"github.com/cho4036/virtualrouter/component/rulemanager/rule/rulelister"
)

func TestRulelister(t *testing.T) {
	r := rulelister.New()
	rule := rule.New()
	rule.Set("test")

	r.Add(rule)
	fmt.Println(r.List())

}
