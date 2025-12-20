package components

import (
	"slices"
	"sync"
)

// Map provides a declarative map collection, similar to C# Dictionary.
// K must be comparable; V can be any type.
type Map[K comparable, V any] struct {
	items map[K]V
}

// NewMap creates a new Map with an initial capacity.
func NewMap[K comparable, V any](capacity int) *Map[K, V] {
	return &Map[K, V]{items: make(map[K]V, capacity)}
}

// NewEmptyMap creates a new Map with default capacity.
func NewEmptyMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{items: make(map[K]V)}
}

func (m *Map[K, V]) Add(key K, value V) {
	m.items[key] = value
}

func (m *Map[K, V]) AddAll(entries map[K]V) {
	if entries == nil {
		return
	}
	for k, v := range entries {
		m.items[k] = v
	}
}

func (m *Map[K, V]) Remove(key K) {
	delete(m.items, key)
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	val, ok := m.items[key]
	return val, ok
}

func (m *Map[K, V]) GetOrDefault(key K, defaultValue V) V {
	if val, ok := m.items[key]; ok {
		return val
	}
	return defaultValue
}

func (m *Map[K, V]) ContainsKey(key K) bool {
	_, ok := m.items[key]
	return ok
}

func (m *Map[K, V]) ContainsValue(value V, equals func(a, b V) bool) bool {
	for _, v := range m.items {
		if equals(v, value) {
			return true
		}
	}
	return false
}

func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return slices.Clip(keys)
}

func (m *Map[K, V]) Values() []V {
	vals := make([]V, 0, len(m.items))
	for _, v := range m.items {
		vals = append(vals, v)
	}
	return slices.Clip(vals)
}

func (m *Map[K, V]) Clear() {
	m.items = nil
	m.items = make(map[K]V)
}

func (m *Map[K, V]) Count() int {
	return len(m.items)
}

func (m *Map[K, V]) ForEach(f func(key K, value V)) {
	for k, v := range m.items {
		f(k, v)
	}
}

func (m *Map[K, V]) ParallelForEach(f func(key K, value V)) {
	var wg sync.WaitGroup
	for k, v := range m.items {
		wg.Add(1)
		go func(kk K, vv V) {
			defer wg.Done()
			f(kk, vv)
		}(k, v)
	}
	wg.Wait()
}
