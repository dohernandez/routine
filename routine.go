// Package routine provides a simple way to execute a function in a goroutine and wait for it to complete.
package routine

import (
	"context"
	"sync"
)

// Operation is a function that will be executed in a goroutine.
type Operation func(ctx context.Context) error

// Routine is a struct to hold a routine execution.
type Routine struct {
	err error

	sm sync.Mutex

	isDone bool
	done   chan struct{}
}

// Err returns the error of the operation.
func (r *Routine) Err() error {
	<-r.done

	return r.err
}

// Stop stops the operation.
func (r *Routine) Stop() {
	r.sm.Lock()
	defer r.sm.Unlock()

	if r.isDone {
		return
	}

	r.isDone = true

	close(r.done)
}

// Done returns a channel that will be closed when the operation is done.
func (r *Routine) Done() <-chan struct{} {
	return r.done
}

// Run executes the given operation in a goroutine and returns a Routine object that can be used to wait for the
// operation to complete.
//
// The operation will be stopped when the given context is canceled or when the given stopper channel is closed.
func Run(ctx context.Context, op Operation, stopper <-chan struct{}) *Routine {
	ctx, cancel := context.WithCancel(ctx)

	r := &Routine{
		done: make(chan struct{}),
	}

	go func() {
		defer func() {
			cancel()

			r.Stop()
		}()

		err := op(ctx)

		r.err = err
	}()

	go func() {
		<-stopper

		cancel()
	}()

	return r
}
