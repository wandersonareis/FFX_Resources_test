package components

import (
	"slices"
	"sync"
)

type List[T any] struct {
	// Items is a slice of items of type T.
	Items []T
}

func NewList[T any](length int) *List[T] {
	return &List[T]{Items: make([]T, 0, length)}
}

func NewEmptyList[T any]() *List[T] {
	return &List[T]{Items: make([]T, 0)}
}

func (l *List[T]) GetItems() []T {
	return l.Items
}

func (l *List[T]) Add(item T) {
	l.Items = append(l.Items, item)
}

func (l *List[T]) Clip() {
	l.Items = slices.Clip(l.Items)
}

func (l *List[T]) Remove(item T, equals func(a, b T) bool) {
	for i, v := range l.Items {
		if equals(v, item) {
			l.Items = append(l.Items[:i], l.Items[i+1:]...)
			break
		}
	}
}

func (l *List[T]) Filter(f func(item T) bool) *List[T] {
	result := NewEmptyList[T]()

	for _, v := range l.Items {
		if f(v) {
			result.Add(v)
		}
	}

	result.Items = slices.Clip(result.Items)

	return result
}

func (l *List[T]) GetLength() int {
	return len(l.Items)
}

func (l *List[T]) IsEmpty() bool {
	return len(l.Items) == 0
}

func (l *List[T]) Clear() {
	l.Items = nil
}

func (l *List[T]) ForIndex(f func(index int, item T)) {
	for i, v := range l.Items {
		func(i int, it T) {
			f(i, it)
		}(i, v)
	}
}

func (l *List[T]) ForEach(f func(item T)) {
	for _, v := range l.Items {
		f(v)
	}
}

func (l *List[T]) ParallelForEach(f func(item T)) {
	var wg sync.WaitGroup

	for _, v := range l.Items {
		wg.Add(1)
		go func(it T) {
			defer wg.Done()
			f(it)
		}(v)
	}

	wg.Wait()
}

func (l *List[T]) ParallelForIndex(f func(index int, item T)) {
	var wg sync.WaitGroup

	for i, v := range l.Items {
		wg.Add(1)
		go func(i int, it T) {
			defer wg.Done()
			f(i, it)
		}(i, v)
	}

	wg.Wait()
}
