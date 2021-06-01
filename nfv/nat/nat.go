package nat

type Interface interface {
	SNAT(match *interface{}, action *interface{}, opt *interface{}) error
	DNAT(match *interface{}, action *interface{}, opt *interface{}) error
}
