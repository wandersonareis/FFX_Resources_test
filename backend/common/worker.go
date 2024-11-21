package common

import (
	"runtime"
	"sync"

	"github.com/rs/zerolog"
)

type Worker[T any] struct {
	in chan func()
	wg sync.WaitGroup
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
    len := len(*data)
	for i := range *data {
        err := workerFunc(i, len, *data)
        if err != nil {
            return err
        }
    }

    return nil
}

func (w *Worker[T]) ForEach(data []T, workerFunc func(index int, item T) error) error{
	for i, item := range data {
        err := workerFunc(i, item)
        if err != nil {
            return err
        }
    }

    return nil
}

func (w *Worker[T]) ParallelForEach(data *[]T, workerFunc func(index int, item T)) {
	for i, item := range *data {
		i, item := i, item
		w.in <- func() {
			workerFunc(i, item)
		}
	}

	close(w.in)
	w.wg.Wait()
}

func (w *Worker[T]) Close() {
	close(w.in)
	w.wg.Wait()
}