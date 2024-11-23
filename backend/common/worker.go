package common

import (
	"runtime"
	"sync"

	"github.com/rs/zerolog"
)

type IWorker[T any] interface {
	Execute(workerFunc func() error, logger zerolog.Logger, errMsg string, errChan chan error)
	ForIndex(data *[]T, workerFunc func(index int, count int, data []T) error) error
	ForEach(data *[]T, workerFunc func(index int, item T) error) error
	VoidForEach(data *[]T, workerFunc func(index int, item T))
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

func (w *Worker[T]) Execute(workerFunc func() error, logger zerolog.Logger, errMsg string, errChan chan error) {
	err := workerFunc()
	if err != nil {
		logger.Error().Err(err).Msg(errMsg)
		errChan <- err
	}
}

func (w *Worker[T]) ForIndex(data *[]T, workerFunc func(index int, count int, data []T) error) error {
	list := *data
	len := len(list)
	for i := range list {
		err := workerFunc(i, len, *data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker[T]) ForEach(data *[]T, workerFunc func(index int, item T) error) error {
	list := *data
	for i, item := range list {
		err := workerFunc(i, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker[T]) VoidForEach(data *[]T, workerFunc func(index int, item T)) {
	list := *data
	for i, item := range list {
		workerFunc(i, item)
	}
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
