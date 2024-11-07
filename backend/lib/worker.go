package lib

import (
	"runtime"
	"sync"
)

type Worker[T any] struct {
    in   chan func()
    wg   sync.WaitGroup
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

func (w *Worker[T]) Process(data []T, workerFunc func(int, T)) {
    // Enviar as tarefas
    for i, item := range data {
        i, item := i, item // Captura o índice e o valor localmente
        w.in <- func() {
            workerFunc(i, item)
        }
    }

    // Finalizar o worker pool após todas as tarefas serem enviadas
    close(w.in)
    w.wg.Wait()
}
