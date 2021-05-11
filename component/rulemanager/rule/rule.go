package rule

type Rule interface {
	Add()
	GetRule() Rule
}

type rule struct {
}

type RuleLister interface {
	Add(Rule)
	List() (ret []Rule, err error)
	Get(name string) (Rule, error)
}

type ruleLister struct {
	indexer Indexer
}

func (r *ruleLister) newRuleLister() {
}

type Indexer interface {
	Store

	// Index(indexName string, rule Rule) ([]Rule, error)

	// GetIndexer() Indexer
}

type Store interface {
	Add(Rule) error
	Update(Rule) error
	Delete(Rule) error
	List() []Rule
	Get()
}
