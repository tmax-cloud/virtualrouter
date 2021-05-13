package rulelister

import (
	"fmt"

	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
	"github.com/cho4036/virtualrouter/util/store"
)

type RuleLister interface {
	Add(rule rule.Rule, overrideFlag bool) error
	Update(rule.Rule) error
	Delete(rule.Rule) error
	List() []rule.Rule
	Get(obj rule.Rule) rule.Rule
}
type ruleLister struct {
	list store.Storer
}

func (r *ruleLister) Add(rule rule.Rule, overrideFlag bool) error {
	err := r.list.Add(rule, overrideFlag)
	return err
}

func (r *ruleLister) Update(rule rule.Rule) error {
	err := r.list.Update(rule)
	return err
}

func (r *ruleLister) Delete(rule rule.Rule) error {
	err := r.list.Delete(rule)
	return err
}

func (r *ruleLister) List() []rule.Rule {
	var list []rule.Rule
	for _, item := range r.list.List() {
		list = append(list, item.(rule.Rule))
	}
	return list
}

func (r *ruleLister) Get(item rule.Rule) rule.Rule {
	storedItem := r.list.Get(item)
	return storedItem.(rule.Rule)
}

var keyFunc = func(obj interface{}) (string, error) {
	keyObj := obj.(rule.Rule)
	key := keyObj.GetMetadata().Name
	if key == "" {
		return key, fmt.Errorf("ruleName is empty")
	}
	return key, nil
}

func New() RuleLister {
	return &ruleLister{
		list: store.NewStore(keyFunc),
	}
}
