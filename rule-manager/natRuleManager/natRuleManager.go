package natRuleManager

type NatRuleManager interface {
}

type natRuleManager struct {
}

func New() NatRuleManager {
	return &natRuleManager{}
}
