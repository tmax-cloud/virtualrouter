package store

import (
	"fmt"

	"github.com/tmax-cloud/virtualrouter/util/threadsafemap"
)

type Storer interface {
	Add(item interface{}, overrideFlag bool) error
	Update(interface{}) error
	Delete(interface{}) error
	List() []interface{}
	Get(obj interface{}) interface{}
}

type KeyFunc func(obj interface{}) (string, error)

type KeyError struct {
	Obj interface{}
	Err error
}

func (k KeyError) Error() string {
	return fmt.Sprintf("couldn't retrieve key for object %+v: %v", k.Obj, k.Err)
}

type ObjectError struct {
	Obj interface{}
	Err error
}

func (o ObjectError) Error() string {
	return fmt.Sprintf("error occurd handle object %+v: %s", o.Obj, o.Err)
}

type store struct {
	list threadsafemap.ThreadSafeMap

	keyFunc KeyFunc
}

func (s *store) Add(obj interface{}, overrideFlag bool) error {
	key, err := s.keyFunc(obj)
	if err != nil {
		return KeyError{Obj: obj, Err: err}
	}
	if !overrideFlag {
		if s.list.Get(key) != nil {
			return KeyError{Obj: obj, Err: fmt.Errorf("already exist")}
		}
	}

	s.list.Add(key, obj)
	return nil
}

func (s *store) Update(obj interface{}) error {
	key, err := s.keyFunc(obj)
	if err != nil {
		return KeyError{Obj: obj, Err: err}
	}
	s.list.Update(key, obj)
	return nil
}

func (s *store) Delete(obj interface{}) error {
	key, err := s.keyFunc(obj)
	if err != nil {
		return KeyError{Obj: obj, Err: err}
	}
	s.list.Delete(key)
	return nil
}

func (s *store) List() []interface{} {
	return s.list.List()
}

func (s *store) Get(obj interface{}) interface{} {
	key, err := s.keyFunc(obj)
	if err != nil {
		return KeyError{Obj: obj, Err: err}
	}
	return s.list.Get(key)
}

func NewStore(keyFunc KeyFunc) Storer {
	return &store{
		list:    threadsafemap.New(),
		keyFunc: keyFunc,
	}
}
