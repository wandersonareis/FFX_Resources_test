package components

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
