package x

import "sync"

// TMap is a helper type for the map type RWMap is using. Usually you use the
// type by definining your own type alias:
//
//	type testersMap = x.TMap[int, testing.TB]
//
// Which shortens type usage:
//
//	x.Tx(testers, func(m testersMap) {
//	    delete(m, goid())
//	})
type TMap[T comparable, U any] map[T]U

// RWMap is a type for a thread-safe Go map. It tries to be short and simple.
type RWMap[T comparable, U any] struct {
	sync.RWMutex
	m TMap[T, U]
}

// NewRWMap creates a new thread-safe map that's as simple as possible. The
// first version had only two functions Tx and Rx to allow interact with the
// map.
func NewRWMap[T comparable, U any](size ...int) *RWMap[T, U] {
	// build in make() have to deal by us
	switch len(size) {
	case 1:
		return &RWMap[T, U]{m: make(map[T]U, size[0])}
	default:
		return &RWMap[T, U]{m: make(map[T]U)}
	}
}

// Tx executes a critical section during the function given as an argument. This
// critical section allows the map be updated. If you only need to read the map
// please use the Rx function that's for a read-only critical section.
func Tx[T comparable, U any](m *RWMap[T, U], f func(m TMap[T, U])) {
	m.Lock()
	defer m.Unlock()
	f(m.m)
}

// Set sets a key value pair to the map.
func Set[T comparable, U any](m *RWMap[T, U], key T, val U) U {
	m.Lock()
	defer m.Unlock()
	m.m[key] = val
	return val
}

// Del deletes a key value pair from the map.
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
