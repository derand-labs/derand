package proverlogic

import (
	"reflect"
	"sync"
)

type waitGroup struct {
	mu    sync.Mutex
	errCh []chan error
}

func newWaitGroup() *waitGroup {
	return &waitGroup{errCh: make([]chan error, 0)}
}

// Run is not thread-safe. Do not call this in multiple threads.
func (w *waitGroup) Run(f func() error) {
	errC := make(chan error, 1)
	w.errCh = append(w.errCh, errC)

	go func() {
		errC <- f()
	}()
}

func (w *waitGroup) Wait() error {
	errC := w.errCh
	w.errCh = nil

	cases := make([]reflect.SelectCase, len(errC))
	for i, ch := range errC {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}

	n := 0
	for n < len(errC) {
		_, v, ok := reflect.Select(cases)
		if ok {
			n++
			if err, ok := v.Interface().(error); ok {
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
