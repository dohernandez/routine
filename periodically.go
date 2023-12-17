package routine

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/dohernandez/errors"
	"github.com/dohernandez/retry"
)

// PeriodicallyOption is a function that configures options for the function RunPeriodically.
type PeriodicallyOption func(*periodicallyOptions)

// periodicallyOptions is a struct to hold options for the function RunPeriodically.
type periodicallyOptions struct {
	onError       retry.OnError
	notifyOnError retry.NotifyOnError
}

// WithOnError configures the OnError option for the function RunPeriodically.
func WithOnError(onError retry.OnError) PeriodicallyOption {
	return func(o *periodicallyOptions) {
		o.onError = onError
	}
}

// WithNotifyOnError configures the NotifyOnError option for the function RunPeriodically.
func WithNotifyOnError(notifyOnError retry.NotifyOnError) PeriodicallyOption {
	return func(o *periodicallyOptions) {
		o.notifyOnError = notifyOnError
	}
}

// RunPeriodically executes the given operation in a goroutine and returns a Routine object that can be used to wait for the
// operation to complete.
//
// The operation will be executed periodically with the given frequency.
// The operation will be stopped when the given context is canceled or when the given stopper channel is closed.
func RunPeriodically(ctx context.Context, op Operation, freq time.Duration, stopper <-chan struct{}, opts ...PeriodicallyOption) *Routine {
	rp := &Routine{
		done: make(chan struct{}),
	}

	options := periodicallyOptions{}

	for _, opt := range opts {
		opt(&options)
	}

	if options.onError == nil {
		options.onError = func(err error) bool {
			return errors.Is(err, retry.Err)
		}
	}

	go func() {
		defer func() {
			rp.Stop()
		}()

		rp.err = retry.UntilFail(
			ctx,
			func(ctx context.Context) error {
				r := Run(ctx, op, stopper)
				if r.Err() != nil {
					return errors.WrapError(r.Err(), retry.Err)
				}

				return retry.Err
			},
			backoff.NewConstantBackOff(freq),
			stopper,
			retry.WithNotifyOnError(options.notifyOnError),
			retry.WithOnError(options.onError),
		)
	}()

	return rp
}
