package parser

type Interface interface {
}

type parser struct {
}

func New() Interface {
	return &parser{}
}
