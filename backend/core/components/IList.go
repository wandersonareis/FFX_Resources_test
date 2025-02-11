package components

type IList[T any] interface {
	Add(item T)
	Clip()
	Remove(item T, equals func(a, b T) bool)
	Filter(f func(item T) bool) *List[T]
	GetItems() []T
	GetLength() int
	IsEmpty() bool
	Clear()
	ForEach(f func(item T))
	ForIndex(f func(index int, item T))
	ParallelForEach(f func(item T))
	ParallelForIndex(f func(index int, item T))
}
