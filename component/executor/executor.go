package executor

type Interface interface {
	Start(stopCh <-chan struct{}) error
}
