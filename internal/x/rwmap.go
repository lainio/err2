package x

import "sync"

// RWMap is a type for a thread-safe Go map. It tries to be short and simple.
// Tip: It's useful to create a type alias (it allows it):
//
//	testersMap = map[int]testing.TB
//
// Which shortens and makes easier to read its usage:
//
//	x.Tx(testers, func(m testersMap) {
//	    delete(m, goid())
//	})
type RWMap[M ~map[T]U, T comparable, U any] struct {
	sync.RWMutex
	m M
}

// NewRWMap creates a new thread-safe map that's as simple as possible. The
// first version had only two functions Tx and Rx to allow interact with the
// map.
func NewRWMap[M ~map[T]U, T comparable, U any](size ...int) *RWMap[M, T, U] {
	// build in make() have to deal by us
	switch len(size) {
	case 1:
		return &RWMap[M, T, U]{m: make(map[T]U, size[0])}
	default:
		return &RWMap[M, T, U]{m: make(map[T]U)}
	}
}

// Tx executes a critical section during the function given as an argument. This
// critical section allows the map be updated. If you only need to read the map
// please use the Rx function that's for a read-only critical section.
func (m *RWMap[M, T, U]) Tx(f func(m M)) {
	m.Lock()
	defer m.Unlock()
	f(m.m)
}

// Set sets a key value pair to the map with Go's normal map semantics.
func (m *RWMap[M, T, U]) Set(key T, val U) U {
	m.Lock()
	defer m.Unlock()
	m.m[key] = val
	return val
}

// Del deletes a key value pair from the map. It checks that key exists. It
// doesn't panic if key doesn't exist. The return value is valid only if key
// exists other is Go's default init value for the type.
func (m *RWMap[M, T, U]) Del(key T) U {
	m.Lock()
	defer m.Unlock()
	val, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	return val
}

// Rx executes a read-only critical section during the function given as an
// argument. This critical section allows the map be read only. If you only need
// to write the map please use the Tx function that's for a read-write critical
// section.
func (m *RWMap[M, T, U]) Rx(f func(m M)) {
	m.RLock()
	defer m.RUnlock()
	f(m.m)
}

// Get returns value for the key. If key doesn't exist it panics as normal map
// access without ok idiom does.
func (m *RWMap[M, T, U]) Get(key T) U {
	m.RLock()
	defer m.RUnlock()
	return m.m[key]
}
