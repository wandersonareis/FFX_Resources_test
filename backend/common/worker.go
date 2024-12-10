package common

import (
	"runtime"
	"sync"
)

type IWorker[T any] interface {
	ParallelForEach(data *[]T, workerFunc func(index int, item T))
	Close()
}

type Worker[T any] struct {
	in        chan func()
	wg        sync.WaitGroup
	closeOnce sync.Once
}

func NewWorker[T any]() *Worker[T] {
	numCPU := runtime.NumCPU()
	w := Worker[T]{
		in: make(chan func()),
	}

	w.wg.Add(numCPU - 1)
	for i := 0; i < numCPU-1; i++ {
		go func() {
			defer w.wg.Done()
			for task := range w.in {
				task()
			}
		}()
	}

	return &w
}

func (w *Worker[T]) ParallelForEach(data *[]T, workerFunc func(index int, item T)) {
	list := *data
	for i, item := range list {
		w.in <- func() {
			workerFunc(i, item)
		}
	}

	w.Close()
	w.wg.Wait()
}

func (w *Worker[T]) Close() {
	w.closeOnce.Do(func() {
		close(w.in)
	})
	w.wg.Wait()
}
