package interfaces

// IList defines a declarative interface for a list collection.
// T can be any type for the list elements.
// Provides both sequential and parallel iteration capabilities.
type IList[T any] interface {
	Add(item T)
	AddAll(items []T)
	TrimToSize()
	Remove(item T, equals func(a, b T) bool)
	Filter(f func(item T) bool) IList[T]
	Get(index int) T
	Items() []T
	Len() int
	IsEmpty() bool
	Clear()
	Range(f func(item T))
	RangeIndex(f func(index int, item T))
	RangeParallel(f func(item T))
	RangeIndexParallel(f func(index int, item T))
}

// IMap defines a declarative interface for a map collection.
// K must be comparable to be used as a map key.
// V can be any type for the map value.
type IMap[K comparable, V any] interface {
	Add(key K, value V)
	AddAll(entries map[K]V)
	Remove(key K)
	Get(key K) (V, bool)
	GetOrDefault(key K, defaultValue V) V
	ContainsKey(key K) bool
	ContainsValue(value V, equals func(a, b V) bool) bool
	Keys() []K
	Values() []V
	Clear()
	Count() int
	ForEach(f func(key K, value V))
	ParallelForEach(f func(key K, value V))
}
