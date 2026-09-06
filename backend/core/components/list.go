package components

import (
	"slices"
	"sync"
)

type List[T any] struct {
	// content is a slice of items of type T.
	content []T
}

func NewList[T any](length int) *List[T] {
	return &List[T]{content: make([]T, 0, length)}
}

func NewEmptyList[T any]() *List[T] {
	return &List[T]{content: make([]T, 0)}
}

func (l *List[T]) Items() []T {
	return l.content
}

func (l *List[T]) Add(item T) {
	l.content = append(l.content, item)
}

func (l *List[T]) AddAll(items []T) {
	if items == nil {
		return
	}
	l.content = append(l.content, items...)
}

func (l *List[T]) TrimToSize() {
	l.content = slices.Clip(l.content)
}

func (l *List[T]) Remove(item T, equals func(a, b T) bool) {
	for i, v := range l.content {
		if equals(v, item) {
			l.content = append(l.content[:i], l.content[i+1:]...)
			break
		}
	}
}

func (l *List[T]) Filter(f func(item T) bool) IList[T] {
	result := NewEmptyList[T]()

	for _, v := range l.content {
		if f(v) {
			result.Add(v)
		}
	}

	result.content = slices.Clip(result.content)

	return result
}

func (l *List[T]) Get(index int) T {
	var zero T
	if l.IsEmpty() || index < 0 || index >= len(l.content) {
		return zero
	}
	return l.content[index]
}

func (l *List[T]) Len() int {
	return len(l.content)
}

func (l *List[T]) IsEmpty() bool {
	return l.content != nil && len(l.content) == 0
}

func (l *List[T]) Clear() {
	l.content = nil
	l.content = make([]T, 0)
}

func (l *List[T]) RangeIndex(f func(index int, item T)) {
	for i, v := range l.content {
		func(i int, it T) {
			f(i, it)
		}(i, v)
	}
}

func (l *List[T]) Range(f func(item T)) {
	for _, v := range l.content {
		f(v)
	}
}

func (l *List[T]) RangeParallel(f func(item T)) {
	var wg sync.WaitGroup

	for _, v := range l.content {
		wg.Add(1)
		go func(it T) {
			defer wg.Done()
			f(it)
		}(v)
	}

	wg.Wait()
}

func (l *List[T]) RangeIndexParallel(f func(index int, item T)) {
	var wg sync.WaitGroup

	for i, v := range l.content {
		wg.Add(1)
		go func(i int, it T) {
			defer wg.Done()
			f(i, it)
		}(i, v)
	}

	wg.Wait()
}
