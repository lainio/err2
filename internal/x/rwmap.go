package x

import "sync"

type TMap[T comparable, U any] map[T]U

type RWMap[T comparable, U any] struct {
	sync.RWMutex
	m TMap[T, U]
}

func NewRWMap[T comparable, U any](size ...int) *RWMap[T, U] {
	// build in make() have to deal by us
	switch len(size) {
	case 1:
		return &RWMap[T, U]{m: make(map[T]U, size[0])}
	default:
		return &RWMap[T, U]{m: make(map[T]U)}
	}
}

func Tx[T comparable, U any](m *RWMap[T, U], f func(m TMap[T, U])) {
	m.Lock()
	defer m.Unlock()
	f(m.m)
}

func Set[T comparable, U any](m *RWMap[T, U], key T, val U) U {
	m.Lock()
	defer m.Unlock()
	m.m[key] = val
	return val
}

func Del[T comparable, U any](m *RWMap[T, U], key T) U {
	m.Lock()
	defer m.Unlock()
	val, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	return val
}

func Rx[T comparable, U any](m *RWMap[T, U], f func(m TMap[T, U])) {
	m.RLock()
	defer m.RUnlock()
	f(m.m)
}

func Get[T comparable, U any](m *RWMap[T, U], key T) U {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}
