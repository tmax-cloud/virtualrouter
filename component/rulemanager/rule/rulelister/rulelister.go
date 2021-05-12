package rulelister

import (
	"github.com/cho4036/virtualrouter/component/rulemanager/rule"
	"github.com/cho4036/virtualrouter/util/store"
)

type RuleLister interface {
	Add(rule.Rule) error
	List() []rule.Rule
}

type RuleIndexer interface {
	store.Store
}

type ruleLister struct {
	indexer RuleIndexer
}

func New() RuleLister {
	return &ruleLister{}
}

func (r *ruleLister) Add(rule rule.Rule) error {
	r.indexer.Add(rule)
	return nil
}

func (r *ruleLister) Update(rule rule.Rule) error {
	return nil
}

func (r *ruleLister) Delete(rule rule.Rule) error {
	return nil
}

func (r *ruleLister) List() []rule.Rule {
	var list []rule.Rule
	for _, item := range r.indexer.List() {
		list = append(list, item.(rule.Rule))
	}
	return list
}

func (r *ruleLister) Get() rule.Rule {
	return nil
}
