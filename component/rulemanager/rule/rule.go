package rule

type Rule interface {
	GetMetadata() RuleMeta
	GetItems() interface{}
}

// type rule struct {
// 	Metadata RuleMata
// 	Item     interface{}
// }

// func New() Rule {
// 	return &rule{}
// }

// func (r rule) GetMetadata() RuleMata {

// 	return r.Metadata
// }

type RuleMeta struct {
	Name string
}
