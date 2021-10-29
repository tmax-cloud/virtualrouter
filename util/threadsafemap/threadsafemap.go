package threadsafemap

import "sync"

type ThreadSafeMap interface {
	Add(key string, obj interface{})
	Update(key string, obj interface{})
	Delete(key string)
	Get(key string) interface{}
	List() []interface{}
}

type threadSafeMap struct {
	lock  sync.RWMutex
	items map[string]interface{}
}

func (t *threadSafeMap) Add(key string, obj interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.items[key] = obj
}

func (t *threadSafeMap) Update(key string, obj interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.items[key] = obj
}

func (t *threadSafeMap) Delete(key string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	delete(t.items, key)
}

func (t *threadSafeMap) Get(key string) interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	item := t.items[key]
	return item
}

func (t *threadSafeMap) List() []interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	list := make([]interface{}, 0, len(t.items))
	for _, item := range t.items {
		list = append(list, item)
	}

	return list
}

func New() ThreadSafeMap {
	return &threadSafeMap{
		items: make(map[string]interface{}),
	}
}
