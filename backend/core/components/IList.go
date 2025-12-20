package components

// IList defines a declarative interface for a list collection.
// T can be any type for the list elements.
// Provides both sequential and parallel iteration capabilities.
type IList[T any] interface {
	Add(item T)
	AddAll(items []T)
	TrimToSize()
	Remove(item T, equals func(a, b T) bool)
	Filter(f func(item T) bool) *List[T]
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
